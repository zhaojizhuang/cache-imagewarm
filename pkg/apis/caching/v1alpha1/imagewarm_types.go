/*
Copyright 2021 The Knative Authors.

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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
)

// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ImageWarm is a Knative abstraction that encapsulates the interface by which Knative
// components express a desire to have a particular image cached.
type ImageWarm struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the desired state of the ImageWarm (from the client).
	// +optional
	Spec ImageWarmSpec `json:"spec,omitempty"`

	// Status communicates the observed state of the ImageWarm (from the reconciler).
	// +optional
	Status ImageWarmStatus `json:"status,omitempty"`
}

// Check that ImageWarm can be validated and defaulted.
var _ apis.Validatable = (*ImageWarm)(nil)
var _ apis.Defaultable = (*ImageWarm)(nil)
var _ kmeta.OwnerRefable = (*ImageWarm)(nil)

// ImageWarmSpec holds the desired state of the ImageWarm (from the client).
type ImageWarmSpec struct {

	// Image is the name of the container image url to cache across the cluster.
	Image string `json:"image"`

	// NodeName is the names of the node where imagewarmer will pull image.
	NodeName string `json:"nodeName"`

	// ImagePullSecrets contains the names of the Kubernetes Secrets containing login
	// information used by the Pods which will run this container.
	// +optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
}

// ImageWarmStatus communicates the observed state of the ImageWarm (from the reconciler).
// ImageStatus communicates the observed state of the Image (from the controller).
type ImageWarmStatus struct {
	duckv1.Status `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ImageWarmList is a list of ImageWarm resources
type ImageWarmList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ImageWarm `json:"items"`
}

// IsReady looks at the conditions and if the Status has a condition
// ImageWarmConditionReady returns true if ConditionStatus is True
func (rs *ImageWarmStatus) IsReady() bool {
	if c := rs.GetCondition(ImageWarmConditionReady); c != nil {
		return c.Status == corev1.ConditionTrue
	}
	return false
}
