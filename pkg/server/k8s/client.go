package k8s

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/deploy"
	"github.com/luizalabs/teresa/pkg/server/exec"
	"github.com/luizalabs/teresa/pkg/server/spec"
	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	k8sv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/apps/v1beta1"
	asv1 "k8s.io/client-go/pkg/apis/autoscaling/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	patchDeployEnvVarsTmpl            = `{"metadata": {"annotations": {"kubernetes.io/change-cause": "update env vars"}}, "spec":{"template":{"metadata": {"annotations": {"date": "%s"}}, "spec":{"containers":%s}}}}`
	patchCronJobEnvVarsTmpl           = `{"metadata": {"annotations": {"kubernetes.io/change-cause": "update env vars"}}, "spec":{"template":{"metadata":{"annotations":{"date": "%s"}}}, "jobTemplate":{"spec": {"template": {"spec": {"containers":%s}}}}}}`
	patchDeployRollbackToRevisionTmpl = `{"spec":{"rollbackTo":{"revision": %s}}}`
	patchDeployReplicasTmpl           = `{"spec":{"replicas": %d}}`
	patchServiceAnnotationsTmpl       = `{"metadata":{"annotations": %s}}`
	revisionAnnotation                = "deployment.kubernetes.io/revision"
)

type Client struct {
	conf          *restclient.Config
	podRunTimeout time.Duration
	ingress       bool
}

func (k *Client) buildClient() (*kubernetes.Clientset, error) {
	c, err := kubernetes.NewForConfig(k.conf)
	if err != nil {
		return nil, errors.Wrap(err, "create k8s client failed")
	}
	return c, nil
}

func (k *Client) HealthCheck() error {
	kc, err := k.buildClient()
	if err != nil {
		return err
	}
	_, err = kc.CoreV1().Namespaces().List(metav1.ListOptions{})
	return err
}

func (k *Client) getNamespace(namespace string) (*k8sv1.Namespace, error) {
	kc, err := k.buildClient()
	if err != nil {
		return nil, err
	}
	ns, err := kc.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return ns, nil
}

func (k *Client) DeployAnnotation(namespace, deployName, annotation string) (string, error) {
	kc, err := k.buildClient()
	if err != nil {
		return "", err
	}

	d, err := kc.AppsV1beta1().
		Deployments(namespace).
		Get(deployName, metav1.GetOptions{})

	if err != nil {
		return "", errors.Wrap(err, "get deploy annotation failed")
	}

	return d.Annotations[annotation], nil
}

func (k *Client) NamespaceAnnotation(namespace, annotation string) (string, error) {
	ns, err := k.getNamespace(namespace)
	if err != nil {
		return "", errors.Wrap(err, "get annotation failed")
	}

	return ns.Annotations[annotation], nil
}

func (k *Client) NamespaceLabel(namespace, label string) (string, error) {
	ns, err := k.getNamespace(namespace)
	if err != nil {
		return "", errors.Wrap(err, "get label failed")
	}

	return ns.Labels[label], nil
}

func (k *Client) PodList(namespace string, opts *app.PodListOptions) ([]*app.Pod, error) {
	kc, err := k.buildClient()
	if err != nil {
		return nil, err
	}
	k8sOpts := appPodListOptsToK8s(opts)
	podList, err := kc.CoreV1().Pods(namespace).List(*k8sOpts)
	if err != nil {
		return nil, err
	}

	pods := make([]*app.Pod, 0)
	for _, pod := range podList.Items {
		p := &app.Pod{Name: pod.Name}

		if pod.Status.StartTime != nil {
			p.Age = int64(time.Since(pod.Status.StartTime.Time))
		}

		for _, status := range pod.Status.ContainerStatuses {
			if status.State.Waiting != nil {
				p.State = status.State.Waiting.Reason
			} else if status.State.Terminated != nil {
				p.State = status.State.Terminated.Reason
			} else if status.State.Running != nil {
				p.State = string(api.PodRunning)
			}
			p.Restarts = status.RestartCount
			p.Ready = status.Ready
			if p.State != "" {
				break
			}
		}
		pods = append(pods, p)
	}
	return pods, nil
}

