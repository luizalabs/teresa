package k8s

import (
	"fmt"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/util/intstr"
)

func addVolMountToContainer(c *api.Container, n string, mp string, ro bool) {
	c.VolumeMounts = append(c.VolumeMounts, api.VolumeMount{
		Name:      n,
		ReadOnly:  ro,
		MountPath: mp,
	})
}

func addVolSecretToPodSpec(ps *api.PodSpec, n string, s string) {
	ps.Volumes = append(ps.Volumes, api.Volume{
		Name: n,
		VolumeSource: api.VolumeSource{
			Secret: &api.SecretVolumeSource{
				SecretName: s,
			},
		},
	})
}

func buildContainer(
	name string,
	image string,
	imagePullPolicy api.PullPolicy,
	args []string,
	env map[string]interface{}) *api.Container {
	c := api.Container{
		Name:            name,
		ImagePullPolicy: imagePullPolicy,
		Image:           image,
		Args:            args,
	}
	for k, v := range env {
		c.Env = append(c.Env, api.EnvVar{
			Name:  k,
			Value: fmt.Sprintf("%v", v),
		})
	}
	return &c
}

func buildPodSpec(rsPolicy api.RestartPolicy, containers []api.Container) *api.PodSpec {
	ps := api.PodSpec{
		RestartPolicy: rsPolicy,
		Containers:    containers,
	}
	return &ps
}

// BuildDeployment builds and returns a deployment pointer.
func BuildDeployment(
	name string,
	namespace string,
	maxUnavailable int,
	maxSurge int,
	replicas int,
	selector string,
	podSpec api.PodSpec) *extensions.Deployment {

	labels := map[string]string{
		"run": name,
	}
	mu := intstr.FromInt(maxUnavailable)
	ms := intstr.FromInt(maxSurge)
	r := int32(replicas)
	selectorMap := map[string]string{"run": selector}

	d := extensions.Deployment{
		// keep this here to create a clear resource to use with kubectl if you need
		TypeMeta: unversioned.TypeMeta{
			APIVersion: "extensions/v1beta1",
			Kind:       "Deployment",
		},
		ObjectMeta: api.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: extensions.DeploymentSpec{
			Replicas: r,
			Strategy: extensions.DeploymentStrategy{
				Type: extensions.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &extensions.RollingUpdateDeployment{
					MaxUnavailable: mu,
					MaxSurge:       ms,
				},
			},
			Template: api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Labels: labels,
				},
				Spec: podSpec,
			},
			Selector: &unversioned.LabelSelector{
				MatchLabels: selectorMap,
			},
		},
	}

	return &d
}

// BuildLoadBalancerService creates a service of type load balancer
func BuildLoadBalancerService(name, namespace, selector string) *api.Service {
	s := api.Service{
		TypeMeta: unversioned.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: api.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"run": name,
			},
		},
		Spec: api.ServiceSpec{
			Type:            api.ServiceTypeLoadBalancer,
			SessionAffinity: api.ServiceAffinityNone,
			Selector: map[string]string{
				"run": selector,
			},
			// Ports: servicePorts,
		},
		// Status: api.ServiceStatus{
		// 	LoadBalancer: api.LoadBalancerStatus{
		// 		Ingress: []api.LoadBalancerIngress{
		// 			api.LoadBalancerIngress{
		// 				Hostname: "adasdasdasdasd",
		// 			},
		// 		},
		// 	},
		// },
	}
	return &s
}

// AddPortConfigToService add port to service
func AddPortConfigToService(
	service *api.Service,
	name string,
	protocol api.Protocol,
	port int,
	targetPort int) {
	service.Spec.Ports = append(service.Spec.Ports, api.ServicePort{
		Name:       name,
		Protocol:   protocol,
		Port:       int32(port),
		TargetPort: intstr.FromInt(targetPort),
	})
}
