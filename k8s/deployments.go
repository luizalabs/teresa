package k8s

import (
	"fmt"
	"io"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/go-openapi/runtime"
	"github.com/luizalabs/teresa-api/helpers"
	"github.com/luizalabs/teresa-api/models"
	"github.com/pborman/uuid"
	"k8s.io/kubernetes/pkg/api"
	k8s_errors "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/autoscaling"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/util/intstr"
	"k8s.io/kubernetes/pkg/util/wait"
)

// DeploymentsInterface is used to allow mock testing
type DeploymentsInterface interface {
	Deployments() DeploymentInterface
}

// DeploymentInterface is used to interact with Kubernetes and also to allow mock testing
type DeploymentInterface interface {
	Get(appName string) (d *extensions.Deployment, err error)
	CreateWelcomeDeployment(app *models.App) error
	Create(appName, description string, file *runtime.File, storage helpers.Storage, tk *Token) (io.ReadCloser, error)
	Update(app *models.App, description string, storage helpers.Storage) error
	CreateAutoScale(app *models.App) error
	UpdateAutoScale(app *models.App) error
}

type deployments struct {
	k *k8sHelper
}

func newDeployments(c *k8sHelper) *deployments {
	return &deployments{k: c}
}

type deploy struct {
	uuid        string
	appName     string
	description string
	storageIn   string
	storageOut  string
	slugPath    string
}

func newDeploy(appName, description string) *deploy {
	d := &deploy{
		uuid: uuid.New()[:8],
	}
	d.description = description
	d.storageIn = fmt.Sprintf("deploys/%s/%s/in/app.tar.gz", appName, d.uuid)
	d.storageOut = fmt.Sprintf("deploys/%s/%s/out", appName, d.uuid)
	d.slugPath = fmt.Sprintf("%s/slug.tgz", d.storageOut)
	return d
}

func (c deployments) CreateWelcomeDeployment(app *models.App) error {
	d := newWelcomeDeployment(app)
	_, err := c.k.k8sClient.Deployments(*app.Name).Create(d)
	return err
}

// Create creates a new deployment for the App
func (c deployments) Create(appName, description string, file *runtime.File, storage helpers.Storage, tk *Token) (io.ReadCloser, error) {
	// get app info...
	app, err := c.k.Apps().Get(appName, tk)
	if err != nil {
		return nil, err
	}
	// check token...
	if tk.IsAuthorized(*app.Team) == false {
		msg := "token not allowed to do a deployment"
		return nil, NewUnauthorizedError(msg)
	}
	// creating deployment params
	deploy := newDeploy(appName, description)
	// create log context with uuid and app name
	lc := log.WithField("app", appName).WithField("deployUUID", deploy.uuid)
	lc.Info("starting deploy...")
	// streaming actions...
	r, w := io.Pipe()
	go func() {
		defer w.Close()
		// upload file to storage
		lcu := lc.WithField("storage", storage.Type()).WithField("storageIn", deploy.storageIn)
		lcu.Debug("uploading app archive to storage...")
		fmt.Fprintln(w, "uploading app archive to storage...")
		if err := storage.UploadFile(deploy.storageIn, file); err != nil {
			lcu.WithError(err).Error("error found when upload app archive to storage")
			fmt.Fprintln(w, "error found when upload app archive to storage")
			return
		}
		lcu.Debug("upload done with success")
		fmt.Fprintln(w, "upload done with success")
		lcu = nil
		// building app...
		lc.Debug("building app...")
		fmt.Fprintln(w, "building app...")
		if err := c.buildApp(app, deploy, storage, w); err != nil {
			lc.WithError(err).Warn("error during build proccess")
			fmt.Fprintln(w, "error during build proccess")
			return
		}
		lc.Debug("build step done without errors")
		fmt.Fprintln(w, "build step done without errors")
		// updating deployment
		lc.Debug("updating deployment for rolling update...")
		fmt.Fprintln(w, "rolling update...")
		if err := c.updateDeployment(app, deploy.slugPath, deploy.description, storage); err != nil {
			lc.WithError(err).Error("error updating deployment")
			fmt.Fprintln(w, "error when doing rolling update")
			return
		}
		fmt.Fprintln(w, "deploy finished with success")
	}()
	return r, nil
}

func (c deployments) Get(appName string) (d *extensions.Deployment, err error) {
	d, err = c.k.k8sClient.Deployments(appName).Get(appName)
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			return nil, NewNotFoundErrorf(`deployment "%s" not found`, appName)
		}
	}
	return
}