func (k *Client) PodLogs(namespace string, podName string, opts *app.LogOptions) (io.ReadCloser, error) {
	kc, err := k.buildClient()
	if err != nil {
		return nil, err
	}
	req := kc.CoreV1().Pods(namespace).GetLogs(
		podName,
		&k8sv1.PodLogOptions{
			Follow:    opts.Follow,
			TailLines: &opts.Lines,
			Previous:  opts.Previous,
			Container: opts.Container,
		},
	)

	return req.Stream()
}

func newNs(a *app.App, user string) *k8sv1.Namespace {
	return &k8sv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: a.Name,
			Labels: map[string]string{
				app.TeresaTeamLabel: a.Team,
			},
			Annotations: map[string]string{
				app.TeresaLastUser: user,
			},
		},
	}
}

func addAppToNs(a *app.App, ns *k8sv1.Namespace) error {
	b, err := json.Marshal(a)
	if err != nil {
		return err
	}

	ns.Annotations[app.TeresaAnnotation] = string(b)
	return nil
}

func addLimitRangeQuantityToResourceList(r *k8sv1.ResourceList, lrQuantity []*app.LimitRangeQuantity) error {
	if lrQuantity == nil {
		return nil
	}

	rl := k8sv1.ResourceList{}
	for _, item := range lrQuantity {
		name := k8sv1.ResourceName(item.Resource)
		q, err := resource.ParseQuantity(item.Quantity)
		if err != nil {
			return err
		}
		rl[name] = q
	}
	*r = rl
	return nil
}

func parseLimitRangeParams(lrItem *k8sv1.LimitRangeItem, lim *app.Limits) error {
	if err := addLimitRangeQuantityToResourceList(&lrItem.Default, lim.Default); err != nil {
		return err
	}
	return addLimitRangeQuantityToResourceList(&lrItem.DefaultRequest, lim.DefaultRequest)
}

func newLimitRange(a *app.App) (*k8sv1.LimitRange, error) {
	lrItem := k8sv1.LimitRangeItem{Type: k8sv1.LimitTypeContainer}
	if err := parseLimitRangeParams(&lrItem, a.Limits); err != nil {
		return nil, err
	}

	lr := &k8sv1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Name: "limits",
		},
		Spec: k8sv1.LimitRangeSpec{
			Limits: []k8sv1.LimitRangeItem{lrItem},
		},
	}
	return lr, nil
}

func newHPA(a *app.App) *asv1.HorizontalPodAutoscaler {
	tcpu := a.Autoscale.CPUTargetUtilization
	minr := a.Autoscale.Min

	return &asv1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      a.Name,
			Namespace: a.Name,
		},
		Spec: asv1.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: asv1.CrossVersionObjectReference{
				APIVersion: "extensions/v1beta1",
				Kind:       "Deployment",
				Name:       a.Name,
			},
			TargetCPUUtilizationPercentage: &tcpu,
			MaxReplicas:                    a.Autoscale.Max,
			MinReplicas:                    &minr,
		},
	}
}

func (k *Client) CreateNamespace(a *app.App, user string) error {
	kc, err := k.buildClient()
	if err != nil {
		return err
	}

	ns := newNs(a, user)
	if err := addAppToNs(a, ns); err != nil {
		return err
	}

	_, err = kc.CoreV1().Namespaces().Create(ns)
	return err
}

func (k *Client) CreateQuota(a *app.App) error {
	kc, err := k.buildClient()
	if err != nil {
		return err
	}

	lr, err := newLimitRange(a)
	if err != nil {
		return err
	}

	_, err = kc.CoreV1().LimitRanges(a.Name).Create(lr)
	return err
}

func (c *Client) GetSecret(namespace, secretName string) (map[string][]byte, error) {
	kc, err := c.buildClient()
	if err != nil {
		return nil, err
	}
	s, err := kc.CoreV1().Secrets(namespace).Get(secretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return s.Data, nil
}

func (c *Client) CreateOrUpdateSecret(namespace, secretName string, data map[string][]byte) error {
	kc, err := c.buildClient()
	if err != nil {
		return err
	}

	s := &k8sv1.Secret{
		Type: k8sv1.SecretTypeOpaque,
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Data: data,
	}

	_, err = kc.CoreV1().Secrets(namespace).Update(s)
	if c.IsNotFound(err) {
		_, err = kc.CoreV1().Secrets(namespace).Create(s)
	}
	return err
}

func (k *Client) CreateOrUpdateAutoscale(a *app.App) error {
	kc, err := k.buildClient()
	if err != nil {
		return err
	}

	hpa := newHPA(a)

	_, err = kc.AutoscalingV1().HorizontalPodAutoscalers(a.Name).Update(hpa)
	if k.IsNotFound(err) {
		_, err = kc.AutoscalingV1().HorizontalPodAutoscalers(a.Name).Create(hpa)
	}
	return err
}

func (k *Client) AddressList(namespace string) ([]*app.Address, error) {
	kc, err := k.buildClient()
	if err != nil {
		return nil, err
	}

	srvs, err := kc.CoreV1().Services(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "get addr list failed")
	}

	addrs := []*app.Address{}
	for _, srv := range srvs.Items {
		for _, i := range srv.Status.LoadBalancer.Ingress {
			h := i.Hostname
			if h == "" {
				h = i.IP
			}
			addrs = append(addrs, &app.Address{Hostname: h})
		}
	}
	return addrs, nil
}

