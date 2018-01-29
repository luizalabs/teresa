package k8s

import (
	"strconv"

	"github.com/luizalabs/teresa/pkg/server/spec"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	k8sv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/apps/v1beta1"
	k8s_extensions "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

const (
	changeCauseAnnotation = "kubernetes.io/change-cause"
	defaultServicePort    = 80
)

func podSpecToK8sContainer(podSpec *spec.Pod) (*k8sv1.Container, error) {
	c := &k8sv1.Container{
		Name:            podSpec.Name,
		ImagePullPolicy: k8sv1.PullIfNotPresent,
		Image:           podSpec.Image,
	}

	if podSpec.ContainerLimits != nil {
		cpu, err := resource.ParseQuantity(podSpec.ContainerLimits.CPU)
		if err != nil {
			return nil, err
		}
		memory, err := resource.ParseQuantity(podSpec.ContainerLimits.Memory)
		if err != nil {
			return nil, err
		}
		c.Resources = k8sv1.ResourceRequirements{
			Limits: k8sv1.ResourceList{
				k8sv1.ResourceCPU:    cpu,
				k8sv1.ResourceMemory: memory,
			},
		}
	}

	c.Args = append(c.Args, podSpec.Args...)

	for k, v := range podSpec.Env {
		c.Env = append(c.Env, k8sv1.EnvVar{Name: k, Value: v})
	}
	for _, vm := range podSpec.VolumeMounts {
		c.VolumeMounts = append(c.VolumeMounts, k8sv1.VolumeMount{
			Name:      vm.Name,
			MountPath: vm.MountPath,
			ReadOnly:  vm.ReadOnly,
		})
	}
	return c, nil
}

func podSpecVolumesToK8sVolumes(vols []*spec.PodVolume) []k8sv1.Volume {
	volumes := make([]k8sv1.Volume, 0)
	for _, v := range vols {
		vol := k8sv1.Volume{Name: v.Name}
		vol.Secret = &k8sv1.SecretVolumeSource{
			SecretName: v.SecretName,
		}
		volumes = append(volumes, vol)
	}
	return volumes
}

func podSpecToK8sPod(podSpec *spec.Pod) (*k8sv1.Pod, error) {
	c, err := podSpecToK8sContainer(podSpec)
	if err != nil {
		return nil, err
	}
	volumes := podSpecVolumesToK8sVolumes(podSpec.Volume)
	f := false

	ps := k8sv1.PodSpec{
		RestartPolicy: k8sv1.RestartPolicyNever,
		Containers:    []k8sv1.Container{*c},
		Volumes:       volumes,
		AutomountServiceAccountToken: &f,
	}

	pod := &k8sv1.Pod{
		TypeMeta: metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      podSpec.Name,
			Namespace: podSpec.Namespace,
		},
		Spec: ps,
	}
	return pod, nil
}

func deploySpecToK8sDeploy(deploySpec *spec.Deploy, replicas int32) (*v1beta1.Deployment, error) {
	c, err := podSpecToK8sContainer(&deploySpec.Pod)
	if err != nil {
		return nil, err
	}
	volumes := podSpecVolumesToK8sVolumes(deploySpec.Volume)

	if deploySpec.HealthCheck != nil {
		if deploySpec.HealthCheck.Liveness != nil {
			c.LivenessProbe = healthCheckProbeToK8sProbe(deploySpec.HealthCheck.Liveness)
		}
		if deploySpec.HealthCheck.Readiness != nil {
			c.ReadinessProbe = healthCheckProbeToK8sProbe(deploySpec.HealthCheck.Readiness)
		}
	}

	if deploySpec.Lifecycle != nil {
		c.Lifecycle = lifecycleToK8sLifecycle(deploySpec.Lifecycle)
	}

	f := false
	ps := k8sv1.PodSpec{
		RestartPolicy: k8sv1.RestartPolicyAlways,
		Containers:    []k8sv1.Container{*c},
		Volumes:       volumes,
		AutomountServiceAccountToken: &f,
	}

	var maxSurge, maxUnavailable *intstr.IntOrString
	if deploySpec.RollingUpdate != nil {
		vMaxSurge, vMaxUnavailable := rollingUpdateToK8sRollingUpdate(deploySpec.RollingUpdate)
		maxSurge, maxUnavailable = &vMaxSurge, &vMaxUnavailable
	}

	rhl := int32(deploySpec.RevisionHistoryLimit)
	d := &v1beta1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "extensions/v1beta1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploySpec.Name,
			Namespace: deploySpec.Namespace,
			Labels:    map[string]string{"run": deploySpec.Name},
			Annotations: map[string]string{
				changeCauseAnnotation: deploySpec.Description,
				spec.SlugAnnotation:   deploySpec.SlugURL,
			},
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas: &replicas,
			Strategy: v1beta1.DeploymentStrategy{
				Type: v1beta1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &v1beta1.RollingUpdateDeployment{
					MaxUnavailable: maxUnavailable,
					MaxSurge:       maxSurge,
				},
			},
			Template: k8sv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"run": deploySpec.Name},
				},
				Spec: ps,
			},
			RevisionHistoryLimit: &rhl,
		},
	}
	return d, nil
}