// buildApp creates a builder POD to build the App, waits the POD to be completed
func (c deployments) buildApp(app *models.App, deploy *deploy, storage helpers.Storage, w io.Writer) error {
	// TODO: fix times for wait start and wait end

	// creating builder POD
	pod, err := c.createBuilderPod(app, deploy, storage)
	if err != nil {
		return err
	}
	// wainting POD to start the builder proccess
	if err = c.waitPodStart(pod, 1*time.Second, 2*time.Minute); err != nil {
		return err
	}
	// stream output log from the builder POD
	s, err := c.streamPodOutput(pod)
	if err != nil {
		return err
	}
	defer s.Close()
	io.Copy(w, s)
	// wait POD finish
	if err = c.waitPodEnd(pod, 1*time.Second, 2*time.Minute); err != nil {
		return err
	}
	// get POD exit code.
	ec, err := c.podExitCode(pod)
	if err != nil {
		return err
	}
	// if any code diff from 0, return build error
	if *ec != 0 {
		return NewAppBuilderErrorf(`builder POD "%s/%s" exited with code %d`, pod.Namespace, pod.Name, *ec)
	}
	return nil
}

// newBuilderPodYaml creates a POD specification (input yaml).
// The returned value will be used to create a builder POD on kubernetes.
// This POD receives a tarball (App tarball) path from a storage server, gets this tarball,
// builds the App, and put the built App on the output path on the storage server.
func newBuilderPod(app *models.App, deploy *deploy, storage helpers.Storage) *api.Pod {
	name := fmt.Sprintf("build-%s", deploy.uuid)
	// create container yaml
	c := api.Container{
		Name:            name,
		ImagePullPolicy: api.PullIfNotPresent,
		Image:           "luizalabs/slugbuilder:git-923c9f8",
		Env: []api.EnvVar{
			api.EnvVar{
				Name:  "TAR_PATH",
				Value: deploy.storageIn,
			},
			api.EnvVar{
				Name:  "PUT_PATH",
				Value: deploy.storageOut,
			},
			api.EnvVar{
				Name:  "BUILDER_STORAGE",
				Value: storage.Type(),
			},
		},
	}
	// load app env vars
	for _, e := range app.EnvVars {
		ce := api.EnvVar{
			Name:  *e.Key,
			Value: *e.Value,
		}
		c.Env = append(c.Env, ce)
	}
	// add volume mount to container yaml (to access app archive from storage)
	c.VolumeMounts = append(c.VolumeMounts, api.VolumeMount{
		Name:      "storage-keys",
		MountPath: "/var/run/secrets/deis/objectstore/creds",
		ReadOnly:  true,
	})
	// create pod specification yaml with secret attached
	p := api.PodSpec{
		RestartPolicy: api.RestartPolicyNever,
		Containers: []api.Container{
			c,
		},
	}
	// add volume (with storage keys)  to pod specification yaml
	v := api.Volume{
		Name: "storage-keys",
	}
	v.Secret = &api.SecretVolumeSource{
		SecretName: storage.GetK8sSecretName(),
	}
	p.Volumes = []api.Volume{v}
	// create builder pod
	pod := &api.Pod{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:      name,
			Namespace: *app.Name,
		},
		Spec: p,
	}
	return pod
}

// createBuilderPod creates a POD that will build the App.
// This function receives the App tarbal in the deploy.storageIn path, builds the App
// and put the built version on the deploy.storageOut.
// The built version of the App could be used only with a "runner POD".
func (c deployments) createBuilderPod(app *models.App, deploy *deploy, storage helpers.Storage) (pod *api.Pod, err error) {
	// creates builder POD yaml
	podYaml := newBuilderPod(app, deploy, storage)
	// create builder POD
	pod, err = c.k.k8sClient.Pods(*app.Name).Create(podYaml)
	if err != nil {
		return nil, err
	}
	return
}

func (c deployments) waitPodStart(pod *api.Pod, checkInterval, timeout time.Duration) error {
	pg := c.k.k8sClient.Pods(pod.Namespace)
	return wait.PollImmediate(checkInterval, timeout, func() (bool, error) {
		// get the last state of the POD
		p, err := pg.Get(pod.Name)
		if err != nil {
			return false, fmt.Errorf(`Error when getting the updated POD state for POD "%s/%s". Err: %s`, pod.Namespace, pod.Name, err)
		}
		// update the received POD with the last state
		pod = p
		if pod.Status.Phase == api.PodFailed {
			return true, fmt.Errorf(`Pod "%s" went into failed status`, pod.Name)
		}
		if pod.Status.Phase == api.PodRunning || pod.Status.Phase == api.PodSucceeded {
			return true, nil
		}
		return false, nil
	})
}

