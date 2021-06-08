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

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
	cachingv1alpha1 "knative.dev/cache-imagewarm/pkg/apis/caching/v1alpha1"
	versioned "knative.dev/cache-imagewarm/pkg/client/clientset/versioned"
	internalinterfaces "knative.dev/cache-imagewarm/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha1 "knative.dev/cache-imagewarm/pkg/client/listers/caching/v1alpha1"
)

// ImageWarmInformer provides access to a shared informer and lister for
// ImageWarms.
type ImageWarmInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.ImageWarmLister
}

type imageWarmInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewImageWarmInformer constructs a new informer for ImageWarm type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewImageWarmInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredImageWarmInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredImageWarmInformer constructs a new informer for ImageWarm type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredImageWarmInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CachingV1alpha1().ImageWarms(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CachingV1alpha1().ImageWarms(namespace).Watch(context.TODO(), options)
			},
		},
		&cachingv1alpha1.ImageWarm{},
		resyncPeriod,
		indexers,
	)
}

func (f *imageWarmInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredImageWarmInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *imageWarmInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&cachingv1alpha1.ImageWarm{}, f.defaultInformer)
}

func (f *imageWarmInformer) Lister() v1alpha1.ImageWarmLister {
	return v1alpha1.NewImageWarmLister(f.Informer().GetIndexer())
}