func (k *Client) Status(namespace string) (*app.Status, error) {
	kc, err := k.buildClient()
	if err != nil {
		return nil, err
	}

	pods, err := k.PodList(namespace, &app.PodListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "get status failed")
	}

	hpa, err := kc.AutoscalingV1().
		HorizontalPodAutoscalers(namespace).
		Get(namespace, metav1.GetOptions{})

	if err != nil {
		if !k.IsNotFound(err) {
			return nil, errors.Wrap(err, "get status failed")
		}
	}

	var cpu int32 = -1
	if hpa != nil && hpa.Status.CurrentCPUUtilizationPercentage != nil {
		cpu = *hpa.Status.CurrentCPUUtilizationPercentage
	}

	stat := &app.Status{
		CPU:  cpu,
		Pods: pods,
	}
	return stat, nil
}

func (k *Client) Autoscale(namespace string) (*app.Autoscale, error) {
	kc, err := k.buildClient()
	if err != nil {
		return nil, err
	}

	hpa, err := kc.AutoscalingV1().
		HorizontalPodAutoscalers(namespace).
		Get(namespace, metav1.GetOptions{})

	if err != nil {
		if k.IsNotFound(err) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "get autoscale failed")
	}

	var cpu, min int32
	if hpa.Spec.TargetCPUUtilizationPercentage != nil {
		cpu = *hpa.Spec.TargetCPUUtilizationPercentage
	}
	if hpa.Spec.MinReplicas != nil {
		min = *hpa.Spec.MinReplicas
	}

	as := &app.Autoscale{
		CPUTargetUtilization: cpu,
		Min:                  min,
		Max:                  hpa.Spec.MaxReplicas,
	}
	return as, nil
}

func (k *Client) Limits(namespace, name string) (*app.Limits, error) {
	kc, err := k.buildClient()
	if err != nil {
		return nil, err
	}

	lr, err := kc.CoreV1().LimitRanges(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "get limits failed")
	}

	var def, defReq []*app.LimitRangeQuantity
	for _, item := range lr.Spec.Limits {
		for k, v := range item.Default {
			q := &app.LimitRangeQuantity{
				Resource: string(k),
				Quantity: v.String(),
			}
			def = append(def, q)
		}
		for k, v := range item.DefaultRequest {
			q := &app.LimitRangeQuantity{
				Resource: string(k),
				Quantity: v.String(),
			}
			defReq = append(defReq, q)
		}
	}

	lim := &app.Limits{
		Default:        def,
		DefaultRequest: defReq,
	}
	return lim, nil
}

func (k *Client) CreateOrUpdateConfigMap(namespace, name string, data map[string]string) error {
	kc, err := k.buildClient()
	if err != nil {
		return err
	}

	cm := configMapSpec(namespace, name, data)
	_, err = kc.CoreV1().ConfigMaps(namespace).Update(cm)
	if k.IsNotFound(err) {
		_, err = kc.CoreV1().ConfigMaps(namespace).Create(cm)
	}
	return err
}

func (k *Client) DeleteConfigMap(namespace, name string) error {
	kc, err := k.buildClient()
	if err != nil {
		return err
	}
	err = kc.CoreV1().ConfigMaps(namespace).Delete(name, &metav1.DeleteOptions{})
	return errors.Wrap(err, "delete configMap failed")
}

