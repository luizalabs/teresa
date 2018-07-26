package k8s

import (
	"reflect"
	"testing"

	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/spec"
)

func TestPodSpecToK8sContainers(t *testing.T) {
	ps := &spec.Pod{
		Containers: []*spec.Container{{
			Name:    "Teresa",
			Image:   "luizalabs/teresa:0.0.1",
			Env:     map[string]string{"ENV-KEY": "ENV-VALUE"},
			Args:    []string{"start", "release"},
			Secrets: []string{"SECRET-1", "SECRET-2"},
			VolumeMounts: []*spec.VolumeMounts{
				{Name: "Vol1", MountPath: "/tmp", ReadOnly: true},
			},
			ContainerLimits: &spec.ContainerLimits{
				CPU:    "800m",
				Memory: "1Gi",
			},
		}},
	}
	containers, err := podSpecToK8sContainers(ps)
	c := containers[0]
	if err != nil {
		t.Fatal("error to convert spec", err)
	}

	if c.Name != ps.Containers[0].Name {
		t.Errorf("expected %s, got %s", ps.Containers[0].Name, c.Name)
	}
	if c.Image != ps.Containers[0].Image {
		t.Errorf("expected %s, got %s", ps.Containers[0].Image, c.Image)
	}

	for _, e := range c.Env {
		if ps.Containers[0].Env[e.Name] != e.Value {
			t.Errorf("expected %s, got %s, for key %s", e.Value, ps.Containers[0].Env[e.Name], e.Name)
		}
	}

	for _, secret := range ps.Containers[0].Secrets {
		found := false
		for _, e := range c.Env {
			found = e.Name == secret
			if found {
				if e.ValueFrom.SecretKeyRef.Key != secret {
					t.Errorf("expected an env with secret key ref for secret %s", secret)
				}
				if e.ValueFrom.SecretKeyRef.Name != app.TeresaAppSecrets {
					t.Errorf("expected an env with secret key ref with name %s", app.TeresaAppSecrets)
				}
				break
			}
		}
		if !found {
			t.Errorf("expected env with secret for secret %s", secret)
		}
	}

	for idx, vm := range ps.Containers[0].VolumeMounts {
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

	for idx, arg := range ps.Containers[0].Args {
		if c.Args[idx] != arg {
			t.Errorf("expected %s, got %s", arg, c.Args[idx])
		}
	}

	expectedCPU, err := resource.ParseQuantity(ps.Containers[0].ContainerLimits.CPU)
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
	expectedMemory, err := resource.ParseQuantity(ps.Containers[0].ContainerLimits.Memory)
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

func TestPodSpecSecretVolumesToK8s(t *testing.T) {
	var testCases = []struct {
		vols []*spec.Volume
	}{
		{vols: []*spec.Volume{{Name: "vTest", SecretName: "sn"}}},
		{vols: []*spec.Volume{{Name: "vTest2", SecretName: "sn2", Items: []spec.VolumeItem{{Key: "k", Path: "p"}}}}},
	}

	for _, tc := range testCases {
		k8sVols := podSpecVolumesToK8sVolumes(tc.vols)

		for idx, vol := range tc.vols {
			if k8sVols[idx].Name != vol.Name {
				t.Errorf("expected %s, got %s", vol.Name, k8sVols[idx].Name)
			}
			if k8sVols[idx].Secret.SecretName != vol.SecretName {
				t.Errorf("expected %s, got %s", vol.SecretName, k8sVols[idx].Secret.SecretName)
			}
			for i := range vol.Items {
				if actual := k8sVols[idx].Secret.Items[i].Key; actual != vol.Items[i].Key {
					t.Errorf("expected %s, got %s", vol.Items[i].Key, actual)
				}
			}
		}
	}
}

func TestPodSpecEmptyDirVolumeToK8s(t *testing.T) {
	vols := []*spec.Volume{
		{Name: "Vol-Test", EmptyDir: true},
	}
	k8sVols := podSpecVolumesToK8sVolumes(vols)

	for idx, vol := range vols {
		if k8sVols[idx].Name != vol.Name {
			t.Errorf("expected %s, got %s", vol.Name, k8sVols[idx].Name)
		}
		if k8sVols[idx].Secret != nil {
			t.Errorf("expected nil, got %v", k8sVols[idx].Secret)
		}
		if k8sVols[idx].ConfigMap != nil {
			t.Errorf("expected nil, got %v", k8sVols[idx].ConfigMap)
		}
		if k8sVols[idx].EmptyDir == nil {
			t.Error("expected pointer to struct, got nil")
		}
	}
}

func TestPodSpecConfigMapVolumeToK8s(t *testing.T) {
	vols := []*spec.Volume{
		{Name: "Vol-Test", ConfigMapName: "test"},
	}
	k8sVols := podSpecVolumesToK8sVolumes(vols)

	for idx, vol := range vols {
		if k8sVols[idx].Name != vol.Name {
			t.Errorf("expected %s, got %s", vol.Name, k8sVols[idx].Name)
		}
		if k8sVols[idx].Secret != nil {
			t.Errorf("expected nil, got %v", k8sVols[idx].Secret)
		}
		if k8sVols[idx].EmptyDir != nil {
			t.Errorf("expected nil, got %v", k8sVols[idx].EmptyDir)
		}
		if k8sVols[idx].ConfigMap == nil {
			t.Error("expected pointer to struct, got nil")
		}
	}
}

func TestPodSpecToK8sPod(t *testing.T) {
	ps := &spec.Pod{
		Containers: []*spec.Container{{
			Name:  "Teresa",
			Image: "luizalabs/teresa:0.0.1",
			Env:   map[string]string{"ENV-KEY": "ENV-VALUE"},
			VolumeMounts: []*spec.VolumeMounts{
				{Name: "Vol1", MountPath: "/tmp", ReadOnly: true},
			}},
		},
		InitContainers: []*spec.Container{{
			Name:  "Teresa",
			Image: "luizalabs/teresa:0.0.1",
		}},
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

	if len(pod.Spec.InitContainers) != len(ps.InitContainers) {
		t.Errorf("expected %d, got %d", len(ps.InitContainers), len(pod.Spec.InitContainers))
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
			Containers: []*spec.Container{{
				Name:  "Teresa",
				Image: "luizalabs/teresa:0.0.1",
				Args:  []string{"run", "web"},
				Ports: []spec.Port{{
					ContainerPort: 5000,
				}},
			}},
			InitContainers: []*spec.Container{{
				Name:  "Teresa",
				Image: "luizalabs/teresa:0.0.1",
			}},
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
		t.Fatalf("error to convert spec %v", err)
	}
	if len(k8sDeploy.Spec.Template.Spec.Containers) != 1 {
		t.Fatalf("expected 1 container, got %d", len(k8sDeploy.Spec.Template.Spec.Containers))
	}
	c := k8sDeploy.Spec.Template.Spec.Containers[0]
	for idx, arg := range ds.Containers[0].Args {
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

	initContainers := k8sDeploy.Spec.Template.Spec.InitContainers
	if len(initContainers) != len(ds.Pod.InitContainers) {
		t.Errorf("expected %d, got %d", len(ds.Pod.InitContainers), len(initContainers))
	}
}

func TestServiceSpec(t *testing.T) {
	name := "teresa"
	srvType := "LoadBalancer"

	s := serviceSpecToK8s(spec.NewDefaultService(name, srvType, "protocol"))
	if s.ObjectMeta.Name != name {
		t.Errorf("expected %s, got %s", name, s.ObjectMeta.Name)
	}
	if s.ObjectMeta.Namespace != name {
		t.Errorf("expected %s, got %s", name, s.ObjectMeta.Namespace)
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
		Containers: []*spec.Container{{
			Name:  "Teresa",
			Image: "luizalabs/teresa:0.0.1",
		}},
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
			Containers: []*spec.Container{{
				Name:  "Teresa",
				Image: "luizalabs/teresa:0.0.1",
				Args:  []string{"run", "web"},
			}},
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

func TestPodSpecToK8sInitContainers(t *testing.T) {
	ps := &spec.Pod{
		InitContainers: []*spec.Container{
			{
				Name: "name1",
			},
			{
				Name: "name2",
			},
		},
	}

	c, err := podSpecToK8sInitContainers(ps)
	if err != nil {
		t.Fatal(err)
	}

	for i, _ := range c {
		if c[i].Name != ps.InitContainers[i].Name {
			t.Errorf("got %s, want %s", c[i].Name, ps.InitContainers[i].Name)
		}
	}
}

func TestCronJobSpecToK8sCronJob(t *testing.T) {
	cs := &spec.CronJob{
		Deploy: spec.Deploy{
			Pod: spec.Pod{
				Containers: []*spec.Container{{
					Name:  "Teresa",
					Image: "luizalabs/teresa:0.0.1",
					Args:  []string{"echo", "hello"},
				}},
				InitContainers: []*spec.Container{{
					Name:  "Teresa",
					Image: "luizalabs/init-teresa:0.0.1",
				}},
			},
		},
		Schedule:                   "*/1 * * * *",
		SuccessfulJobsHistoryLimit: 42,
		FailedJobsHistoryLimit:     33,
	}

	k8sCron, err := cronJobSpecToK8sCronJob(cs)
	if err != nil {
		t.Fatalf("error to convert spec %v", err)
	}

	actualSchedule := k8sCron.Spec.Schedule
	if actualSchedule != cs.Schedule {
		t.Errorf("expected %s, got %s", cs.Schedule, actualSchedule)
	}

	if len(k8sCron.Spec.JobTemplate.Spec.Template.Spec.Containers) != 1 {
		t.Fatalf(
			"expected 1 container, got %d",
			len(k8sCron.Spec.JobTemplate.Spec.Template.Spec.Containers),
		)
	}
	c := k8sCron.Spec.JobTemplate.Spec.Template.Spec.Containers[0]
	for idx, arg := range cs.Containers[0].Args {
		if c.Args[idx] != arg {
			t.Errorf("expected %s, got %s", arg, c.Args[idx])
		}
	}

	initContainers := k8sCron.Spec.JobTemplate.Spec.Template.Spec.InitContainers
	if len(initContainers) != len(cs.Pod.InitContainers) {
		t.Errorf("expected %d, got %d", len(cs.Pod.InitContainers), len(initContainers))
	}

	lim := k8sCron.Spec.SuccessfulJobsHistoryLimit
	if *lim != cs.SuccessfulJobsHistoryLimit {
		t.Errorf("expected %d, got %d", cs.SuccessfulJobsHistoryLimit, *lim)
	}

	lim = k8sCron.Spec.FailedJobsHistoryLimit
	if *lim != cs.FailedJobsHistoryLimit {
		t.Errorf("expected %d, got %d", cs.FailedJobsHistoryLimit, *lim)
	}

	if actual := *k8sCron.Spec.JobTemplate.Spec.BackoffLimit; actual != int32(3) {
		t.Errorf("expected 3, got %d", actual)
	}
}

func TestConfigMapSpec(t *testing.T) {
	name := "teresa"
	namespace := "teresa"
	data := map[string]string{"foo": "bar"}

	o := configMapSpec(namespace, name, data)

	if o.ObjectMeta.Name != name {
		t.Errorf("expected %s, got %s", name, o.ObjectMeta.Name)
	}
	if o.ObjectMeta.Namespace != namespace {
		t.Errorf("expected %s, got %s", namespace, o.ObjectMeta.Namespace)
	}
	if o.Data["foo"] != data["foo"] {
		t.Errorf("expected %s, got %s", data["foo"], o.Data["foo"])
	}
}

func TestServicePortsToK8sServicePorts(t *testing.T) {
	ports := []spec.ServicePort{
		{Name: "port1", TargetPort: 1},
		{Name: "port2", Port: 2, TargetPort: 2},
	}

	k8sPorts := servicePortsToK8sServicePorts(ports)

	if len(k8sPorts) != len(ports) {
		t.Errorf("got %d; want %d", len(k8sPorts), len(ports))
	}
	for i := range ports {
		if ports[i].Name != k8sPorts[i].Name {
			t.Errorf("got %s; want %s", k8sPorts[i].Name, ports[i].Name)
		}
		tp := intstr.FromInt(ports[i].TargetPort)
		if tp != k8sPorts[i].TargetPort {
			t.Errorf("got %v; want %v", k8sPorts[i].TargetPort, tp)
		}
		if int32(ports[i].Port) != k8sPorts[i].Port {
			t.Errorf("got %d; want %d", k8sPorts[i].Port, ports[i].Port)
		}
	}
}

func TestK8sExplicitEnvToAppEnv(t *testing.T) {
	env := []k8sv1.EnvVar{
		{Name: "name1", Value: "value1"},
		{Name: "name2", Value: "value2", ValueFrom: &k8sv1.EnvVarSource{}},
	}
	want := []*app.EnvVar{{Key: "name1", Value: "value1"}}

	got := k8sExplicitEnvToAppEnv(env)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v; want %v", got, want)
	}
}

func TestK8sServiceToService(t *testing.T) {
	name := "test"
	namespace := "namespace"
	svcType := "LoadBalancer"
	labels := map[string]string{"key": "value"}
	ranges := []string{"range1", "range2"}
	ports := []spec.ServicePort{{}}
	k8sSvc := &k8sv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels:    labels,
			Name:      name,
			Namespace: namespace,
		},
		Spec: k8sv1.ServiceSpec{
			Type: k8sv1.ServiceType(svcType),
			LoadBalancerSourceRanges: ranges,
			Ports: []k8sv1.ServicePort{{}},
		},
	}
	svc := k8sServiceToService(k8sSvc)

	if svc.Name != name {
		t.Errorf("want %s; got %s", name, svc.Name)
	}
	if svc.Namespace != namespace {
		t.Errorf("want %s; got %s", namespace, svc.Namespace)
	}
	if svc.Type != svcType {
		t.Errorf("want %s; got %s", svcType, svc.Type)
	}
	if !reflect.DeepEqual(svc.Labels, labels) {
		t.Errorf("want %v; got %v", labels, svc.Labels)
	}
	if !reflect.DeepEqual(svc.SourceRanges, ranges) {
		t.Errorf("want %v; got %v", ranges, svc.SourceRanges)
	}
	if !reflect.DeepEqual(svc.Ports, ports) {
		t.Errorf("want %v; got %v", ports, svc.Ports)
	}
}

func TestK8sServicePortsToServicePorts(t *testing.T) {
	want := []spec.ServicePort{
		{
			Name:       "port1",
			Port:       1,
			TargetPort: 5000,
		},
		{
			Name:       "port2",
			Port:       2,
			TargetPort: 6000,
		},
	}
	k8sPorts := []k8sv1.ServicePort{
		{
			Name:       want[0].Name,
			Port:       int32(want[0].Port),
			TargetPort: intstr.FromInt(want[0].TargetPort),
		},
		{
			Name:       want[1].Name,
			Port:       int32(want[1].Port),
			TargetPort: intstr.FromInt(want[1].TargetPort),
		},
	}
	ports := k8sServicePortsToServicePorts(k8sPorts)

	if !reflect.DeepEqual(ports, want) {
		t.Errorf("want %v; got %v", want, ports)
	}
}

func TestK8sPodToAppPodRunningAndReady(t *testing.T) {
	k8sPod := &k8sv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "pod",
		},
		Status: k8sv1.PodStatus{
			Phase: k8sv1.PodRunning,
			ContainerStatuses: []k8sv1.ContainerStatus{
				{
					Ready:        true,
					RestartCount: 1,
					State: k8sv1.ContainerState{
						Running: &k8sv1.ContainerStateRunning{},
					},
				},
				{
					Ready:        true,
					RestartCount: 1,
					State: k8sv1.ContainerState{
						Running: &k8sv1.ContainerStateRunning{},
					},
				},
			},
		},
	}
	want := &app.Pod{
		Name:     "pod",
		Restarts: 2,
		Ready:    true,
		State:    "Running",
	}
	pod := k8sPodToAppPod(k8sPod)

	if !reflect.DeepEqual(pod, want) {
		t.Errorf("want %v; got %v", want, pod)
	}
}

func TestK8sPodToAppPodTerminating(t *testing.T) {
	k8sPod := &k8sv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "pod",
			DeletionTimestamp: &metav1.Time{},
		},
		Status: k8sv1.PodStatus{
			Phase: k8sv1.PodRunning,
			ContainerStatuses: []k8sv1.ContainerStatus{
				{
					Ready:        true,
					RestartCount: 1,
					State: k8sv1.ContainerState{
						Running: &k8sv1.ContainerStateRunning{},
					},
				},
			},
		},
	}
	want := &app.Pod{
		Name:     "pod",
		Restarts: 1,
		Ready:    true,
		State:    "Terminating",
	}
	pod := k8sPodToAppPod(k8sPod)

	if !reflect.DeepEqual(pod, want) {
		t.Errorf("want %v; got %v", want, pod)
	}
}

func TestDeploySpecPodTemplateAnnotations(t *testing.T) {
	ds := &spec.Deploy{}
	want := map[string]string{
		clusterAutoscalerAnnotation: "true",
	}

	k8sDeploy, err := deploySpecToK8sDeploy(ds, 1)
	if err != nil {
		t.Fatal("got unexpected error:", err)
	}
	got := k8sDeploy.Spec.Template.ObjectMeta.Annotations

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v; want %v", got, want)
	}
}
