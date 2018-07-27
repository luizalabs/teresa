package k8s

import (
	"testing"

	"github.com/luizalabs/teresa/pkg/server/spec"
	"k8s.io/api/apps/v1beta2"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/batch/v1beta1"
	k8sv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAddVolumeMountOfSecrets(t *testing.T) {
	var testCases = []struct {
		vols       []k8sv1.VolumeMount
		volName    string
		path       string
		expectedVM int
	}{
		{
			vols:       []k8sv1.VolumeMount{{Name: "vm"}},
			volName:    "s",
			path:       "p",
			expectedVM: 2,
		},
		{
			vols:       []k8sv1.VolumeMount{{Name: "s"}},
			volName:    "s",
			path:       "p",
			expectedVM: 1,
		},
		{
			vols:       []k8sv1.VolumeMount{{Name: "vm"}, {Name: "s"}},
			volName:    "s",
			path:       "p",
			expectedVM: 2,
		},
	}

	for _, tc := range testCases {
		vols := addVolumeMountOfSecrets(tc.vols, tc.volName, tc.path)
		if actual := len(vols); actual != tc.expectedVM {
			t.Errorf("expected %d vols, got %d", tc.expectedVM, actual)
		}
		found := false
		for _, vol := range vols {
			if vol.Name == tc.volName {
				found = true
				break
			}
		}
		if !found {
			t.Error("volume ref to secret not found")
		}
	}
}

func TestAddVolumeOfSecretFile(t *testing.T) {
	var testCases = []struct {
		vols       []k8sv1.Volume
		volName    string
		secretName string
		fileName   string
	}{
		{
			vols:       []k8sv1.Volume{{Name: "vl"}},
			volName:    "volSecret",
			secretName: "secret",
			fileName:   "fileSecret",
		},
		{
			vols: []k8sv1.Volume{{
				Name: "volSecret",
				VolumeSource: k8sv1.VolumeSource{
					Secret: &k8sv1.SecretVolumeSource{
						SecretName: "secret",
						Items:      []k8sv1.KeyToPath{},
					},
				},
			}},
			volName:    "volSecret",
			secretName: "secret",
			fileName:   "fileSecret",
		},
		{
			vols: []k8sv1.Volume{{
				Name: "volSecret",
				VolumeSource: k8sv1.VolumeSource{
					Secret: &k8sv1.SecretVolumeSource{
						SecretName: "secret",
						Items:      []k8sv1.KeyToPath{{Key: "fs", Path: "fs"}},
					},
				},
			}},
			volName:    "volSecret",
			secretName: "secret",
			fileName:   "fs",
		},
	}

	for _, tc := range testCases {
		vols := addVolumeOfSecretFile(tc.vols, tc.volName, tc.secretName, tc.fileName)
		found := false
		for _, vol := range vols {
			if vol.Name == tc.volName {
				found = true
				if actual := vol.Secret.SecretName; actual != tc.secretName {
					t.Errorf("expected %s, got %s", tc.secretName, actual)
				}
				foundItem := false
				for _, item := range vol.Secret.Items {
					if item.Key == tc.fileName {
						foundItem = true
						break
					}
				}
				if !foundItem {
					t.Error("item not found in volume")
				}
				break
			}
		}
		if !found {
			t.Error("volume ref to secret not found")
		}
	}
}

func deploySpec(envs []k8sv1.EnvVar) *v1beta2.Deployment {
	evs := make([]k8sv1.EnvVar, len(envs))
	copy(evs, envs)

	return &v1beta2.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
		Spec: v1beta2.DeploymentSpec{
			Template: k8sv1.PodTemplateSpec{
				Spec: k8sv1.PodSpec{
					Containers: []k8sv1.Container{
						{
							Name: "test",
							Env:  evs,
						},
					},
				},
			},
		},
	}
}

func cronJobSpec(envs []k8sv1.EnvVar) *v1beta1.CronJob {
	evs := make([]k8sv1.EnvVar, len(envs))
	copy(evs, envs)

	return &v1beta1.CronJob{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
		Spec: v1beta1.CronJobSpec{
			JobTemplate: v1beta1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: k8sv1.PodTemplateSpec{
						Spec: k8sv1.PodSpec{
							Containers: []k8sv1.Container{
								{
									Name: "test",
									Env:  evs,
								},
							},
						},
					},
				},
			},
		},
	}
}

func TestRemoveEnvVarsWithSecretsFromDeployAndCronJob(t *testing.T) {
	var testCases = []struct {
		envs     []k8sv1.EnvVar
		toRemove []string
		expected []k8sv1.EnvVar
	}{
		{
			envs:     []k8sv1.EnvVar{{Name: "FOO"}},
			toRemove: []string{"FOO"},
			expected: []k8sv1.EnvVar{},
		},
		{
			envs:     []k8sv1.EnvVar{{Name: "FOO"}, {Name: "BAR"}},
			toRemove: []string{"FOO"},
			expected: []k8sv1.EnvVar{{Name: "BAR"}},
		},
		{
			envs:     []k8sv1.EnvVar{{Name: "FOO"}, {Name: "BAR"}},
			toRemove: []string{"BAR"},
			expected: []k8sv1.EnvVar{{Name: "FOO"}},
		},
		{
			envs:     []k8sv1.EnvVar{{Name: "FOO"}, {Name: "BAR"}},
			toRemove: []string{"FOO", "BAR"},
			expected: []k8sv1.EnvVar{},
		},
	}

	for _, tc := range testCases {
		d := removeEnvVarsWithSecretsFromDeploy(deploySpec(tc.envs), tc.toRemove)
		cj := removeEnvVarsWithSecretsFromCronJob(cronJobSpec(tc.envs), tc.toRemove)
		cns := []k8sv1.Container{
			d.Spec.Template.Spec.Containers[0],
			cj.Spec.JobTemplate.Spec.Template.Spec.Containers[0],
		}
		for _, cn := range cns {
			if len(tc.expected) == 0 && len(cn.Env) != 0 {
				t.Errorf("expected no envs, there are some: %v - %v - %v", cn.Env, tc.toRemove, tc.envs)
			}
			for i := range tc.expected {
				actual := cn.Env[i]
				if actual != tc.expected[i] {
					t.Errorf("expected %s, got %s", tc.expected[i], actual)
				}
			}
		}
	}
}

