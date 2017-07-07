package k8s

import (
	"testing"

	"k8s.io/client-go/pkg/util/intstr"

	"github.com/luizalabs/teresa-api/pkg/server/deploy"
)

func TestPodSpecToK8sContainer(t *testing.T) {
	ps := &deploy.PodSpec{
		Name:  "Teresa",
		Image: "luizalabs/teresa:0.0.1",
		Env:   map[string]string{"ENV-KEY": "ENV-VALUE"},
		VolumeMounts: []*deploy.PodVolumeMountsSpec{
			&deploy.PodVolumeMountsSpec{Name: "Vol1", MountPath: "/tmp", ReadOnly: true},
		},
	}
	c := podSpecToK8sContainer(ps)

	if c.Name != ps.Name {
		t.Errorf("expected %s, got %s", ps.Name, c.Name)
	}
	if c.Image != ps.Image {
		t.Errorf("expected %s, got %s", ps.Image, c.Image)
	}

	for _, e := range c.Env {
		if ps.Env[e.Name] != e.Value {
			t.Errorf("expected %s, got %s, for key %s", e.Value, ps.Env[e.Name], e.Name)
		}
	}

	for idx, vm := range ps.VolumeMounts {
		if c.VolumeMounts[idx].Name != vm.Name {
			t.Errorf("expected %s, got %s", vm.Name, c.VolumeMounts[idx].Name)
		}
		if c.VolumeMounts[idx].MountPath != vm.MountPath {
			t.Errorf("expected %s, got %s", vm.MountPath, c.VolumeMounts[idx].MountPath)
		}
		if c.VolumeMounts[idx].ReadOnly != vm.ReadOnly {
			t.Errorf("expected %v, got %v", vm.ReadOnly, c.VolumeMounts[idx].ReadOnly)
		}
	}
}

func TestPodSpecVolumesToK8sVolumes(t *testing.T) {
	vols := []*deploy.PodVolumeSpec{
		&deploy.PodVolumeSpec{Name: "Vol-Test", SecretName: "Bond"},
	}
	k8sVols := podSpecVolumesToK8sVolumes(vols)

	for idx, vol := range vols {
		if k8sVols[idx].Name != vol.Name {
			t.Errorf("expected %s, got %s", vol.Name, k8sVols[idx].Name)
		}
		if k8sVols[idx].Secret.SecretName != vol.SecretName {
			t.Errorf("expected %s, got %s", vol.SecretName, k8sVols[idx].Secret.SecretName)
		}
	}
}

func TestPodSpecToK8sPod(t *testing.T) {
	ps := &deploy.PodSpec{
		Name:  "Teresa",
		Image: "luizalabs/teresa:0.0.1",
		Env:   map[string]string{"ENV-KEY": "ENV-VALUE"},
		VolumeMounts: []*deploy.PodVolumeMountsSpec{
			&deploy.PodVolumeMountsSpec{Name: "Vol1", MountPath: "/tmp", ReadOnly: true},
		},
	}
	pod := podSpecToK8sPod(ps)
	if pod.ObjectMeta.Name != ps.Name {
		t.Errorf("expected %s, got %s", ps.Name, pod.ObjectMeta.Name)
	}
	if pod.ObjectMeta.Namespace != ps.Namespace {
		t.Errorf("expected %s, got %s", ps.Namespace, pod.ObjectMeta.Namespace)
	}
}

func TestRollingUpdateToK8sRollingUpdate(t *testing.T) {
	ru := &deploy.RollingUpdate{MaxSurge: "3", MaxUnavailable: "30%"}
	maxSurge, maxUnavailable := rollingUpdateToK8sRollingUpdate(ru)

	if maxSurge != intstr.FromInt(3) {
		t.Errorf("expected %s, got %v", ru.MaxSurge, maxSurge)
	}

	if maxUnavailable != intstr.FromString(ru.MaxUnavailable) {
		t.Errorf("expected %s, got %v", ru.MaxUnavailable, maxUnavailable)
	}
}

