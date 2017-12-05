package k8s

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/deploy"
	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	k8sv1 "k8s.io/client-go/pkg/api/v1"
	asv1 "k8s.io/client-go/pkg/apis/autoscaling/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	patchDeployEnvVarsTmpl            = `{"spec":{"template":{"spec":{"containers":[{"name": "%s", "env":%s}]}}}}`
	patchDeployRollbackToRevisionTmpl = `{"spec":{"rollbackTo":{"revision": %s}}}`
	revisionAnnotation                = "deployment.kubernetes.io/revision"
)

type k8sClient struct {
	conf               *restclient.Config
	defaultServiceType string
	podRunTimeout      time.Duration
	ingress            bool
}

func (k *k8sClient) buildClient() (*kubernetes.Clientset, error) {
	c, err := kubernetes.NewForConfig(k.conf)
	if err != nil {
		return nil, errors.Wrap(err, "create k8s client failed")
	}
	return c, nil
}

func (k *k8sClient) HealthCheck() error {
	kc, err := k.buildClient()
	if err != nil {
		return err
	}
	_, err = kc.CoreV1().Namespaces().List(metav1.ListOptions{})
	return err
}

func (k *k8sClient) getNamespace(namespace string) (*k8sv1.Namespace, error) {
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

func (k *k8sClient) NamespaceAnnotation(namespace, annotation string) (string, error) {
	ns, err := k.getNamespace(namespace)
	if err != nil {
		return "", errors.Wrap(err, "get annotation failed")
	}

	return ns.Annotations[annotation], nil
}

func (k k8sClient) NamespaceLabel(namespace, label string) (string, error) {
	ns, err := k.getNamespace(namespace)
	if err != nil {
		return "", errors.Wrap(err, "get label failed")
	}

	return ns.Labels[label], nil
}

func (k *k8sClient) PodList(namespace string) ([]*app.Pod, error) {
	kc, err := k.buildClient()
	if err != nil {
		return nil, err
	}
	podList, err := kc.CoreV1().Pods(namespace).List(metav1.ListOptions{})
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

func (k *k8sClient) PodLogs(namespace string, podName string, lines int64, follow bool) (io.ReadCloser, error) {
	kc, err := k.buildClient()
	if err != nil {
		return nil, err
	}
	req := kc.CoreV1().Pods(namespace).GetLogs(
		podName,
		&k8sv1.PodLogOptions{
			Follow:    follow,
			TailLines: &lines,
		},
	)

	return req.Stream()
}

func newNs(nsName, teamName, user string) *k8sv1.Namespace {
	return &k8sv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: nsName,
			Labels: map[string]string{
				app.TeresaTeamLabel: teamName,
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

func (k *k8sClient) CreateNamespace(a *app.App, user string) error {
	kc, err := k.buildClient()
	if err != nil {
		return err
	}

	ns := newNs(a.Name, a.Team, user)
	if err := addAppToNs(a, ns); err != nil {
		return err
	}

	_, err = kc.CoreV1().Namespaces().Create(ns)
	return err
}

func (k *k8sClient) CreateNamespaceFromName(nsName, teamName, user string) error {
	kc, err := k.buildClient()
	if err != nil {
		return err
	}

	ns := newNs(nsName, teamName, user)

	_, err = kc.CoreV1().Namespaces().Create(ns)
	return err
}

func (k *k8sClient) CreateQuota(a *app.App) error {
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

func (k *k8sClient) CreateSecret(appName, secretName string, data map[string][]byte) error {
	kc, err := k.buildClient()
	if err != nil {
		return err
	}

	s := &k8sv1.Secret{
		Type: k8sv1.SecretTypeOpaque,
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: appName,
		},
		Data: data,
	}

	_, err = kc.CoreV1().Secrets(appName).Create(s)
	return err
}

func (k *k8sClient) CreateOrUpdateAutoscale(a *app.App) error {
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

func (k *k8sClient) AddressList(namespace string) ([]*app.Address, error) {
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

func (k *k8sClient) Status(namespace string) (*app.Status, error) {
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
		return nil, errors.Wrap(err, "get status failed")
	}

	pods, err := k.PodList(namespace)
	if err != nil {
		return nil, errors.Wrap(err, "get status failed")
	}

	var cpu int32
	if hpa.Status.CurrentCPUUtilizationPercentage != nil {
		cpu = *hpa.Status.CurrentCPUUtilizationPercentage
	}

	stat := &app.Status{
		CPU:  cpu,
		Pods: pods,
	}
	return stat, nil
}

func (k *k8sClient) Autoscale(namespace string) (*app.Autoscale, error) {
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

func (k *k8sClient) Limits(namespace, name string) (*app.Limits, error) {
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

func (k *k8sClient) CreateOrUpdateDeploy(deploySpec *deploy.DeploySpec) error {
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

func (k *k8sClient) PodRun(podSpec *deploy.PodSpec) (io.ReadCloser, <-chan int, error) {
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
			return
		}

		stream, err := k.PodLogs(podSpec.Namespace, podSpec.Name, 10, true)
		if err != nil {
			return
		}
		io.Copy(w, stream)

		if err = k.waitPodEnd(pod, 3*time.Second, k.podRunTimeout); err != nil {
			return
		}

		exitCode, err := k.podExitCode(pod)
		if err != nil {
			return
		}
		exitCodeChan <- exitCode
		go k.killPod(pod)
	}()
	return r, exitCodeChan, nil
}

func (k *k8sClient) hasService(namespace, appName string) (bool, error) {
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

func (k *k8sClient) createService(namespace, appName string) error {
	kc, err := k.buildClient()
	if err != nil {
		return err
	}
	srvSpec := serviceSpec(namespace, appName, k.defaultServiceType)
	_, err = kc.CoreV1().Services(namespace).Create(srvSpec)
	return errors.Wrap(err, "create service failed")
}

func (k *k8sClient) hasIngress(namespace, appName string) (bool, error) {
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

func (k *k8sClient) createIngress(namespace, appName, vHost string) error {
	kc, err := k.buildClient()
	if err != nil {
		return err
	}
	igsSpec := ingressSpec(namespace, appName, vHost)
	_, err = kc.ExtensionsV1beta1().Ingresses(namespace).Create(igsSpec)
	return errors.Wrap(err, "create ingress failed")
}

// ExposeDeploy creates a service and/or a ingress if needed
func (k *k8sClient) ExposeDeploy(namespace, appName, vHost string, w io.Writer) error {
	hasSrv, err := k.hasService(namespace, appName)
	if err != nil {
		return err
	}
	if !hasSrv {
		fmt.Fprintln(w, "Exposing service")
		if err := k.createService(namespace, appName); err != nil {
			return err
		}
	}

	if !k.ingress {
		return nil
	}
	hasIgs, err := k.hasIngress(namespace, appName)
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

func (k *k8sClient) killPod(pod *k8sv1.Pod) error {
	kc, err := k.buildClient()
	if err != nil {
		return err
	}
	return kc.Pods(pod.Namespace).Delete(pod.Name, &metav1.DeleteOptions{})
}

func (k *k8sClient) waitPodStart(pod *k8sv1.Pod, checkInterval, timeout time.Duration) error {
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

func (k *k8sClient) waitPodEnd(pod *k8sv1.Pod, checkInterval, timeout time.Duration) error {
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

func (k *k8sClient) podExitCode(pod *k8sv1.Pod) (int, error) {
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

func (k *k8sClient) currentPodReplicasFromDeploy(namespace, appName string) int32 {
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

func (k *k8sClient) SetNamespaceAnnotations(namespace string, annotations map[string]string) error {
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

func (k *k8sClient) SetNamespaceLabels(namespace string, labels map[string]string) error {
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

func (k *k8sClient) patchDeployEnvVars(namespace, name string, v interface{}) error {
	kc, err := k.buildClient()
	if err != nil {
		return err
	}
	b, err := json.Marshal(v)
	if err != nil {
		return errors.Wrap(err, "failed to json encode env vars")
	}
	data := fmt.Sprintf(patchDeployEnvVarsTmpl, name, string(b))

	_, err = kc.ExtensionsV1beta1().Deployments(namespace).Patch(
		name,
		types.StrategicMergePatchType,
		[]byte(data),
	)

	return errors.Wrap(err, "patch deploy failed")
}

func (k *k8sClient) DeleteDeployEnvVars(namespace, name string, evNames []string) error {
	type EnvVar struct {
		Name  string `json:"name"`
		Patch string `json:"$patch"`
	}
	env := make([]*EnvVar, len(evNames))
	for i := range evNames {
		env[i] = &EnvVar{Name: evNames[i], Patch: "delete"}
	}

	return k.patchDeployEnvVars(namespace, name, env)
}

func (k *k8sClient) CreateOrUpdateDeployEnvVars(namespace, name string, evs []*app.EnvVar) error {
	type EnvVar struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}
	env := make([]*EnvVar, len(evs))
	for i := range evs {
		env[i] = &EnvVar{Name: evs[i].Key, Value: evs[i].Value}
	}

	return k.patchDeployEnvVars(namespace, name, env)
}

func (k *k8sClient) DeleteNamespace(namespace string) error {
	kc, err := k.buildClient()
	if err != nil {
		return err
	}
	err = kc.CoreV1().Namespaces().Delete(namespace, &metav1.DeleteOptions{})
	return errors.Wrap(err, "delete ns failed")
}

func (k k8sClient) NamespaceListByLabel(label, value string) ([]string, error) {
	kc, err := k.buildClient()
	if err != nil {
		return nil, err
	}
	labelSelector := fmt.Sprintf("%s=%s", label, value)
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

func (k *k8sClient) ReplicaSetListByLabel(namespace, label, value string) ([]*deploy.ReplicaSetListItem, error) {
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
			Age:         int64(time.Since(item.CreationTimestamp.Time)),
			Current:     item.Status.ReadyReplicas > 0,
			Description: item.Annotations[changeCauseAnnotation],
		}
	}

	return resp, nil
}

func (k *k8sClient) DeployRollbackToRevision(namespace, name, revision string) error {
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

func newInClusterK8sClient(conf *Config) (Client, error) {
	k8sConf, err := restclient.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return &k8sClient{
		conf:               k8sConf,
		defaultServiceType: conf.DefaultServiceType,
		ingress:            conf.Ingress,
	}, nil
}

func newOutOfClusterK8sClient(conf *Config) (Client, error) {
	k8sConf, err := clientcmd.BuildConfigFromFlags("", conf.ConfigFile)
	if err != nil {
		return nil, err
	}
	return &k8sClient{
		conf:               k8sConf,
		defaultServiceType: conf.DefaultServiceType,
		podRunTimeout:      conf.PodRunTimeout,
	}, nil
}
