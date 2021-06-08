package imagewarm

import (
	"fmt"
	"knative.dev/cache-imagewarm/pkg/apis/caching"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	imagecachev1alpha1 "knative.dev/caching/pkg/apis/caching/v1alpha1"
	"knative.dev/pkg/kmeta"
	"knative.dev/serving/pkg/apis/serving"

	cachingv1alpha1 "knative.dev/cache-imagewarm/pkg/apis/caching/v1alpha1"
)

const NodeLabelKey = serving.GroupName + "/nodeName"
const OwnerRefName = caching.GroupName + "/ownerRefName"
const OwnerRefNameSpace = caching.GroupName + "/ownerRefNameSpace"
const UpdateTimeLabelKey = serving.GroupName + "/updateTimestamp"

func MakeImageWarm(imageCache *imagecachev1alpha1.Image, nodeName string) *cachingv1alpha1.ImageWarm {

	warm := &cachingv1alpha1.ImageWarm{
		ObjectMeta: metav1.ObjectMeta{
			Name:            fmt.Sprintf("%s-on-%s", imageCache.Name, nodeName),
			Namespace:       imageCache.Namespace,
			Labels:          imageCache.Labels,
			Annotations:     imageCache.Annotations,
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(imageCache)},
		},
		Spec: cachingv1alpha1.ImageWarmSpec{
			Image:            imageCache.Spec.Image,
			NodeName:         nodeName,
			ImagePullSecrets: imageCache.Spec.ImagePullSecrets,
		},
	}
	if warm.Labels == nil {
		warm.Labels = make(map[string]string)
	}
	if warm.Annotations == nil {
		warm.Annotations = make(map[string]string)
	}
	warm.Labels[NodeLabelKey] = nodeName
	warm.Labels[OwnerRefName] = imageCache.Name
	warm.Labels[OwnerRefNameSpace] = imageCache.Namespace

	return warm
}

func GetImageWarmByImageAndNode(i *imagecachev1alpha1.Image, nodeName string) string {
	// imagewarm' name  <image cache name>-on-<node name>
	imageWarmName := fmt.Sprintf("%s-on-%s", i.Name, nodeName)
	return imageWarmName
}