func (k *Client) CreateOrUpdateDeploy(deploySpec *spec.Deploy) error {
	kc, err := k.buildClient()
	if err != nil {
		return err
	}

	replicas := k.currentPodReplicasFromDeploy(deploySpec.Namespace, deploySpec.Name)
	deployYaml, err := deploySpecToK8sDeploy(deploySpec, replicas)
	if err != nil {
		return err
	}

	_, err = kc.AppsV1beta1().Deployments(deploySpec.Namespace).Update(deployYaml)
	if k.IsNotFound(err) {
		_, err = kc.AppsV1beta1().Deployments(deploySpec.Namespace).Create(deployYaml)
	}
	return err
}

func (c *Client) CreateOrUpdateCronJob(cronJobSpec *spec.CronJob) error {
	kc, err := c.buildClient()
	if err != nil {
		return err
	}

	cronJobYaml, err := cronJobSpecToK8sCronJob(cronJobSpec)
	if err != nil {
		return err
	}

	_, err = kc.CronJobs(cronJobSpec.Namespace).Update(cronJobYaml)
	if c.IsNotFound(err) {
		_, err = kc.CronJobs(cronJobSpec.Namespace).Create(cronJobYaml)
	}
	return err
}

func (k *Client) PodRun(podSpec *spec.Pod) (io.ReadCloser, <-chan int, error) {
	kc, err := k.buildClient()
	if err != nil {
		return nil, nil, err
	}

	podYaml, err := podSpecToK8sPod(podSpec)
	if err != nil {
		return nil, nil, errors.Wrap(err, "define build pod spec failed")
	}
	pod, err := kc.Pods(podSpec.Namespace).Create(podYaml)
	if err != nil {
		return nil, nil, errors.Wrap(err, "pod create failed")
	}

	exitCodeChan := make(chan int)
	r, w := io.Pipe()
	go func() {
		defer func() {
			w.Close()
			close(exitCodeChan)
		}()

		if err := k.waitPodStart(pod, 1*time.Second, 5*time.Minute); err != nil {
			if err == wait.ErrWaitTimeout {
				exitCodeChan <- exec.ExitCodeTimeout
			} else {
				exitCodeChan <- exec.ExitCodeError
			}
			return
		}

		opts := &app.LogOptions{Lines: 10, Follow: true}
		stream, err := k.PodLogs(podSpec.Namespace, podSpec.Name, opts)
		if err != nil {
			exitCodeChan <- exec.ExitCodeError
			return
		}
		io.Copy(w, stream)

		if err = k.waitPodEnd(pod, 3*time.Second, k.podRunTimeout); err != nil {
			if err == wait.ErrWaitTimeout {
				exitCodeChan <- exec.ExitCodeTimeout
			} else {
				exitCodeChan <- exec.ExitCodeError
			}
			return
		}

		exitCode, err := k.podExitCode(pod)
		if err != nil {
			exitCodeChan <- exec.ExitCodeError
			return
		}
		exitCodeChan <- exitCode
		go k.DeletePod(pod.Namespace, pod.Name)
	}()
	return r, exitCodeChan, nil
}

func (k *Client) hasService(namespace, appName string) (bool, error) {
	kc, err := k.buildClient()
	if err != nil {
		return false, err
	}
	_, err = kc.CoreV1().Services(namespace).Get(appName, metav1.GetOptions{})
	if err != nil {
		if k.IsNotFound(err) {
			return false, nil
		}
		return false, errors.Wrap(err, "get service failed")
	}
	return true, nil
}

func (k *Client) CreateService(svcSpec *spec.Service) error {
	kc, err := k.buildClient()
	if err != nil {
		return err
	}
	ss := serviceSpecToK8s(svcSpec)
	_, err = kc.CoreV1().Services(svcSpec.Namespace).Create(ss)
	return errors.Wrap(err, "create service failed")
}

func (k *Client) HasIngress(namespace, appName string) (bool, error) {
	kc, err := k.buildClient()
	if err != nil {
		return false, err
	}
	_, err = kc.ExtensionsV1beta1().
		Ingresses(namespace).
		Get(appName, metav1.GetOptions{})

	if err != nil {
		if k.IsNotFound(err) {
			return false, nil
		}
		return false, errors.Wrap(err, "get ingress failed")
	}
	return true, nil
}

func (k *Client) createIngress(namespace, appName, vHost string) error {
	kc, err := k.buildClient()
	if err != nil {
		return err
	}
	igsSpec := ingressSpec(namespace, appName, vHost)
	_, err = kc.ExtensionsV1beta1().Ingresses(namespace).Create(igsSpec)
	return errors.Wrap(err, "create ingress failed")
}

