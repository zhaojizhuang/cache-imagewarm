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

package reconciler

import (
	"context"
	"time"

	"k8s.io/client-go/tools/cache"
	"knative.dev/caching/pkg/apis/caching/v1alpha1"
	imagecacheinformer "knative.dev/caching/pkg/client/injection/informers/caching/v1alpha1/image"
	cachereconciler "knative.dev/caching/pkg/client/injection/reconciler/caching/v1alpha1/image"
	nodeinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/node"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"

	servingclient "knative.dev/cache-imagewarm/pkg/client/injection/client"
	imagewarmerinformer "knative.dev/cache-imagewarm/pkg/client/injection/informers/caching/v1alpha1/imagewarm"
	"knative.dev/cache-imagewarm/pkg/reconciler/image"
)

// Recheck image every 10 minutes
const ControllerResyncPerion = 10 * time.Second

// NewController creates a Reconciler and returns the result of NewImpl.
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	logger := logging.FromContext(ctx)

	imageWarmInformer := imagewarmerinformer.Get(ctx)
	imageCacheInformer := imagecacheinformer.Get(ctx)
	nodeInformer := nodeinformer.Get(ctx)

	r := &image.Reconciler{
		ImageWarmerLister: imageWarmInformer.Lister(),
		ImageCacheLister:  imageCacheInformer.Lister(),
		ImageWarmClient:   servingclient.Get(ctx),
		NodeLister:        nodeInformer.Lister(),
	}
	impl := cachereconciler.NewImpl(ctx, r)

	logger.Info("Setting up event handlers.")

	imageCacheInformer.Informer().AddEventHandlerWithResyncPeriod(controller.HandleAll(impl.Enqueue), ControllerResyncPerion)

	imageWarmInformer.Informer().AddEventHandlerWithResyncPeriod(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterControllerGK(v1alpha1.Kind("Image")),
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc:    impl.EnqueueControllerOf,
			UpdateFunc: controller.PassNew(impl.EnqueueControllerOf),
		},
	}, ControllerResyncPerion)

	nodeInformer.Informer().AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc:    r.AddNode(ctx, impl.Enqueue),
		UpdateFunc: r.UpdateNode(ctx, impl.Enqueue),
		DeleteFunc: r.AddNode(ctx, impl.Enqueue),
	}, ControllerResyncPerion)

	return impl
}
