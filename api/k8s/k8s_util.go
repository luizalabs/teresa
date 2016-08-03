package k8s

import (
	"fmt"
	"time"

	"github.com/pborman/uuid"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/util/wait"
)

const (
	slugBuilderName = "deis-slugbuilder"
	tarPath         = "TAR_PATH"
	putPath         = "PUT_PATH"
	debugKey        = "DEIS_DEBUG"
	builderStorage  = "BUILDER_STORAGE"
	objectStore     = "s3-storage"
	objectStorePath = "/var/run/secrets/deis/objectstore/creds"
)

// SlugBuilderPodName is used to generate a temp name to the builder pod
func SlugBuilderPodName(appName, shortSha string) string {
	uid := uuid.New()[:8]
	return fmt.Sprintf("slugbuild-%s-%s-%s", appName, shortSha, uid)
}

// BuildSlugbuilderPod is used to create an builder pod
func BuildSlugbuilderPod(
	debug bool,
	name,
	namespace,
	tarKey,
	putKey,
	buildpackURL string) *api.Pod {
	bn := fmt.Sprintf("slugbuilder-%s", name)
	e := make(map[string]interface{})
	e["BUILDER_STORAGE"] = "s3"
	e["TAR_PATH"] = tarKey
	e["PUT_PATH"] = putKey
	if debug {
		e["DEIS_DEBUG"] = 1
	}
	if buildpackURL != "" {
		e["BUILDPACK_URL"] = buildpackURL
	}
	c := buildContainer("deis-slugbuilder", "luizalabs/slugbuilder:git-923c9f8",
		api.PullIfNotPresent, nil, e)
	addVolMountToContainer(c, "storage-keys", "/var/run/secrets/deis/objectstore/creds", true)
	podSpec := buildPodSpec(api.RestartPolicyNever, []api.Container{*c})
	addVolSecretToPodSpec(podSpec, "storage-keys", "s3-storage")
	l := map[string]string{
		"heritage": bn,
	}
	p := BuildPod(bn, namespace, l, podSpec)

	return p
}

// WaitForPod waits for running stated, among others
func WaitForPod(c *client.Client, ns, podName string, interval, timeout time.Duration) error {
	condition := func(pod *api.Pod) (bool, error) {
		if pod.Status.Phase == api.PodRunning {
			return true, nil
		}
		if pod.Status.Phase == api.PodSucceeded {
			return true, nil
		}
		if pod.Status.Phase == api.PodFailed {
			return true, fmt.Errorf("Giving up; pod went into failed status: \n%s", fmt.Sprintf("%#v", pod))
		}
		return false, nil
	}
	err := waitForPodCondition(c, ns, podName, condition, interval, timeout)
	return err
}

// WaitForPodEnd waits for a pod in state succeeded or failed
func WaitForPodEnd(c *client.Client, ns, podName string, interval, timeout time.Duration) error {
	condition := func(pod *api.Pod) (bool, error) {
		if pod.Status.Phase == api.PodSucceeded {
			return true, nil
		}
		if pod.Status.Phase == api.PodFailed {
			return true, nil
		}
		return false, nil
	}

	return waitForPodCondition(c, ns, podName, condition, interval, timeout)
}

// waitForPodCondition waits for a pod in state defined by a condition (func)
func waitForPodCondition(c *client.Client, ns, podName string, condition func(pod *api.Pod) (bool, error),
	interval, timeout time.Duration) error {
	return wait.PollImmediate(interval, timeout, func() (bool, error) {
		pods, err := c.Pods(ns).List(api.ListOptions{LabelSelector: labels.Set{"heritage": podName}.AsSelector()})
		if err != nil || len(pods.Items) == 0 {
			return false, nil
		}

		done, err := condition(&pods.Items[0])
		if err != nil {
			return false, err
		}
		if done {
			return true, nil
		}

		return false, nil
	})
}

// BuildSlugRunnerDeployment builds a deployment using a slugrunner image
// with the slug built by the slugbuilder some steps earlier.
func BuildSlugRunnerDeployment(
	name string,
	namespace string,
	maxUnavailable int,
	maxSurge int,
	replicas int,
	selector string,
	slugURL string,
	env map[string]string) *extensions.Deployment {
	// TODO: we should remove the selector and use the name
	e := make(map[string]interface{})
	e["PORT"] = "5000"
	e["BUILDER_STORAGE"] = "s3"
	e["DEIS_DEBUG"] = true
	e["SLUG_URL"] = slugURL
	for k, v := range env {
		e[k] = v
	}
	c := buildContainer("deis-slugrunner", "luizalabs/slugrunner:git-044f85c",
		api.PullIfNotPresent, []string{"start", "web"}, e)
	addVolMountToContainer(c, "storage-keys", "/var/run/secrets/deis/objectstore/creds", true)
	podSpec := buildPodSpec(api.RestartPolicyAlways, []api.Container{*c})
	addVolSecretToPodSpec(podSpec, "storage-keys", "s3-storage")

	return BuildDeployment(name, namespace, maxUnavailable, maxSurge, replicas, selector, *podSpec)
}

// BuildSlugRunnerLBService helps to create a slugrunner service pointing to port 5000
func BuildSlugRunnerLBService(name, namespace, selector string) *api.Service {
	s := BuildLoadBalancerService(name, namespace, selector)
	AddPortConfigToService(s, "", api.ProtocolTCP, 5000, 5000)
	return s
}
