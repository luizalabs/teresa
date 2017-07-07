package k8s

import (
	"strconv"

	"github.com/luizalabs/teresa-api/pkg/server/deploy"
	"k8s.io/client-go/pkg/api/unversioned"
	k8sv1 "k8s.io/client-go/pkg/api/v1"
	k8s_extensions "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/util/intstr"
)

func podSpecToK8sContainer(podSpec *deploy.PodSpec) k8sv1.Container {
	c := k8sv1.Container{
		Name:            podSpec.Name,
		ImagePullPolicy: k8sv1.PullIfNotPresent,
		Image:           podSpec.Image,
	}
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
	return c
}

func podSpecVolumesToK8sVolumes(vols []*deploy.PodVolumeSpec) []k8sv1.Volume {
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

func podSpecToK8sPod(podSpec *deploy.PodSpec) *k8sv1.Pod {
	c := podSpecToK8sContainer(podSpec)
	volumes := podSpecVolumesToK8sVolumes(podSpec.Volume)

	ps := k8sv1.PodSpec{
		RestartPolicy: k8sv1.RestartPolicyNever,
		Containers:    []k8sv1.Container{c},
		Volumes:       volumes,
	}

	pod := &k8sv1.Pod{
		TypeMeta: unversioned.TypeMeta{Kind: "Pod", APIVersion: "v1"},
		ObjectMeta: k8sv1.ObjectMeta{
			Name:      podSpec.Name,
			Namespace: podSpec.Namespace,
		},
		Spec: ps,
	}
	return pod
}

func deploySpecToK8sDeploy(deploySpec *deploy.DeploySpec, replicas int32) *k8s_extensions.Deployment {
	c := podSpecToK8sContainer(&deploySpec.PodSpec)
	volumes := podSpecVolumesToK8sVolumes(deploySpec.Volume)

	for _, arg := range deploySpec.Args {
		c.Args = append(c.Args, arg)
	}

	if deploySpec.HealthCheck != nil {
		if deploySpec.HealthCheck.Liveness != nil {
			c.LivenessProbe = healthCheckProbeToK8sProbe(deploySpec.HealthCheck.Liveness)
		}
		if deploySpec.HealthCheck.Readiness != nil {
			c.ReadinessProbe = healthCheckProbeToK8sProbe(deploySpec.HealthCheck.Readiness)
		}
	}

	ps := k8sv1.PodSpec{
		RestartPolicy: k8sv1.RestartPolicyAlways,
		Containers:    []k8sv1.Container{c},
		Volumes:       volumes,
	}

	var maxSurge, maxUnavailable *intstr.IntOrString
	if deploySpec.RollingUpdate != nil {
		vMaxSurge, vMaxUnavailable := rollingUpdateToK8sRollingUpdate(deploySpec.RollingUpdate)
		maxSurge, maxUnavailable = &vMaxSurge, &vMaxUnavailable
	}

	rhl := int32(deploySpec.RevisionHistoryLimit)
	d := &k8s_extensions.Deployment{
		TypeMeta: unversioned.TypeMeta{
			APIVersion: "extensions/v1beta1",
			Kind:       "Deployment",
		},
		ObjectMeta: k8sv1.ObjectMeta{
			Name:      deploySpec.Name,
			Namespace: deploySpec.Namespace,
			Labels:    map[string]string{"run": deploySpec.Name},
			Annotations: map[string]string{
				"kubernetes.io/change-cause": deploySpec.Description,
				"teresa.io/slug":             deploySpec.SlugURL,
			},
		},
		Spec: k8s_extensions.DeploymentSpec{
			Replicas: &replicas,
			Strategy: k8s_extensions.DeploymentStrategy{
				Type: k8s_extensions.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &k8s_extensions.RollingUpdateDeployment{
					MaxUnavailable: maxUnavailable,
					MaxSurge:       maxSurge,
				},
			},
			Template: k8sv1.PodTemplateSpec{
				ObjectMeta: k8sv1.ObjectMeta{
					Labels: map[string]string{"run": deploySpec.Name},
				},
				Spec: ps,
			},
			RevisionHistoryLimit: &rhl,
		},
	}
	return d
}

func rollingUpdateToK8sRollingUpdate(ru *deploy.RollingUpdate) (maxSurge, maxUnavailable intstr.IntOrString) {
	conv := func(value string) intstr.IntOrString {
		v, err := strconv.Atoi(value)
		if err != nil {
			return intstr.FromString(value)
		}
		return intstr.FromInt(v)
	}
	return conv(ru.MaxSurge), conv(ru.MaxUnavailable)
}

func healthCheckProbeToK8sProbe(probe *deploy.HealthCheckProbe) *k8sv1.Probe {
	return &k8sv1.Probe{
		InitialDelaySeconds: probe.InitialDelaySeconds,
		TimeoutSeconds:      probe.TimeoutSeconds,
		PeriodSeconds:       probe.PeriodSeconds,
		FailureThreshold:    probe.FailureThreshold,
		SuccessThreshold:    probe.SuccessThreshold,
		Handler: k8sv1.Handler{
			HTTPGet: &k8sv1.HTTPGetAction{
				Port: intstr.FromInt(deploy.DefaultPort),
				Path: probe.Path,
			},
		},
	}
}
