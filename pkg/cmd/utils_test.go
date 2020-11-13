/*
Copyright 2020 the Velero contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/vmware-tanzu/velero-plugin-for-vsphere/pkg/constants"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeclientfake "k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestGetVersionFromImage(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		containers []corev1.Container
		expected   string
	} {
		{
			name:       "Valid image string should return non-empty version",
			key:        "cloud-provider-vsphere/csi/release/driver",
			containers: []corev1.Container{
				{
					Image: "gcr.io/cloud-provider-vsphere/csi/release/driver:corev1.0.1",
				},
			},
			expected:   "corev1.0.1",
		},
		{
			name:       "Valid image string should return non-empty version",
			key:        "cloud-provider-vsphere/csi/release/driver",
			containers: []corev1.Container{
				{
					Image: "cloud-provider-vsphere/csi/release/driver:v2.0.0",
				},
			},
			expected:   "v2.0.0",
		},
		{
			name:       "Valid image string should return non-empty version",
			key:        "cloud-provider-vsphere/csi/release/driver",
			containers: []corev1.Container{
				{
					Image: "myregistry/cloud-provider-vsphere/csi/release/driver:v2.0.0",
				},
			},
			expected:   "v2.0.0",
		},
		{
			name:       "Valid image string should return non-empty version",
			key:        "cloud-provider-vsphere/csi/release/driver",
			containers: []corev1.Container{
				{
					Image: "myregistry/level1/level2/cloud-provider-vsphere/csi/release/driver:v2.0.0",
				},
			},
			expected:   "v2.0.0",
		},
		{
			name:     "Invalid image name should return empty string",
			key:      "cloud-provider-vsphere/csi/release/driver",
			containers: []corev1.Container{
				{
					Image: "gcr.io/csi/release/driver:corev1.0.1",
				},
			},
			expected: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			version := GetVersionFromImage(test.containers, test.key)
			assert.Equal(t, test.expected, version)
		})
	}
}

func TestGetCompatibleRepoAndTagFromPluginImage(t *testing.T) {
	tests := []struct {
		name string
		veleroDeployment *appsv1.Deployment
		targetContainer string
		expectedImage string
		expectedError error
	}{
		{
			name: "ExpectedPluginImageFromDockerhub",
			veleroDeployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "velero",
					Name:      constants.VeleroDeployment,
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							InitContainers: []corev1.Container{
								{
									Name: "velero-plugin-for-vsphere",
									Image: "vsphereveleroplugin/velero-plugin-for-vsphere:1.1.0-rc2",
								},
							},
						},
					},
				},
			},
			targetContainer: constants.BackupDriverForPlugin,
			expectedImage: "vsphereveleroplugin/" + constants.BackupDriverForPlugin + ":1.1.0-rc2",
			expectedError: nil,
		},
		{
			name: "UnexpectedDeployment",
			veleroDeployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "xyz",
					Name:      "not-velero",
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							InitContainers: []corev1.Container{
								{
									Name: "velero-plugin-for-vsphere",
									Image: "vsphereveleroplugin/velero-plugin-for-vsphere:1.1.0-rc2",
								},
							},
						},
					},
				},
			},
			targetContainer: constants.BackupDriverForPlugin,
			expectedImage: "",
			expectedError: errors.Errorf("Failed to get velero deployment in namespace %s", "xyz"),
		},
		{
			name: "NoExpectedPluginImageAvailable",
			veleroDeployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "velero",
					Name:      constants.VeleroDeployment,
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							InitContainers: []corev1.Container{
								{
									Name: "velero-plugin-for-aws",
									Image: "velero/velero-plugin-for-aws:1.1.0",
								},
							},
						},
					},
				},
			},
			targetContainer: constants.BackupDriverForPlugin,
			expectedImage: "",
			expectedError: errors.New("The plugin, velero-plugin-for-vsphere, was not added as an init container of Velero deployment"),
		},
		{
			name: "ExpectedPluginImageFromOnPremRegistry",
			veleroDeployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "velero",
					Name:      constants.VeleroDeployment,
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							InitContainers: []corev1.Container{
								{
									Name: "velero-plugin-for-vsphere",
									Image: "xyz-repo.opq.abc:8888/one/two/velero-plugin-for-vsphere:1.1.0-rc2",
								},
							},
						},
					},
				},
			},
			targetContainer: constants.BackupDriverForPlugin,
			expectedImage: "xyz-repo.opq.abc:8888/one/two/" + constants.BackupDriverForPlugin + ":1.1.0-rc2",
			expectedError: nil,
		},
		{
			name: "ExpectedPluginImageWithoutTag",
			veleroDeployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "velero",
					Name:      constants.VeleroDeployment,
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							InitContainers: []corev1.Container{
								{
									Name: "velero-plugin-for-vsphere",
									Image: "xyz-repo.opq.abc:8888/one/two/velero-plugin-for-vsphere",
								},
							},
						},
					},
				},
			},
			targetContainer: constants.DataManagerForPlugin,
			expectedImage: "xyz-repo.opq.abc:8888/one/two/" + constants.DataManagerForPlugin,
			expectedError: nil,
		},
		{
			name: "ExpectedPluginImageWithoutRepo",
			veleroDeployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "velero",
					Name:      constants.VeleroDeployment,
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							InitContainers: []corev1.Container{
								{
									Name: "velero-plugin-for-vsphere",
									Image: "velero-plugin-for-vsphere:1.1.0-rc2",
								},
							},
						},
					},
				},
			},
			targetContainer: constants.DataManagerForPlugin,
			expectedImage: constants.DataManagerForPlugin + ":1.1.0-rc2",
			expectedError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			kubeClient := kubeclientfake.NewSimpleClientset(test.veleroDeployment)
			actualImage, actualError := GetCompatibleRepoAndTagFromPluginImage(kubeClient, test.veleroDeployment.Namespace, test.targetContainer)
			assert.Equal(t, actualImage, test.expectedImage)
			assert.Equal(t, actualError == nil, test.expectedError == nil)
			if actualError != nil {
				assert.Equal(t, actualError.Error(), test.expectedError.Error())
			}
		})
	}
}