func rollingUpdateToK8sRollingUpdate(ru *spec.RollingUpdate) (maxSurge, maxUnavailable intstr.IntOrString) {
	conv := func(value string) intstr.IntOrString {
		v, err := strconv.Atoi(value)
		if err != nil {
			return intstr.FromString(value)
		}
		return intstr.FromInt(v)
	}
	return conv(ru.MaxSurge), conv(ru.MaxUnavailable)
}

func healthCheckProbeToK8sProbe(probe *spec.HealthCheckProbe) *k8sv1.Probe {
	return &k8sv1.Probe{
		InitialDelaySeconds: probe.InitialDelaySeconds,
		TimeoutSeconds:      probe.TimeoutSeconds,
		PeriodSeconds:       probe.PeriodSeconds,
		FailureThreshold:    probe.FailureThreshold,
		SuccessThreshold:    probe.SuccessThreshold,
		Handler: k8sv1.Handler{
			HTTPGet: &k8sv1.HTTPGetAction{
				Port: intstr.FromInt(spec.DefaultPort),
				Path: probe.Path,
			},
		},
	}
}

func lifecycleToK8sLifecycle(lc *spec.Lifecycle) *k8sv1.Lifecycle {
	k8sLc := new(k8sv1.Lifecycle)

	if lc.PreStop != nil {
		k8sLc.PreStop = &k8sv1.Handler{
			Exec: &k8sv1.ExecAction{
				Command: []string{"/bin/sleep", strconv.Itoa(lc.PreStop.DrainTimeoutSeconds)},
			},
		}
	}

	return k8sLc
}

func serviceSpec(namespace, name, srvType string) *k8sv1.Service {
	serviceType := k8sv1.ServiceType(srvType)
	return &k8sv1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"run": name,
			},
			Name:      name,
			Namespace: namespace,
		},
		Spec: k8sv1.ServiceSpec{
			Type:            serviceType,
			SessionAffinity: k8sv1.ServiceAffinityNone,
			Selector: map[string]string{
				"run": name,
			},
			Ports: []k8sv1.ServicePort{
				{
					Port:       80,
					Protocol:   k8sv1.ProtocolTCP,
					TargetPort: intstr.FromInt(spec.DefaultPort),
				},
			},
		},
	}
}

func ingressSpec(namespace, name, vHost string) *k8s_extensions.Ingress {
	return &k8s_extensions.Ingress{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "extensions/v1beta1",
			Kind:       "Ingress",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: k8s_extensions.IngressSpec{
			Rules: []k8s_extensions.IngressRule{
				{
					Host: vHost,
					IngressRuleValue: k8s_extensions.IngressRuleValue{
						HTTP: &k8s_extensions.HTTPIngressRuleValue{
							Paths: []k8s_extensions.HTTPIngressPath{
								{
									Path: "/",
									Backend: k8s_extensions.IngressBackend{
										ServiceName: name,
										ServicePort: intstr.FromInt(defaultServicePort),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
