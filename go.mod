module knative.dev/cache-imagewarm

go 1.14

require (
	github.com/Microsoft/go-winio v0.4.16 // indirect
	github.com/containerd/containerd v1.4.4 // indirect
	github.com/docker/docker v20.10.5+incompatible
	go.uber.org/zap v1.16.0
	k8s.io/api v0.19.7
	k8s.io/apimachinery v0.19.7
	k8s.io/client-go v0.19.7
	k8s.io/code-generator v0.19.7
	k8s.io/kube-openapi v0.0.0-20210113233702-8566a335510f
	k8s.io/kubernetes v1.19.7
	knative.dev/caching v0.0.0-20210506040209-3d48f8dc4abc
	knative.dev/hack v0.0.0-20210428122153-93ad9129c268
	knative.dev/hack/schema v0.0.0-20210325223819-b6ab329907d3
	knative.dev/pkg v0.0.0-20210428023153-5a308fa62139
	knative.dev/serving v0.22.0
)

replace (
	k8s.io/api => k8s.io/api v0.19.7
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.19.7
	k8s.io/apimachinery => k8s.io/apimachinery v0.19.7
	k8s.io/apiserver => k8s.io/apiserver v0.19.7
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.19.7
	k8s.io/client-go => k8s.io/client-go v0.19.7
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.19.7
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.19.7
	k8s.io/code-generator => k8s.io/code-generator v0.19.7
	k8s.io/component-base => k8s.io/component-base v0.19.7
	k8s.io/cri-api => k8s.io/cri-api v0.19.7
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.19.7
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.19.7
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.19.7
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.19.7
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.19.7
	k8s.io/kubectl => k8s.io/kubectl v0.19.7
	k8s.io/kubelet => k8s.io/kubelet v0.19.7
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.19.7
	k8s.io/metrics => k8s.io/metrics v0.19.7
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.19.7
)
