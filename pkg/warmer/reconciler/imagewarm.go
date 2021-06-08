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
	"os"

	corev1 "k8s.io/client-go/listers/core/v1"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/reconciler"

	"knative.dev/cache-imagewarm/pkg/apis/caching/v1alpha1"
	imagewarmclientset "knative.dev/cache-imagewarm/pkg/client/clientset/versioned"
	"knative.dev/cache-imagewarm/pkg/client/injection/reconciler/caching/v1alpha1/imagewarm"
	imagewarmlisters "knative.dev/cache-imagewarm/pkg/client/listers/caching/v1alpha1"
	"knative.dev/cache-imagewarm/pkg/warmer/images"
)

const GlobalPullSecret = "pullsecret"

var NodeName string

func init() {
	NodeName = os.Getenv("NODE_NAME")
	if NodeName == "" {
		panic("NODE_NAME environment not set")
	}
}

// Reconciler implements addressableservicereconciler.Interface for
// AddressableService resources.
type Reconciler struct {

	// Listers index properties about resources
	ImageWarmerLister imagewarmlisters.ImageWarmLister
	//ImageCacheLister  imagecachelisters.ImageLister

	ImageWarmClient imagewarmclientset.Interface

	Secretlister corev1.SecretLister

	ImagePuller images.ImagePuller
}

// Check that our Reconciler implements Interface
var (
	_ imagewarm.Interface = (*Reconciler)(nil)
	_ imagewarm.Interface = (*Reconciler)(nil)
	_ imagewarm.Finalizer = (*Reconciler)(nil)
)

// FinalizeKind removes all images and imagewarm
func (r *Reconciler) FinalizeKind(ctx context.Context, i *v1alpha1.ImageWarm) reconciler.Event {
	logger := logging.FromContext(ctx)
	logger.Infof("ImageCache  %s/%s for image:%s is being deleted, we will gc image for it.", i.Namespace, i.Name, i.Spec.Image)

	r.ImagePuller.StopPullImage(ctx, i.Spec.Image)

	return nil
}

// ReconcileKind implements Interface.ReconcileKind.
func (r *Reconciler) ReconcileKind(ctx context.Context, i *v1alpha1.ImageWarm) reconciler.Event {
	logger := logging.FromContext(ctx)

	logger.Infof("Reconcile ImageCache %s/%s, image:%s", i.Namespace, i.Name, i.Spec.Image)

	if !i.DeletionTimestamp.IsZero() {
		logger.Infof("ImageCache %s/%s is being deleted, so we will not pull image for it.", i.Namespace, i.Name)
		return nil
	}

	if exists, _ := r.ImagePuller.ImageExists(ctx, i.Spec.Image); exists {
		logger.Infof("Image %s for image %s/%s exists, no need to pull ! ", i.Spec.Image, i.Namespace, i.Name)
		// TODO reconcile image.status in another reconciler
		if !i.IsReady() {
			i.Status.MarkReadyTrue()
		}
		return nil
	}

	var secretName string

	// use Default Secret
	if len(i.Spec.ImagePullSecrets) == 0 {
		secretName = GlobalPullSecret
	} else {
		secretName = i.Spec.ImagePullSecrets[0].Name
	}

	secret, err := r.Secretlister.Secrets(i.Namespace).Get(secretName)
	if err != nil {
		logger.Warnf("get secret for imagecache %s/%s,err: %s", i.Namespace, i.Name, err.Error())
	}

	r.ImagePuller.PullImage(ctx, i.Spec.Image, secret)
	// TODO reconcile image.status in another reconciler
	i.Status.MarkReadyUnknown()
	return nil
}