// ExposeDeploy creates a service and/or a ingress if needed
func (k *Client) ExposeDeploy(namespace, appName, vHost, svcType, portName string, w io.Writer) error {
	hasSrv, err := k.hasService(namespace, appName)
	if err != nil {
		return err
	}
	if !hasSrv {
		fmt.Fprintln(w, "Exposing service")
		svcSpec := spec.NewDefaultService(appName, svcType, portName)
		if err := k.CreateService(svcSpec); err != nil {
			return err
		}
	}

	if !k.ingress {
		return nil
	}
	hasIgs, err := k.HasIngress(namespace, appName)
	if err != nil {
		return err
	}
	if !hasIgs {
		fmt.Fprintln(w, "Creating ingress")
		if err := k.createIngress(namespace, appName, vHost); err != nil {
			return err
		}
	}

	return nil
}

func (k *Client) DeletePod(namespace, podName string) error {
	kc, err := k.buildClient()
	if err != nil {
		return err
	}
	err = kc.Pods(namespace).Delete(podName, &metav1.DeleteOptions{})
	return errors.Wrap(err, "could not delete pod")
}

func (k *Client) waitPodStart(pod *k8sv1.Pod, checkInterval, timeout time.Duration) error {
	kc, err := k.buildClient()
	if err != nil {
		return err
	}
	podsClient := kc.Pods(pod.Namespace)
	return wait.PollImmediate(checkInterval, timeout, func() (bool, error) {
		p, err := podsClient.Get(pod.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if p.Status.Phase == k8sv1.PodFailed {
			return true, ErrPodRunFailed
		}
		result := p.Status.Phase == k8sv1.PodRunning || p.Status.Phase == k8sv1.PodSucceeded
		return result, nil
	})
}

func (k *Client) waitPodEnd(pod *k8sv1.Pod, checkInterval, timeout time.Duration) error {
	kc, err := k.buildClient()
	if err != nil {
		return err
	}
	podsClient := kc.Pods(pod.Namespace)
	return wait.PollImmediate(checkInterval, timeout, func() (bool, error) {
		p, err := podsClient.Get(pod.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		result := p.Status.Phase == k8sv1.PodSucceeded || p.Status.Phase == k8sv1.PodFailed
		return result, nil
	})
}

func (k *Client) podExitCode(pod *k8sv1.Pod) (int, error) {
	kc, err := k.buildClient()
	if err != nil {
		return 1, err
	}

	p, err := kc.Pods(pod.Namespace).Get(pod.Name, metav1.GetOptions{})
	if err != nil {
		return 1, err
	}
	for _, containerStatus := range p.Status.ContainerStatuses {
		state := containerStatus.State.Terminated
		if state == nil {
			continue
		}
		return int(state.ExitCode), nil
	}
	return 1, ErrPodStillRunning
}

func (k *Client) currentPodReplicasFromDeploy(namespace, appName string) int32 {
	kc, err := k.buildClient()
	if err != nil {
		return 1
	}

	d, err := kc.AppsV1beta1().Deployments(
		namespace).Get(appName, metav1.GetOptions{})
	if err != nil || d.Status.Replicas < 1 {
		return 1
	}
	return d.Status.Replicas
}

func (k *Client) SetNamespaceAnnotations(namespace string, annotations map[string]string) error {
	kc, err := k.buildClient()
	if err != nil {
		return err
	}

	ns, err := k.getNamespace(namespace)
	if err != nil {
		return err
	}

	for key, value := range annotations {
		ns.Annotations[key] = value
	}
	_, err = kc.CoreV1().Namespaces().Update(ns)
	return err
}

func (k *Client) SetNamespaceLabels(namespace string, labels map[string]string) error {
	kc, err := k.buildClient()
	if err != nil {
		return err
	}

	ns, err := k.getNamespace(namespace)
	if err != nil {
		return err
	}

	for key, value := range labels {
		ns.Labels[key] = value
	}
	_, err = kc.CoreV1().Namespaces().Update(ns)
	return err
}

func (c Client) getDeployContainerList(namespace, deploy string) ([]string, error) {
	kc, err := c.buildClient()
	if err != nil {
		return nil, err
	}

	d, err := kc.AppsV1beta1().
		Deployments(namespace).
		Get(deploy, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get deploy spec")
	}

	containers := make([]string, len(d.Spec.Template.Spec.Containers))
	for i, c := range d.Spec.Template.Spec.Containers {
		containers[i] = c.Name
	}
	return containers, nil
}

func prepareEnvVarsPath(name, template string, containers []string, v interface{}) ([]byte, error) {
	type containerEnvVars struct {
		Name string      `json:"name"`
		Env  interface{} `json:"env"`
	}
	containersEnvVars := make([]containerEnvVars, len(containers))
	for i, container := range containers {
		containersEnvVars[i] = containerEnvVars{
			Name: container, Env: v,
		}
	}
	b, err := json.Marshal(containersEnvVars)
	if err != nil {
		return nil, errors.Wrap(err, "failed to json encode env vars")
	}

	data := fmt.Sprintf(template, time.Now(), string(b))
	return []byte(data), nil
}

func (c *Client) patchDeployEnvVars(namespace, name string, v interface{}) error {
	containers, err := c.getDeployContainerList(namespace, name)
	if err != nil {
		return err
	}

	data, err := prepareEnvVarsPath(name, patchDeployEnvVarsTmpl, containers, v)
	if err != nil {
		return err
	}

	kc, err := c.buildClient()
	if err != nil {
		return err
	}

	_, err = kc.ExtensionsV1beta1().Deployments(namespace).Patch(
		name,
		types.StrategicMergePatchType,
		data,
	)

	return errors.Wrap(err, "patch deploy failed")
}

func (c *Client) patchCronJobEnvVars(namespace, name string, v interface{}) error {
	data, err := prepareEnvVarsPath(name, patchCronJobEnvVarsTmpl, []string{name}, v)
	if err != nil {
		return err
	}

	kc, err := c.buildClient()
	if err != nil {
		return err
	}

	_, err = kc.CronJobs(namespace).Patch(
		name,
		types.StrategicMergePatchType,
		data,
	)

	return errors.Wrap(err, "patch cronjob failed")
}

func convertAppSecretEnvVar(secretName string, secrets []string) interface{} {
	type secretKeyRef struct {
		Key  string `json:"key"`
		Name string `json:"name"`
	}
	type valueFrom struct {
		SecretKeyRef secretKeyRef `json:"secretKeyRef"`
	}
	type envVar struct {
		Name      string    `json:"name"`
		ValueFrom valueFrom `json:"valueFrom"`
	}

	env := make([]*envVar, len(secrets))
	for i := range secrets {
		env[i] = &envVar{
			Name: secrets[i],
			ValueFrom: valueFrom{
				SecretKeyRef: secretKeyRef{
					Key:  secrets[i],
					Name: secretName,
				},
			},
		}
	}

	return env
}

func convertAppEnvVar(evs []*app.EnvVar) interface{} {
	type EnvVar struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}
	env := make([]*EnvVar, len(evs))
	for i := range evs {
		env[i] = &EnvVar{Name: evs[i].Key, Value: evs[i].Value}
	}

	return env
}

func convertAppDeleteEnvVar(evNames []string) interface{} {
	type EnvVar struct {
		Name  string `json:"name"`
		Patch string `json:"$patch"`
	}
	env := make([]*EnvVar, len(evNames))
	for i := range evNames {
		env[i] = &EnvVar{Name: evNames[i], Patch: "delete"}
	}

	return env
}

func (c *Client) CreateOrUpdateDeployEnvVars(namespace, name string, evs []*app.EnvVar) error {
	return c.patchDeployEnvVars(namespace, name, convertAppEnvVar(evs))
}

func (c *Client) CreateOrUpdateCronJobEnvVars(namespace, name string, evs []*app.EnvVar) error {
	return c.patchCronJobEnvVars(namespace, name, convertAppEnvVar(evs))
}

func (c *Client) CreateOrUpdateDeploySecretEnvVars(namespace, name, secretName string, secrets []string) error {
	return c.patchDeployEnvVars(namespace, name, convertAppSecretEnvVar(secretName, secrets))
}

func (c *Client) CreateOrUpdateCronJobSecretEnvVars(namespace, name, secretName string, secrets []string) error {
	return c.patchCronJobEnvVars(namespace, name, convertAppSecretEnvVar(secretName, secrets))
}

func (k *Client) DeleteDeployEnvVars(namespace, name string, evNames []string) error {
	return k.patchDeployEnvVars(namespace, name, convertAppDeleteEnvVar(evNames))
}

func (k *Client) DeleteCronJobEnvVars(namespace, name string, evNames []string) error {
	return k.patchCronJobEnvVars(namespace, name, convertAppDeleteEnvVar(evNames))
}

func (k *Client) DeleteNamespace(namespace string) error {
	kc, err := k.buildClient()
	if err != nil {
		return err
	}
	err = kc.CoreV1().Namespaces().Delete(namespace, &metav1.DeleteOptions{})
	return errors.Wrap(err, "delete ns failed")
}

func (k *Client) NamespaceListByLabel(label, value string) ([]string, error) {
	kc, err := k.buildClient()
	if err != nil {
		return nil, err
	}
	labelSelector := fmt.Sprintf("%s=%s", label, value)
	if value == "" {
		labelSelector = fmt.Sprintf("%s", label)
	}
	nl, err := kc.CoreV1().Namespaces().List(metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, err
	}
	namespaces := make([]string, 0)
	for _, item := range nl.Items {
		namespaces = append(namespaces, item.ObjectMeta.Name)
	}
	return namespaces, nil
}

func (k *Client) ReplicaSetListByLabel(namespace, label, value string) ([]*deploy.ReplicaSetListItem, error) {
	cli, err := k.buildClient()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build client")
	}

	labelSelector := fmt.Sprintf("%s=%s", label, value)
	opts := metav1.ListOptions{LabelSelector: labelSelector}
	rs, err := cli.ExtensionsV1beta1().ReplicaSets(namespace).List(opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get replicasets")
	}

	resp := make([]*deploy.ReplicaSetListItem, len(rs.Items))
	for i, item := range rs.Items {
		resp[i] = &deploy.ReplicaSetListItem{
			Revision:    item.Annotations[revisionAnnotation],
			CreatedAt:   item.CreationTimestamp.Time.String(),
			Current:     item.Status.ReadyReplicas > 0,
			Description: item.Annotations[changeCauseAnnotation],
		}
	}

	return resp, nil
}

