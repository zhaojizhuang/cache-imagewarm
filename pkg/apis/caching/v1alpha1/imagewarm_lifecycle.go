/*
Copyright 2021 The Knative Authors

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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

const (
	// ImageWarmConditionReady is set when the revision is starting to materialize
	// runtime resources, and becomes true when those resources are ready.
	ImageWarmConditionReady = apis.ConditionReady
)

var condSet = apis.NewLivingConditionSet()

// GetGroupVersionKind implements kmeta.OwnerRefable
func (i *ImageWarm) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("ImageWarm")
}

// GetConditionSet retrieves the condition set for this resource. Implements the KRShaped interface.
func (i *ImageWarm) GetConditionSet() apis.ConditionSet {
	return condSet
}

// InitializeConditions sets the initial values to the conditions.
func (is *ImageWarmStatus) InitializeConditions() {
	condSet.Manage(is).InitializeConditions()
}

// IsReady looks at the conditions and if the Status has a condition
// ImageWarmConditionReady returns true if ConditionStatus is True
func (i *ImageWarm) IsReady() bool {
	is := i.Status
	return is.ObservedGeneration == i.Generation &&
		is.GetCondition(ImageWarmConditionReady).IsTrue()
}

// GetStatus retrieves the status of the Image. Implements the KRShaped interface.
func (i *ImageWarm) GetStatus() *duckv1.Status {
	return &i.Status.Status
}

// MarkImageNotReady marks the "ImageWarmConditionReady" condition to unknown.
func (is *ImageWarmStatus) MarkReadyUnknown() {
	condSet.Manage(is).MarkUnknown(ImageWarmConditionReady, "Uninitialized", "Waiting for Resource to be ready")
}

// MarkImageReady marks the "ImageWarmConditionReady" condition to unknown.
func (is *ImageWarmStatus) MarkReadyTrue() {
	condSet.Manage(is).MarkTrue(ImageWarmConditionReady)
}

// MarkImageFailed marks the "ImageWarmConditionReady" condition to false.
func (is *ImageWarmStatus) MarkReadyFalse(reason, message string) {
	condSet.Manage(is).MarkFalse(ImageWarmConditionReady, reason, message)
}