func TestHealthCheckProbeToK8sProbe(t *testing.T) {
	hc := &deploy.HealthCheckProbe{
		FailureThreshold:    2,
		InitialDelaySeconds: 5,
		PeriodSeconds:       5,
		SuccessThreshold:    2,
		TimeoutSeconds:      3,
		Path:                "/hc/",
	}
	k8sHC := healthCheckProbeToK8sProbe(hc)

	if k8sHC.InitialDelaySeconds != hc.InitialDelaySeconds {
		t.Errorf("expected %d, got %d", hc.InitialDelaySeconds, k8sHC.InitialDelaySeconds)
	}
	if k8sHC.FailureThreshold != hc.FailureThreshold {
		t.Errorf("expected %d, got %d", hc.FailureThreshold, k8sHC.FailureThreshold)
	}
	if k8sHC.PeriodSeconds != hc.PeriodSeconds {
		t.Errorf("expected %d, got %d", hc.PeriodSeconds, k8sHC.PeriodSeconds)
	}
	if k8sHC.SuccessThreshold != hc.SuccessThreshold {
		t.Errorf("expected %d, got %d", hc.SuccessThreshold, k8sHC.SuccessThreshold)
	}
	if k8sHC.TimeoutSeconds != hc.TimeoutSeconds {
		t.Errorf("expected %d, got %d", hc.TimeoutSeconds, k8sHC.TimeoutSeconds)
	}
	if k8sHC.Handler.HTTPGet.Path != hc.Path {
		t.Errorf("expected %s, got %s", hc.Path, k8sHC.Handler.HTTPGet.Path)
	}
}

func TestDeploySpecToK8sDeploy(t *testing.T) {
	ds := &deploy.DeploySpec{
		PodSpec: deploy.PodSpec{
			Name:  "Teresa",
			Image: "luizalabs/teresa:0.0.1",
		},
		Args: []string{"run", "web"},
		TeresaYaml: deploy.TeresaYaml{
			HealthCheck: &deploy.HealthCheck{
				Liveness:  &deploy.HealthCheckProbe{PeriodSeconds: 2},
				Readiness: &deploy.HealthCheckProbe{PeriodSeconds: 5},
			},
			RollingUpdate: &deploy.RollingUpdate{MaxSurge: "3", MaxUnavailable: "30%"},
		},
	}

	var expectedReplicas int32 = 5
	k8sDeploy := deploySpecToK8sDeploy(ds, expectedReplicas)

	if len(k8sDeploy.Spec.Template.Spec.Containers) != 1 {
		t.Fatal("expected 1 container, got %d", len(k8sDeploy.Spec.Template.Spec.Containers))
	}
	c := k8sDeploy.Spec.Template.Spec.Containers[0]
	for idx, arg := range ds.Args {
		if c.Args[idx] != arg {
			t.Errorf("expected %s, got %s", arg, c.Args[idx])
		}
	}

	if c.LivenessProbe.PeriodSeconds != ds.HealthCheck.Liveness.PeriodSeconds {
		t.Errorf("expected %d, got %d", ds.HealthCheck.Liveness.PeriodSeconds, c.LivenessProbe.PeriodSeconds)
	}
	if c.ReadinessProbe.PeriodSeconds != ds.HealthCheck.Readiness.PeriodSeconds {
		t.Errorf("expected %d, got %d", ds.HealthCheck.Readiness.PeriodSeconds, c.ReadinessProbe.PeriodSeconds)
	}

	k8sReplicas := k8sDeploy.Spec.Replicas
	if *k8sReplicas != int32(expectedReplicas) {
		t.Errorf("expected %d, got %d", expectedReplicas, *k8sReplicas)
	}

	k8sRollingUpdate := k8sDeploy.Spec.Strategy.RollingUpdate
	if *k8sRollingUpdate.MaxUnavailable != intstr.FromString("30%") {
		t.Errorf("expected 30%%, got %v", *k8sRollingUpdate.MaxUnavailable)
	}
	if *k8sRollingUpdate.MaxSurge != intstr.FromInt(3) {
		t.Errorf("expected 3, got %v", k8sRollingUpdate.MaxSurge)
	}
}