func (c deployments) waitPodEnd(pod *api.Pod, checkInterval, timeout time.Duration) error {
	pg := c.k.k8sClient.Pods(pod.Namespace)
	return wait.PollImmediate(checkInterval, timeout, func() (bool, error) {
		// get the last state of the POD
		p, err := pg.Get(pod.Name)
		if err != nil {
			return false, fmt.Errorf(`Error when getting the updated POD state for POD "%s/%s". Err: %s`, pod.Namespace, pod.Name, err)
		}
		// update the received POD with the last state
		pod = p
		if pod.Status.Phase == api.PodSucceeded || pod.Status.Phase == api.PodFailed {
			return true, nil
		}
		return false, nil
	})
}

// streamPodOutput returns a io.ReadCloser with the output log from the POD.
func (c deployments) streamPodOutput(pod *api.Pod) (stream io.ReadCloser, err error) {
	req := c.k.k8sClient.Pods(pod.Namespace).GetLogs(pod.Name, &api.PodLogOptions{
		Follow: true,
	})
	if stream, err = req.Stream(); err != nil {
		return nil, fmt.Errorf(`error when trying to stream logs from builder POD "%s/%s". Err: %s`, pod.Namespace, pod.Name, err)
	}
	return
}

func (c deployments) podExitCode(pod *api.Pod) (code *int32, err error) {
	p, err := c.k.k8sClient.Pods(pod.Namespace).Get(pod.Name)
	if err != nil {
		return nil, fmt.Errorf(`error trying to discover the builder POD "%s/%s" exit code. Err: %s`, pod.Namespace, pod.Name, err)
	}
	for _, containerStatus := range p.Status.ContainerStatuses {
		state := containerStatus.State.Terminated
		if state.ExitCode != 0 {
			return &state.ExitCode, nil
		}
	}
	zero := int32(0)
	return &zero, nil
}

// newContainer is a helper to create a new container
func newContainer(name, image string) (c *api.Container) {
	c = &api.Container{
		Name:            name,
		ImagePullPolicy: api.PullIfNotPresent,
		Image:           image,
	}
	return
}

// appendContainerEnvVar appends a new env var to a container
func appendContainerEnvVar(c *api.Container, name, value string) {
	c.Env = append(c.Env, api.EnvVar{
		Name:  name,
		Value: value,
	})
}

// appendContainerVolumeMount is a helper to append a volume mount to a container
func appendContainerVolumeMount(c *api.Container, name, mountPath string, readOnly bool) {
	c.VolumeMounts = append(c.VolumeMounts, api.VolumeMount{
		Name:      name,
		ReadOnly:  readOnly,
		MountPath: mountPath,
	})
	return
}

func newWelcomeContainer(app *models.App) (c *api.Container) {
	c = newContainer(*app.Name, "luizalabs/hello-world:0.0.1")
	appendContainerEnvVar(c, "APP", *app.Name)
	appendContainerEnvVar(c, "PORT", "5000")
	return
}

func newSlugRunnerContainer(app *models.App, slug string, storageType string) (c *api.Container) {
	// creating runner container
	c = newContainer(*app.Name, "luizalabs/slugrunner:git-044f85c")
	c.Args = []string{"start", "web"}
	// appending env vars...
	// append App name to env var
	appendContainerEnvVar(c, "APP", *app.Name)
	appendContainerEnvVar(c, "PORT", "5000")
	appendContainerEnvVar(c, "BUILDER_STORAGE", storageType)
	appendContainerEnvVar(c, "SLUG_URL", slug)
	// appending app env vars
	for _, e := range app.EnvVars {
		appendContainerEnvVar(c, *e.Key, *e.Value)
	}
	// appending volume mount
	appendContainerVolumeMount(c, "storage-keys", "/var/run/secrets/deis/objectstore/creds", true)
	return
}

// newPodSpec creates a new Pod spec
func newPodSpec(c *api.Container) (ps *api.PodSpec) {
	ps = &api.PodSpec{
		RestartPolicy: api.RestartPolicyAlways,
		Containers: []api.Container{
			*c,
		},
	}
	return
}

// appendPodSpecSecretVolume appends a secret volume source to the pod
func appendPodSpecSecretVolume(ps *api.PodSpec, volumeName, secretName string) {
	ps.Volumes = []api.Volume{
		api.Volume{
			Name: volumeName,
			VolumeSource: api.VolumeSource{
				Secret: &api.SecretVolumeSource{
					SecretName: secretName,
				},
			},
		},
	}
	return
}

