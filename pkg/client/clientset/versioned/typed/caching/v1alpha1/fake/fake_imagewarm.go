/*
Copyright 2020 The Knative Authors

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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
	v1alpha1 "knative.dev/cache-imagewarm/pkg/apis/caching/v1alpha1"
)

// FakeImageWarms implements ImageWarmInterface
type FakeImageWarms struct {
	Fake *FakeCachingV1alpha1
	ns   string
}

var imagewarmsResource = schema.GroupVersionResource{Group: "caching.knative.dev", Version: "v1alpha1", Resource: "imagewarms"}

var imagewarmsKind = schema.GroupVersionKind{Group: "caching.knative.dev", Version: "v1alpha1", Kind: "ImageWarm"}

// Get takes name of the imageWarm, and returns the corresponding imageWarm object, and an error if there is any.
func (c *FakeImageWarms) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.ImageWarm, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(imagewarmsResource, c.ns, name), &v1alpha1.ImageWarm{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ImageWarm), err
}

// List takes label and field selectors, and returns the list of ImageWarms that match those selectors.
func (c *FakeImageWarms) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.ImageWarmList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(imagewarmsResource, imagewarmsKind, c.ns, opts), &v1alpha1.ImageWarmList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ImageWarmList{ListMeta: obj.(*v1alpha1.ImageWarmList).ListMeta}
	for _, item := range obj.(*v1alpha1.ImageWarmList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested imageWarms.
func (c *FakeImageWarms) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(imagewarmsResource, c.ns, opts))

}

// Create takes the representation of a imageWarm and creates it.  Returns the server's representation of the imageWarm, and an error, if there is any.
func (c *FakeImageWarms) Create(ctx context.Context, imageWarm *v1alpha1.ImageWarm, opts v1.CreateOptions) (result *v1alpha1.ImageWarm, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(imagewarmsResource, c.ns, imageWarm), &v1alpha1.ImageWarm{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ImageWarm), err
}

// Update takes the representation of a imageWarm and updates it. Returns the server's representation of the imageWarm, and an error, if there is any.
func (c *FakeImageWarms) Update(ctx context.Context, imageWarm *v1alpha1.ImageWarm, opts v1.UpdateOptions) (result *v1alpha1.ImageWarm, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(imagewarmsResource, c.ns, imageWarm), &v1alpha1.ImageWarm{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ImageWarm), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeImageWarms) UpdateStatus(ctx context.Context, imageWarm *v1alpha1.ImageWarm, opts v1.UpdateOptions) (*v1alpha1.ImageWarm, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(imagewarmsResource, "status", c.ns, imageWarm), &v1alpha1.ImageWarm{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ImageWarm), err
}

// Delete takes name of the imageWarm and deletes it. Returns an error if one occurs.
func (c *FakeImageWarms) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(imagewarmsResource, c.ns, name), &v1alpha1.ImageWarm{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeImageWarms) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(imagewarmsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.ImageWarmList{})
	return err
}

// Patch applies the patch and returns the patched imageWarm.
func (c *FakeImageWarms) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.ImageWarm, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(imagewarmsResource, c.ns, name, pt, data, subresources...), &v1alpha1.ImageWarm{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ImageWarm), err
}