package k8s

import (
	"fmt"
	"strconv"

	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/spec"
	"k8s.io/api/apps/v1beta2"
	k8sbatch "k8s.io/api/batch/v1"
	k8sv1beta1 "k8s.io/api/batch/v1beta1"
	k8sv1 "k8s.io/api/core/v1"
	k8s_extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	changeCauseAnnotation = "kubernetes.io/change-cause"
	appTypeAnnotation     = "teresa.io/app-type"
)

func podSpecToK8sContainers(podSpec *spec.Pod) ([]k8sv1.Container, error) {
	return containerSpecsToK8sContainers(podSpec.Containers)
}

func containerSpecsToK8sContainers(containerSpecs []*spec.Container) ([]k8sv1.Container, error) {
	containers := make([]k8sv1.Container, len(containerSpecs))
	for i, cs := range containerSpecs {
		c := k8sv1.Container{
			Name:            cs.Name,
			ImagePullPolicy: k8sv1.PullAlways,
			Image:           cs.Image,
		}

		if cs.ContainerLimits != nil {
			cpu, err := resource.ParseQuantity(cs.ContainerLimits.CPU)
			if err != nil {
				return nil, err
			}
			memory, err := resource.ParseQuantity(cs.ContainerLimits.Memory)
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

		if len(cs.Command) > 0 {
			c.Command = cs.Command
		}
		c.Args = append(c.Args, cs.Args...)

		for k, v := range cs.Env {
			c.Env = append(c.Env, k8sv1.EnvVar{Name: k, Value: v})
		}
		for _, secret := range cs.Secrets {
			c.Env = append(c.Env, k8sv1.EnvVar{
				Name: secret,
				ValueFrom: &k8sv1.EnvVarSource{
					SecretKeyRef: &k8sv1.SecretKeySelector{
						Key: secret,
						LocalObjectReference: k8sv1.LocalObjectReference{
							Name: app.TeresaAppSecrets,
						},
					},
				},
			})
		}
		for _, vm := range cs.VolumeMounts {
			c.VolumeMounts = append(c.VolumeMounts, k8sv1.VolumeMount{
				Name:      vm.Name,
				MountPath: vm.MountPath,
				ReadOnly:  vm.ReadOnly,
			})
		}
		for _, p := range cs.Ports {
			c.Ports = append(c.Ports, k8sv1.ContainerPort{
				Name:          p.Name,
				ContainerPort: p.ContainerPort,
			})
		}
		containers[i] = c
	}
	return containers, nil
}

func podSpecVolumesToK8sVolumes(vols []*spec.Volume) []k8sv1.Volume {
	volumes := make([]k8sv1.Volume, 0)
	for _, v := range vols {
		vol := k8sv1.Volume{Name: v.Name}
		if v.EmptyDir {
			vol.EmptyDir = &k8sv1.EmptyDirVolumeSource{}
		} else if v.SecretName != "" {
			vol.Secret = &k8sv1.SecretVolumeSource{
				SecretName: v.SecretName,
			}
		} else if v.ConfigMapName != "" {
			vol.ConfigMap = &k8sv1.ConfigMapVolumeSource{}
			vol.ConfigMap.Name = v.ConfigMapName
		}
		volumes = append(volumes, vol)
	}
	return volumes
}

func podSpecToK8sPod(podSpec *spec.Pod) (*k8sv1.Pod, error) {
	containers, err := podSpecToK8sContainers(podSpec)
	if err != nil {
		return nil, err
	}
	volumes := podSpecVolumesToK8sVolumes(podSpec.Volumes)
	f := false

	initContainers, err := podSpecToK8sInitContainers(podSpec)
	if err != nil {
		return nil, err
	}

	ps := k8sv1.PodSpec{
		RestartPolicy: k8sv1.RestartPolicyNever,
		Containers:    containers,
		Volumes:       volumes,
		AutomountServiceAccountToken: &f,
		InitContainers:               initContainers,
	}

	pod := &k8sv1.Pod{
		TypeMeta: metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      podSpec.Name,
			Namespace: podSpec.Namespace,
			Labels:    podSpec.Labels,
		},
		Spec: ps,
	}
	return pod, nil
}

func deploySpecToK8sDeploy(deploySpec *spec.Deploy, replicas int32) (*v1beta2.Deployment, error) {
	containers, err := podSpecToK8sContainers(&deploySpec.Pod)
	if err != nil {
		return nil, err
	}
	volumes := podSpecVolumesToK8sVolumes(deploySpec.Volumes)

	if deploySpec.HealthCheck != nil {
		if deploySpec.HealthCheck.Liveness != nil {
			containers[0].LivenessProbe = healthCheckProbeToK8sProbe(
				deploySpec.HealthCheck.Liveness,
				containers[0].Ports[0].ContainerPort,
			)
		}
		if deploySpec.HealthCheck.Readiness != nil {
			containers[0].ReadinessProbe = healthCheckProbeToK8sProbe(
				deploySpec.HealthCheck.Readiness,
				containers[0].Ports[0].ContainerPort,
			)
		}
	}

	if deploySpec.Lifecycle != nil {
		containers[0].Lifecycle = lifecycleToK8sLifecycle(deploySpec.Lifecycle)
	}

	f := false
	initContainers, err := podSpecToK8sInitContainers(&deploySpec.Pod)
	if err != nil {
		return nil, err
	}
	ps := k8sv1.PodSpec{
		RestartPolicy: k8sv1.RestartPolicyAlways,
		Containers:    containers,
		Volumes:       volumes,
		AutomountServiceAccountToken: &f,
		InitContainers:               initContainers,
	}

	var maxSurge, maxUnavailable *intstr.IntOrString
	if deploySpec.RollingUpdate != nil {
		vMaxSurge, vMaxUnavailable := rollingUpdateToK8sRollingUpdate(deploySpec.RollingUpdate)
		maxSurge, maxUnavailable = &vMaxSurge, &vMaxUnavailable
	}

	rhl := int32(deploySpec.RevisionHistoryLimit)
	d := &v1beta2.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "extensions/v1beta2",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploySpec.Name,
			Namespace: deploySpec.Namespace,
			Labels:    deploySpec.Labels,
			Annotations: map[string]string{
				changeCauseAnnotation: deploySpec.Description,
				spec.SlugAnnotation:   deploySpec.SlugURL,
			},
		},
		Spec: v1beta2.DeploymentSpec{
			Replicas: &replicas,
			Strategy: v1beta2.DeploymentStrategy{
				Type: v1beta2.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &v1beta2.RollingUpdateDeployment{
					MaxUnavailable: maxUnavailable,
					MaxSurge:       maxSurge,
				},
			},
			Template: k8sv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: deploySpec.Labels,
				},
				Spec: ps,
			},
			RevisionHistoryLimit: &rhl,
			Selector: &metav1.LabelSelector{
				MatchLabels: deploySpec.MatchLabels,
			},
		},
	}
	return d, nil
}

