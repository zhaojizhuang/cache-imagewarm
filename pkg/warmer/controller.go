/*
Copyright 2019 The Knative Authors

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

package warmer

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	secretinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/secret"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"

	servingclient "knative.dev/cache-imagewarm/pkg/client/injection/client"
	imagewarmerinformer "knative.dev/cache-imagewarm/pkg/client/injection/informers/caching/v1alpha1/imagewarm"
	imagewarmreconciler "knative.dev/cache-imagewarm/pkg/client/injection/reconciler/caching/v1alpha1/imagewarm"
	"knative.dev/cache-imagewarm/pkg/reconciler/imagewarm"
	"knative.dev/cache-imagewarm/pkg/warmer/cri/docker"
	"knative.dev/cache-imagewarm/pkg/warmer/images"
	"knative.dev/cache-imagewarm/pkg/warmer/reconciler"
)

// Recheck image every 10 minutes
const ControllerResyncPerion = 5 * time.Second
const DockerRuntimeURI = "unix:///var/run/docker.sock"

// NewWarmDaemon creates a Reconciler and returns the result of NewImpl.
func NewWarmDaemon(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	logger := logging.FromContext(ctx)

	imageWarmInformer := imagewarmerinformer.Get(ctx)
	secretInformer := secretinformer.Get(ctx)

	r := &reconciler.Reconciler{
		ImageWarmerLister: imageWarmInformer.Lister(),
		Secretlister:      secretInformer.Lister(),
		ImageWarmClient:   servingclient.Get(ctx),
	}
	impl := imagewarmreconciler.NewImpl(ctx, r)

	logger.Info("Setting up event handlers.")

	imageWarmInformer.Informer().AddEventHandlerWithResyncPeriod(cache.FilteringResourceEventHandler{
		FilterFunc: FilterWithLabel(imagewarm.NodeLabelKey, reconciler.NodeName),
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc:    impl.EnqueueControllerOf,
			UpdateFunc: controller.PassNew(impl.EnqueueControllerOf),
		},
	}, ControllerResyncPerion)

	imageService, err := docker.NewDockerImageService(DockerRuntimeURI)
	if err != nil {
		logger.Errorf("err:%#v", err)
		return nil
	}

	puller := images.NewSerialImagePuller(imageService)

	r.ImagePuller = puller

	puller.Start()
	logger.Info("Setting up ImagePuller")

	return impl
}

// FilterWithLabel makes it simple to create FilterFunc's for use with
// cache.FilteringResourceEventHandler that filter based on a label .
func FilterWithLabel(labelKey, labelValue string) func(obj interface{}) bool {
	return func(obj interface{}) bool {
		object, ok := obj.(metav1.Object)
		if !ok {
			return false
		}

		if len(object.GetLabels()) == 0 {
			return false
		}

		value, exists := object.GetLabels()[labelKey]
		if !exists {
			return false
		} else {
			return value == labelValue
		}
	}
}