func (k *Client) DeployRollbackToRevision(namespace, name, revision string) error {
	kc, err := k.buildClient()
	if err != nil {
		return err
	}

	data := fmt.Sprintf(patchDeployRollbackToRevisionTmpl, revision)

	_, err = kc.ExtensionsV1beta1().Deployments(namespace).Patch(
		name,
		types.StrategicMergePatchType,
		[]byte(data),
	)

	return errors.Wrap(err, "patch deploy failed")
}

func (k *Client) DeploySetReplicas(namespace, name string, replicas int32) error {
	kc, err := k.buildClient()
	if err != nil {
		return err
	}

	data := fmt.Sprintf(patchDeployReplicasTmpl, replicas)

	_, err = kc.ExtensionsV1beta1().Deployments(namespace).Patch(
		name,
		types.StrategicMergePatchType,
		[]byte(data),
	)

	return errors.Wrap(err, "patch deploy failed")
}

func (c *Client) CloudProviderName() (string, error) {
	kc, err := c.buildClient()
	if err != nil {
		return "", err
	}
	ls := "kubernetes.io/role=master"
	nodes, err := kc.CoreV1().Nodes().List(metav1.ListOptions{LabelSelector: ls})
	if err != nil {
		return "", errors.Wrap(err, "node list failed")
	}
	if len(nodes.Items) == 0 {
		return "", errors.New("empty cluster")
	}
	id := nodes.Items[0].Spec.ProviderID
	idx := strings.Index(id, "://")
	if idx <= 0 {
		return "", errors.New("invalid provider id")
	}
	return id[:idx], nil
}

