package k8s

import (
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	k8sv1 "k8s.io/client-go/pkg/api/v1"

	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/spec"
)

func TestPodSpecToK8sContainer(t *testing.T) {
	ps := &spec.Pod{
		Container: spec.Container{
			Name:  "Teresa",
			Image: "luizalabs/teresa:0.0.1",
			Env:   map[string]string{"ENV-KEY": "ENV-VALUE"},
			Args:  []string{"start", "release"},
			VolumeMounts: []*spec.VolumeMounts{
				{Name: "Vol1", MountPath: "/tmp", ReadOnly: true},
			},
			ContainerLimits: &spec.ContainerLimits{
				CPU:    "800m",
				Memory: "1Gi",
			},
		},
	}
	c, err := podSpecToK8sContainer(ps)
	if err != nil {
		t.Fatal("error to convert spec", err)
	}

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

	for idx, arg := range ps.Args {
		if c.Args[idx] != arg {
			t.Errorf("expected %s, got %s", arg, c.Args[idx])
		}
	}

	expectedCPU, err := resource.ParseQuantity(ps.ContainerLimits.CPU)
	if err != nil {
		t.Fatal("error in default cpu limit:", err)
	}
	if c.Resources.Limits[k8sv1.ResourceCPU] != expectedCPU {
		t.Errorf(
			"expected %v, got %v",
			expectedCPU,
			c.Resources.Limits[k8sv1.ResourceCPU],
		)
	}
	expectedMemory, err := resource.ParseQuantity(ps.ContainerLimits.Memory)
	if err != nil {
		t.Fatal("error in default memory limit:", err)
	}
	if c.Resources.Limits[k8sv1.ResourceMemory] != expectedMemory {
		t.Errorf(
			"expected %v, got %v",
			expectedMemory,
			c.Resources.Limits[k8sv1.ResourceMemory],
		)
	}
}