func deploySpecWithVolumes(vols []k8sv1.Volume, volMounts []k8sv1.VolumeMount) *v1beta2.Deployment {
	d := deploySpec([]k8sv1.EnvVar{})
	d.Spec.Template.Spec.Volumes = vols
	d.Spec.Template.Spec.Containers[0].VolumeMounts = volMounts
	return d
}

func cronJobSpecWithVolumes(vols []k8sv1.Volume, volMounts []k8sv1.VolumeMount) *v1beta1.CronJob {
	cj := cronJobSpec([]k8sv1.EnvVar{})
	cj.Spec.JobTemplate.Spec.Template.Spec.Volumes = vols
	cj.Spec.JobTemplate.Spec.Template.Spec.Containers[0].VolumeMounts = volMounts
	return cj
}

func testRemoveVolumesWithSecrets(fromDeploy bool) func(*testing.T) {
	return func(t *testing.T) {
		var testCases = []struct {
			Volumes           []k8sv1.Volume
			VolumeMounts      []k8sv1.VolumeMount
			toRemove          []string
			expectedVols      []k8sv1.Volume
			expectedVolMounts []k8sv1.VolumeMount
		}{
			{
				Volumes: []k8sv1.Volume{{
					Name: spec.AppSecretName,
					VolumeSource: k8sv1.VolumeSource{
						Secret: &k8sv1.SecretVolumeSource{
							Items: []k8sv1.KeyToPath{{Key: "FOO"}},
						},
					},
				}},
				VolumeMounts: []k8sv1.VolumeMount{{
					Name: spec.AppSecretName,
				}},
				toRemove:          []string{"FOO"},
				expectedVols:      []k8sv1.Volume{},
				expectedVolMounts: []k8sv1.VolumeMount{},
			},
			{
				Volumes: []k8sv1.Volume{{
					Name: spec.AppSecretName,
					VolumeSource: k8sv1.VolumeSource{
						Secret: &k8sv1.SecretVolumeSource{
							Items: []k8sv1.KeyToPath{{Key: "FOO"}, {Key: "BAR"}},
						},
					},
				}},
				VolumeMounts: []k8sv1.VolumeMount{{
					Name: spec.AppSecretName,
				}},
				toRemove: []string{"FOO"},
				expectedVols: []k8sv1.Volume{{
					Name: spec.AppSecretName,
					VolumeSource: k8sv1.VolumeSource{
						Secret: &k8sv1.SecretVolumeSource{
							Items: []k8sv1.KeyToPath{{Key: "BAR"}},
						},
					},
				}},
				expectedVolMounts: []k8sv1.VolumeMount{{
					Name: spec.AppSecretName,
				}},
			},
			{
				Volumes: []k8sv1.Volume{{
					Name: spec.AppSecretName,
					VolumeSource: k8sv1.VolumeSource{
						Secret: &k8sv1.SecretVolumeSource{
							Items: []k8sv1.KeyToPath{{Key: "FOO"}, {Key: "BAR"}},
						},
					},
				}},
				VolumeMounts: []k8sv1.VolumeMount{{
					Name: spec.AppSecretName,
				}},
				toRemove:          []string{"FOO", "BAR"},
				expectedVols:      []k8sv1.Volume{},
				expectedVolMounts: []k8sv1.VolumeMount{},
			},
		}

		for _, tc := range testCases {
			var spec k8sv1.PodSpec
			if fromDeploy {
				d := deploySpecWithVolumes(tc.Volumes, tc.VolumeMounts)
				d = removeVolumesWithSecretsFromDeploy(d, tc.toRemove)
				spec = d.Spec.Template.Spec
			} else {
				cj := cronJobSpecWithVolumes(tc.Volumes, tc.VolumeMounts)
				cj = removeVolumesWithSecretsFromCronJob(cj, tc.toRemove)
				spec = cj.Spec.JobTemplate.Spec.Template.Spec
			}

			if actual := len(spec.Volumes); actual != len(tc.expectedVols) {
				t.Fatalf("expected %d, got %d vols", len(tc.expectedVols), actual)
			}
			if len(tc.expectedVols) > 0 {
				for i, item := range tc.expectedVols[0].Secret.Items {
					actual := spec.Volumes[0].Secret.Items[i].Key
					if actual != item.Key {
						t.Errorf("expected %s, got %s", item.Key, actual)
					}
				}
			}

			if actual := len(spec.Containers[0].VolumeMounts); actual != len(tc.expectedVolMounts) {
				t.Errorf("expected %d, got %d vol mounts", len(tc.expectedVolMounts), actual)
			}
		}
	}
}

func TestRemoveSecretVols(t *testing.T) {
	t.Run("TestRemoveVolumesWithSecretsFromDeploy", testRemoveVolumesWithSecrets(true))
	t.Run("TestRemoveVolumesWithSecretsFromCronJob", testRemoveVolumesWithSecrets(false))
}