func (c *Client) SetServiceAnnotations(namespace, svcName string, annotations map[string]string) error {
	return c.patchServiceAnnotations(namespace, svcName, annotations)
}

func (c *Client) UpdateServicePorts(namespace, svcName string, ports []spec.ServicePort) error {
	kc, err := c.buildClient()
	if err != nil {
		return err
	}
	svc, err := kc.CoreV1().Services(namespace).Get(svcName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	k8sPorts := servicePortsToK8sServicePorts(ports)
	svc.Spec.Ports = k8sPorts
	_, err = kc.CoreV1().Services(namespace).Update(svc)
	return errors.Wrap(err, "update service failed")
}

func (c *Client) patchServiceAnnotations(namespace, svcName string, annotations map[string]string) error {
	data, err := prepareServiceAnnotations(patchServiceAnnotationsTmpl, annotations)
	if err != nil {
		return err
	}
	kc, err := c.buildClient()
	if err != nil {
		return err
	}
	_, err = kc.CoreV1().Services(namespace).Patch(
		svcName,
		types.StrategicMergePatchType,
		data,
	)
	return errors.Wrap(err, "patch namespace failed")
}

func (c *Client) ServiceAnnotations(namespace, svcName string) (map[string]string, error) {
	kc, err := c.buildClient()
	if err != nil {
		return nil, err
	}
	svc, err := kc.CoreV1().Services(namespace).Get(svcName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return svc.Annotations, nil
}

func (c *Client) ServicePorts(namespace, svcName string) ([]*spec.ServicePort, error) {
	kc, err := c.buildClient()
	if err != nil {
		return nil, err
	}
	svc, err := kc.CoreV1().Services(namespace).Get(svcName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	ports := make([]*spec.ServicePort, len(svc.Spec.Ports))
	for i := range svc.Spec.Ports {
		ports[i] = &spec.ServicePort{
			Port:       int(svc.Spec.Ports[i].Port),
			Name:       svc.Spec.Ports[i].Name,
			TargetPort: int(svc.Spec.Ports[i].TargetPort.IntVal),
		}
	}
	return ports, nil
}

func (c *Client) ContainerExplicitEnvVars(namespace, deployName, containerName string) ([]*app.EnvVar, error) {
	kc, err := c.buildClient()
	if err != nil {
		return nil, err
	}
	deploy, err := kc.AppsV1beta1().
		Deployments(namespace).
		Get(deployName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "get deploy failed")
	}
	var con k8sv1.Container
	s := deploy.Spec.Template.Spec.Containers
	for _, c := range s {
		if c.Name == containerName {
			con = c
			break
		}
	}
	return k8sExplicitEnvToAppEnv(con.Env), nil
}

func (c *Client) WatchDeploy(namespace, deployName string) error {
	kc, err := c.buildClient()
	if err != nil {
		return err
	}
	opts := metav1.ListOptions{
		Watch:         true,
		FieldSelector: fmt.Sprintf("metadata.name=%s", deployName),
	}
	ts := time.Now()
	for {
		w, err := kc.AppsV1beta1().Deployments(namespace).Watch(opts)
		if err != nil {
			return errors.Wrap(err, "watch deploy failed")
		}
		ch := watch.Filter(w, filterDeployEvents).ResultChan()
	inner:
		for {
			select {
			case ev, ok := <-ch:
				if !ok {
					break inner
				}
				d := ev.Object.(*v1beta1.Deployment)
				cond := d.Status.Conditions[len(d.Status.Conditions)-1]
				if cond.LastUpdateTime.After(ts) && isRollingUpdateFinished(cond) {
					return nil
				} else if isRollingUpdateStalled(cond) {
					return errors.New("rolling update stalled, still running the old deploy")
				}
			}
		}
	}
}

func prepareServiceAnnotations(tmpl string, annotations map[string]string) ([]byte, error) {
	b, err := json.Marshal(annotations)
	if err != nil {
		return nil, errors.Wrap(err, "failed to json encode")
	}
	data := fmt.Sprintf(tmpl, string(b))
	return []byte(data), nil
}

func filterDeployEvents(in watch.Event) (watch.Event, bool) {
	_, ok := in.Object.(*v1beta1.Deployment)
	if !ok || (string(in.Type) != "MODIFIED" && string(in.Type) != "ADDED") {
		return in, false
	}
	return in, true
}

func isRollingUpdateFinished(cond v1beta1.DeploymentCondition) bool {
	return string(cond.Status) == "True" && cond.Reason == "NewReplicaSetAvailable"
}

func isRollingUpdateStalled(cond v1beta1.DeploymentCondition) bool {
	return string(cond.Status) == "False" && cond.Reason == "ProgressDeadlineExceeded"
}

func newInClusterK8sClient(conf *Config) (*Client, error) {
	k8sConf, err := restclient.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return &Client{
		conf:    k8sConf,
		ingress: conf.Ingress,
	}, nil
}

func newOutOfClusterK8sClient(conf *Config) (*Client, error) {
	k8sConf, err := clientcmd.BuildConfigFromFlags("", conf.ConfigFile)
	if err != nil {
		return nil, err
	}
	return &Client{
		conf:          k8sConf,
		podRunTimeout: conf.PodRunTimeout,
	}, nil
}