func TestPodSpecVolumesToK8sVolumes(t *testing.T) {
	vols := []*spec.Volume{
		{Name: "Vol-Test", SecretName: "Bond"},
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
	ps := &spec.Pod{
		Container: spec.Container{
			Name:  "Teresa",
			Image: "luizalabs/teresa:0.0.1",
			Env:   map[string]string{"ENV-KEY": "ENV-VALUE"},
			VolumeMounts: []*spec.VolumeMounts{
				{Name: "Vol1", MountPath: "/tmp", ReadOnly: true},
			},
		},
	}
	pod, err := podSpecToK8sPod(ps)
	if err != nil {
		t.Fatal("error to convert spec", err)
	}
	if pod.ObjectMeta.Name != ps.Name {
		t.Errorf("expected %s, got %s", ps.Name, pod.ObjectMeta.Name)
	}
	if pod.ObjectMeta.Namespace != ps.Namespace {
		t.Errorf("expected %s, got %s", ps.Namespace, pod.ObjectMeta.Namespace)
	}
}

func TestRollingUpdateToK8sRollingUpdate(t *testing.T) {
	ru := &spec.RollingUpdate{MaxSurge: "3", MaxUnavailable: "30%"}
	maxSurge, maxUnavailable := rollingUpdateToK8sRollingUpdate(ru)

	if maxSurge != intstr.FromInt(3) {
		t.Errorf("expected %s, got %v", ru.MaxSurge, maxSurge)
	}

	if maxUnavailable != intstr.FromString(ru.MaxUnavailable) {
		t.Errorf("expected %s, got %v", ru.MaxUnavailable, maxUnavailable)
	}
}

func TestHealthCheckProbeToK8sProbe(t *testing.T) {
	hc := &spec.HealthCheckProbe{
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
	ds := &spec.Deploy{
		Pod: spec.Pod{
			Container: spec.Container{
				Name:  "Teresa",
				Image: "luizalabs/teresa:0.0.1",
				Args:  []string{"run", "web"},
			},
		},
		TeresaYaml: spec.TeresaYaml{
			HealthCheck: &spec.HealthCheck{
				Liveness:  &spec.HealthCheckProbe{PeriodSeconds: 2},
				Readiness: &spec.HealthCheckProbe{PeriodSeconds: 5},
			},
			RollingUpdate: &spec.RollingUpdate{MaxSurge: "3", MaxUnavailable: "30%"},
		},
	}

	var expectedReplicas int32 = 5
	k8sDeploy, err := deploySpecToK8sDeploy(ds, expectedReplicas)
	if err != nil {
		t.Fatalf("error to convert specL %v", err)
	}
	if len(k8sDeploy.Spec.Template.Spec.Containers) != 1 {
		t.Fatalf("expected 1 container, got %d", len(k8sDeploy.Spec.Template.Spec.Containers))
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

func TestServiceSpec(t *testing.T) {
	name := "teresa"
	namespace := "teresa"
	srvType := "LoadBalancer"

	s := serviceSpec(namespace, name, srvType)
	if s.ObjectMeta.Name != name {
		t.Errorf("expected %s, got %s", name, s.ObjectMeta.Name)
	}
	if s.ObjectMeta.Namespace != namespace {
		t.Errorf("expected %s, got %s", namespace, s.ObjectMeta.Namespace)
	}
	if s.Spec.Type != k8sv1.ServiceType(srvType) {
		t.Errorf("expected %s, got %v", srvType, s.Spec.Type)
	}
}

func TestIngressSpec(t *testing.T) {
	name := "teresa"
	namespace := "teresa"
	vHost := "test.teresa-apps.io"

	i := ingressSpec(namespace, name, vHost)
	if i.ObjectMeta.Name != name {
		t.Errorf("expected %s, got %s", name, i.ObjectMeta.Name)
	}
	if i.ObjectMeta.Namespace != namespace {
		t.Errorf("expected %s, got %s", namespace, i.ObjectMeta.Namespace)
	}
	if i.Spec.Rules[0].Host != vHost {
		t.Errorf("expected %s, got %s", vHost, i.Spec.Rules[0].Host)
	}
}

func TestPodSpecToK8sPodShouldAddAutomountSATokenField(t *testing.T) {
	ps := &spec.Pod{
		Container: spec.Container{
			Name:  "Teresa",
			Image: "luizalabs/teresa:0.0.1",
		},
	}

	pod, err := podSpecToK8sPod(ps)
	if err != nil {
		t.Fatal("error converting spec:", err)
	}
	if pod.Spec.AutomountServiceAccountToken == nil {
		t.Fatal("got nil AutomountServiceAccountToken")
	}

	if *pod.Spec.AutomountServiceAccountToken {
		t.Error("got true, expected false")
	}
}

func TestDeploySpecToK8sDeployShouldAddAutomountSATokenField(t *testing.T) {
	ds := &spec.Deploy{
		Pod: spec.Pod{
			Container: spec.Container{
				Name:  "Teresa",
				Image: "luizalabs/teresa:0.0.1",
				Args:  []string{"run", "web"},
			},
		},
	}

	k8sDeploy, err := deploySpecToK8sDeploy(ds, 1)
	if err != nil {
		t.Fatal("error converting spec:", err)
	}

	ps := k8sDeploy.Spec.Template.Spec
	if ps.AutomountServiceAccountToken == nil {
		t.Fatal("got nil AutomountServiceAccountToken")
	}

	if *ps.AutomountServiceAccountToken {
		t.Error("got true, expected false")
	}
}

func TestAppPodListOptsToK8s(t *testing.T) {
	opts := &app.PodListOptions{PodName: "test-1234"}
	expectedFs := "metadata.name=test-1234"

	k8sOpts := appPodListOptsToK8s(opts)

	if k8sOpts.FieldSelector != expectedFs {
		t.Errorf("got %s, want %s", k8sOpts.FieldSelector, expectedFs)
	}
}