func podSpecToK8sInitContainers(podSpec *spec.Pod) ([]k8sv1.Container, error) {
	return containerSpecsToK8sContainers(podSpec.InitContainers)
}

func cronJobSpecToK8sCronJob(cronJobSpec *spec.CronJob) (*k8sv1beta1.CronJob, error) {
	containers, err := podSpecToK8sContainers(&cronJobSpec.Pod)
	if err != nil {
		return nil, err
	}
	volumes := podSpecVolumesToK8sVolumes(cronJobSpec.Volumes)

	initContainers, err := podSpecToK8sInitContainers(&cronJobSpec.Pod)
	if err != nil {
		return nil, err
	}

	f := false
	ps := k8sv1.PodSpec{
		RestartPolicy: k8sv1.RestartPolicyNever,
		Containers:    containers,
		Volumes:       volumes,
		AutomountServiceAccountToken: &f,
		InitContainers:               initContainers,
	}

	successfulLim := cronJobSpec.SuccessfulJobsHistoryLimit
	failedLim := cronJobSpec.FailedJobsHistoryLimit
	cj := &k8sv1beta1.CronJob{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1beta1",
			Kind:       "CronJob",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cronJobSpec.Name,
			Namespace: cronJobSpec.Namespace,
			Annotations: map[string]string{
				changeCauseAnnotation: cronJobSpec.Description,
				spec.SlugAnnotation:   cronJobSpec.SlugURL,
				appTypeAnnotation:     "cronjob",
			},
		},
		Spec: k8sv1beta1.CronJobSpec{
			Schedule: cronJobSpec.Schedule,
			JobTemplate: k8sv1beta1.JobTemplateSpec{
				Spec: k8sbatch.JobSpec{
					Template: k8sv1.PodTemplateSpec{
						Spec: ps,
					},
				},
			},
			SuccessfulJobsHistoryLimit: &successfulLim,
			FailedJobsHistoryLimit:     &failedLim,
		},
	}
	return cj, nil
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

func healthCheckProbeToK8sProbe(probe *spec.HealthCheckProbe, port int32) *k8sv1.Probe {
	return &k8sv1.Probe{
		InitialDelaySeconds: probe.InitialDelaySeconds,
		TimeoutSeconds:      probe.TimeoutSeconds,
		PeriodSeconds:       probe.PeriodSeconds,
		FailureThreshold:    probe.FailureThreshold,
		SuccessThreshold:    probe.SuccessThreshold,
		Handler: k8sv1.Handler{
			HTTPGet: &k8sv1.HTTPGetAction{
				Port: intstr.FromInt(int(port)),
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

func serviceSpecToK8s(svcSpec *spec.Service) *k8sv1.Service {
	serviceType := k8sv1.ServiceType(svcSpec.Type)
	return &k8sv1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels:    svcSpec.Labels,
			Name:      svcSpec.Name,
			Namespace: svcSpec.Namespace,
		},
		Spec: k8sv1.ServiceSpec{
			Type:            serviceType,
			SessionAffinity: k8sv1.ServiceAffinityNone,
			Selector:        svcSpec.Labels,
			Ports:           servicePortsToK8sServicePorts(svcSpec.Ports),
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
										ServicePort: intstr.FromInt(spec.DefaultExternalPort),
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

func appPodListOptsToK8s(opts *app.PodListOptions) *metav1.ListOptions {
	var k8sOpts metav1.ListOptions

	if opts.PodName != "" {
		k8sOpts.FieldSelector = fmt.Sprintf("metadata.name=%s", opts.PodName)
	}

	return &k8sOpts
}

func configMapSpec(namespace, name string, data map[string]string) *k8sv1.ConfigMap {
	return &k8sv1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}
}

func servicePortsToK8sServicePorts(ports []spec.ServicePort) []k8sv1.ServicePort {
	k8sPorts := make([]k8sv1.ServicePort, len(ports))
	for i := range ports {
		k8sPorts[i] = k8sv1.ServicePort{
			Name:       ports[i].Name,
			Port:       int32(ports[i].Port),
			TargetPort: intstr.FromInt(ports[i].TargetPort),
		}
	}
	return k8sPorts
}

func k8sExplicitEnvToAppEnv(env []k8sv1.EnvVar) []*app.EnvVar {
	evs := []*app.EnvVar{}
	for _, e := range env {
		if e.ValueFrom != nil {
			continue
		}
		evs = append(evs, &app.EnvVar{
			Key:   e.Name,
			Value: e.Value,
		})
	}
	return evs
}
