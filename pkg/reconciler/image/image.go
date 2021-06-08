package image

import (
	"context"
	"fmt"
	"reflect"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	corev1 "k8s.io/client-go/listers/core/v1"
	"knative.dev/caching/pkg/apis/caching/v1alpha1"
	imagecachereconciler "knative.dev/caching/pkg/client/injection/reconciler/caching/v1alpha1/image"
	imagecachelisters "knative.dev/caching/pkg/client/listers/caching/v1alpha1"
	"knative.dev/pkg/apis/duck"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/reconciler"

	imagewarmclientset "knative.dev/cache-imagewarm/pkg/client/clientset/versioned"
	imagewarmlisters "knative.dev/cache-imagewarm/pkg/client/listers/caching/v1alpha1"
	"knative.dev/cache-imagewarm/pkg/reconciler/imagewarm"
)

const (
	notReconciledReason  = "ReconcileImageCacheFailed"
	notReconciledMessage = "ImageCache reconciliation failed"
)

// Reconciler implements controller.Reconciler for Image resources.
type Reconciler struct {

	// Listers index properties about resources
	ImageWarmerLister imagewarmlisters.ImageWarmLister
	ImageCacheLister  imagecachelisters.ImageLister
	NodeLister        corev1.NodeLister
	ImageWarmClient   imagewarmclientset.Interface
}

// Check that our Reconciler implements Interface
var (
	_ imagecachereconciler.Interface = (*Reconciler)(nil)
	_ imagecachereconciler.Interface = (*Reconciler)(nil)
)

func (r Reconciler) ReconcileKind(ctx context.Context, i *v1alpha1.Image) reconciler.Event {
	logger := logging.FromContext(ctx)

	logger.Infof("Reconcile ImageWarm %s/%s", i.Namespace, i.Name)
	if !i.DeletionTimestamp.IsZero() {
		logger.Infof("ImageCache %s/%s is being deleted... ...", i.Namespace, i.Name)
		return nil
	}

	reconcileErr := r.reconcileImageWarm(ctx, i)
	if reconcileErr != nil {
		logger.Errorw("Failed to reconcile ImageWarm: ", reconcileErr.Error())
		i.Status.MarkReadyFalse(notReconciledReason, notReconciledMessage)
		return reconcileErr
	}

	return r.PropagateImageCacheReadyStatus(i)

}
func (r Reconciler) reconcileImageWarm(ctx context.Context, i *v1alpha1.Image) error {
	logger := logging.FromContext(ctx)
	nodeList, err := r.NodeLister.List(labels.Everything())
	if err != nil {
		return fmt.Errorf("failed to list nodes :%s/%s when reconcileImageWarm, err: %s", i.Namespace, i.Name, err.Error())
	}
	for _, node := range nodeList {

		if r.shouldPullImage(node) {
			err := r.applyImageWarm(ctx, i, node.Name)
			if err != nil {
				logger.Errorf("failed to apply imageWarm for Node:%s", i.Name, err.Error())
				return err
			}
			// delete imageWarm
		} else {
			err := r.deleteImageWarm(ctx, i, node.Name)
			if err != nil {
				logger.Errorf("failed to delete imageWarm for Node:%s", i.Name, err.Error())
				return err
			}
		}
	}
	return nil
}

func (r Reconciler) deleteImageWarm(ctx context.Context, i *v1alpha1.Image, nodeName string) error {
	imageWarmName := imagewarm.GetImageWarmByImageAndNode(i, nodeName)
	_, err := r.ImageWarmClient.CachingV1alpha1().ImageWarms(i.Namespace).Get(ctx, imageWarmName, metav1.GetOptions{})

	if errors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to list imagewarm of Node: %s for imageCache :%s/%s when deleteImageWarm, err: %s ",
			nodeName, i.Namespace, i.Name, err.Error())
	}

	err = r.ImageWarmClient.CachingV1alpha1().ImageWarms(i.Namespace).Delete(ctx, imageWarmName, metav1.DeleteOptions{})

	if err != nil {
		return fmt.Errorf("Fail to delete imagewarm of Node: %s for imageCache :%s/%s when deleteImageWarm, err: %s ",
			nodeName, i.Namespace, i.Name, err.Error())
	}

	return nil
}