// newDeployment creates a new deployment
func newDeployment(app *models.App, ps *api.PodSpec) (d *extensions.Deployment) {
	// get rolling update values...
	var ruMaxSurge, ruMaxUnavailable intstr.IntOrString
	if v, err := strconv.Atoi(*app.RollingUpdate.MaxUnavailable); err != nil {
		ruMaxUnavailable = intstr.FromString(*app.RollingUpdate.MaxUnavailable)
	} else {
		ruMaxUnavailable = intstr.FromInt(v)
	}
	if v, err := strconv.Atoi(*app.RollingUpdate.MaxSurge); err != nil {
		ruMaxSurge = intstr.FromString(*app.RollingUpdate.MaxSurge)
	} else {
		ruMaxSurge = intstr.FromInt(v)
	}
	d = &extensions.Deployment{
		TypeMeta: unversioned.TypeMeta{
			APIVersion: "extensions/v1beta1",
			Kind:       "Deployment",
		},
		ObjectMeta: api.ObjectMeta{
			Name:      *app.Name,
			Namespace: *app.Name,
			Labels: map[string]string{
				"run": *app.Name,
			},
		},
		Spec: extensions.DeploymentSpec{
			Replicas: int32(app.Scale),
			Strategy: extensions.DeploymentStrategy{
				Type: extensions.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &extensions.RollingUpdateDeployment{
					MaxUnavailable: ruMaxUnavailable,
					MaxSurge:       ruMaxSurge,
				},
			},
			Template: api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Labels: map[string]string{
						"run": *app.Name,
					},
				},
				Spec: *ps,
			},
		},
	}
	return
}

// appendDeploymentAnnotation appends an annotation to deployment
func appendDeploymentAnnotation(d *extensions.Deployment, key, value string) {
	if d.ObjectMeta.Annotations == nil {
		d.ObjectMeta.Annotations = make(map[string]string)
	}
	d.ObjectMeta.Annotations[key] = value
}

// newSlugRunnerDeployment creates a new slug runner deployment based on the app info
func newSlugRunnerDeployment(app *models.App, slug, description string, storage helpers.Storage) (d *extensions.Deployment) {
	// creating slug runner container
	c := newSlugRunnerContainer(app, slug, storage.Type())
	// creating PodSpec
	ps := newPodSpec(c)
	// appending volume to PodSpec
	appendPodSpecSecretVolume(ps, "storage-keys", storage.GetK8sSecretName())
	// creating deployment yaml...
	d = newDeployment(app, ps)
	// appending annotations
	appendDeploymentAnnotation(d, "kubernetes.io/change-cause", description)
	appendDeploymentAnnotation(d, "teresa.io/slug", slug)
	return
}

func newWelcomeDeployment(app *models.App) (d *extensions.Deployment) {
	c := newWelcomeContainer(app)
	ps := newPodSpec(c)
	d = newDeployment(app, ps)
	return
}

func (c deployments) updateDeployment(app *models.App, slug, description string, storage helpers.Storage) error {
	d := newSlugRunnerDeployment(app, slug, description, storage)
	_, err := c.k.k8sClient.Deployments(*app.Name).Update(d)
	return err
}

func (c deployments) Update(app *models.App, description string, storage helpers.Storage) error {
	d, err := c.Get(*app.Name)
	if err != nil {
		return err
	}
	slug := d.Annotations["teresa.io/slug"]
	err = c.updateDeployment(app, slug, description, storage)
	return err
}

func newHorizontalPodAutoscaler(app *models.App) (hpa *autoscaling.HorizontalPodAutoscaler) {
	tcpu := int32(*app.AutoScale.CPUTargetUtilization)
	minr := int32(app.AutoScale.Min)
	hpa = &autoscaling.HorizontalPodAutoscaler{
		TypeMeta: unversioned.TypeMeta{
			APIVersion: "autoscaling/v1",
			Kind:       "HorizontalPodAutoscaler",
		},
		ObjectMeta: api.ObjectMeta{
			Name:      *app.Name,
			Namespace: *app.Name,
		},
		Spec: autoscaling.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscaling.CrossVersionObjectReference{
				APIVersion: "extensions/v1beta1",
				Kind:       "Deployment",
				Name:       *app.Name,
			},
			TargetCPUUtilizationPercentage: &tcpu,
			MaxReplicas:                    int32(app.AutoScale.Max),
			MinReplicas:                    &minr,
		},
	}
	return
}

func (c deployments) CreateAutoScale(app *models.App) error {
	hpa := newHorizontalPodAutoscaler(app)
	_, err := c.k.k8sClient.HorizontalPodAutoscalers(*app.Name).Create(hpa)
	return err
}

func (c deployments) UpdateAutoScale(app *models.App) error {
	hpa := newHorizontalPodAutoscaler(app)
	_, err := c.k.k8sClient.HorizontalPodAutoscalers(*app.Name).Update(hpa)
	return err
}