func (r Reconciler) applyImageWarm(ctx context.Context, i *v1alpha1.Image, nodeName string) error {
	logger := logging.FromContext(ctx)

	// imagewarm' name  <image cache name>-on-<node name>
	imageWarmName := fmt.Sprintf("%s-on-%s", i.Name, nodeName)
	originImagewarm, err := r.ImageWarmClient.CachingV1alpha1().ImageWarms(i.Namespace).Get(ctx, imageWarmName, metav1.GetOptions{})

	imageWarm := imagewarm.MakeImageWarm(i, nodeName)

	// create imagewarm
	if errors.IsNotFound(err) {

		_, err = r.ImageWarmClient.CachingV1alpha1().ImageWarms(i.Namespace).Create(ctx, imageWarm, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("Fail to create imagewarm for image %s on node %s ", i.Spec.Image, nodeName)
		}
	} else if err != nil {
		return fmt.Errorf("Fail to get imagewarm for imagecache %s/%s, err: %s ", i.Namespace, i.Name, err.Error())
	}

	// update imagewarm if it exists, to trigger reconciler to enqueue image again.
	newImagewarm := originImagewarm.DeepCopy()
	newImagewarm.Annotations = imageWarm.Annotations
	newImagewarm.Labels = imageWarm.Labels
	newImagewarm.Spec = imageWarm.Spec

	if reflect.DeepEqual(newImagewarm.Labels, originImagewarm.Labels) &&
		reflect.DeepEqual(newImagewarm.Annotations, originImagewarm.Annotations) &&

		reflect.DeepEqual(newImagewarm.Spec.ImagePullSecrets, originImagewarm.Spec.ImagePullSecrets) &&
		newImagewarm.Spec.NodeName == originImagewarm.Spec.NodeName {
		return nil
	}

	patchImagewarm, err := duck.CreateMergePatch(originImagewarm, newImagewarm)
	if err != nil {
		logger.Errorf("failed to createMergePatch for imagewarm %s/%s: %w", originImagewarm.Namespace, originImagewarm.Name, err)
		return err
	}

	if len(patchImagewarm) == 0 {
		return nil
	}

	_, err = r.ImageWarmClient.CachingV1alpha1().ImageWarms(i.Namespace).Patch(ctx, originImagewarm.Name,
		types.MergePatchType, patchImagewarm, metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("Fail to patch imagewarm for image %s on node %s ", i.Spec.Image, nodeName)
	}
	return err
}

func (r Reconciler) PropagateImageCacheReadyStatus(i *v1alpha1.Image) error {
	imageWarmList, err := r.ImageWarmerLister.List(labels.SelectorFromSet(map[string]string{
		imagewarm.OwnerRefName:      i.Name,
		imagewarm.OwnerRefNameSpace: i.Namespace,
	}))
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to list imagewarm for imageCache :%s/%s when propagate status, err: %s", i.Namespace, i.Name, err.Error())
	}

	for _, warm := range imageWarmList {
		if !warm.Status.IsReady() {
			i.Status.MarkReadyFalse("ResourceNotReady", fmt.Sprintf("ImageWarm %s on Node: %s Not Ready", warm.Name, warm.Spec.NodeName))
			return nil
		}
	}
	i.Status.MarkReadyTrue()
	return nil
}

func (r Reconciler) shouldPullImage(node *v1.Node) bool {
	if node.Spec.Taints != nil || node.Spec.Unschedulable == true {
		return false
	}
	return true
}

func (r Reconciler) AddNode(ctx context.Context, h func(interface{})) func(obj interface{}) {

	return func(obj interface{}) {
		logger := logging.FromContext(ctx)

		newNode, ok := obj.(*v1.Node)
		if !ok {
			logger.Errorf("unexpected type %T, expected Node")
			return
		}

		if !r.shouldPullImage(newNode) {
			return
		}

		imageList, err := r.ImageCacheLister.List(labels.Everything())

		if err != nil {
			logger.Errorf("Error enqueueing imageCache sets: %v", err)
			return
		}
		for _, cache := range imageList {
			h(cache)
		}
	}
}

func (r Reconciler) UpdateNode(ctx context.Context, h func(interface{})) func(oldObj, newObj interface{}) {

	return func(oldObj, newObj interface{}) {
		logger := logging.FromContext(ctx)
		newNode, ok1 := newObj.(*v1.Node)
		oldNode, ok2 := oldObj.(*v1.Node)
		if !ok1 || !ok2 {
			logger.Errorf("unexpected type %T, expected Node")
			return
		}

		// Node with Taints will not pull image
		if r.shouldPullImage(newNode) && r.shouldPullImage(oldNode) ||
			!r.shouldPullImage(newNode) && !r.shouldPullImage(oldNode) {
			return
		}

		imageList, err := r.ImageCacheLister.List(labels.Everything())

		if err != nil {
			logger.Errorf("Error enqueueing imageCache sets: %v", err)
			return
		}

		for _, cache := range imageList {
			h(cache)
		}
	}
}
