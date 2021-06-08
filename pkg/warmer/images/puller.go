package images

import (
	"context"
	"strings"
	"sync"

	v1 "k8s.io/api/core/v1"
	"knative.dev/pkg/logging"

	"knative.dev/cache-imagewarm/pkg/warmer/cri"
)

type pullResult struct {
	imageRef string
	err      error
}

type ImagePuller interface {
	PullImage(context.Context, string, *v1.Secret)
	// TODO remove pullTask
	StopPullImage(context.Context, string)
	Start()
	ImageExists(ctx context.Context, imageRef string) (bool, error)
}

var _ ImagePuller = &serialImagePuller{}

// Maximum number of image pull requests than can be queued.
const maxImagePullRequests = 10

type serialImagePuller struct {
	imageService   cri.ImageService
	pullRequests   chan *imagePullRequest
	imagesNeedPull map[string]*imagePullRequest

	sync.RWMutex
}

func (sip *serialImagePuller) getImagePullRequest(imageRef string) (*imagePullRequest, bool) {
	sip.RLock()
	iR, ok := sip.imagesNeedPull[imageRef]
	sip.RUnlock()
	return iR, ok
}

func (sip *serialImagePuller) putImagePullRequest(imageRequest *imagePullRequest) {
	sip.Lock()
	sip.imagesNeedPull[imageRequest.imageRef] = imageRequest
	sip.Unlock()
}

func (sip *serialImagePuller) removeImagePullRequest(imageRef string) {
	sip.Lock()
	delete(sip.imagesNeedPull, imageRef)
	sip.Unlock()
}

func (sip *serialImagePuller) StopPullImage(ctx context.Context, imageRef string) {
	logger := logging.FromContext(ctx)
	logger.Infof("StopPullImage start to remote pull task for image: %s.", imageRef)

	if imagePullRequest, ok := sip.getImagePullRequest(imageRef); ok {
		if imagePullRequest.finishPull == false {
			logger.Infof("PullTask for image: %s is Running, we will stop it!", imageRef)
			imagePullRequest.cancel()
		}

		sip.removeImagePullRequest(imageRef)
	}
}

func (sip *serialImagePuller) Start() {
	go sip.processImagePullRequests()
}

func NewSerialImagePuller(imageService cri.ImageService) ImagePuller {
	imagePuller := &serialImagePuller{
		imageService:   imageService,
		pullRequests:   make(chan *imagePullRequest, maxImagePullRequests),
		imagesNeedPull: make(map[string]*imagePullRequest)}

	return imagePuller
}

type imagePullRequest struct {
	imageRef string
	//spec            cri.ImageInfo
	pullSecret *v1.Secret
	//pullChan   chan<- pullResult
	// finishPull specific whether image has been pulled
	finishPull bool
	// cancel pull image
	cancel context.CancelFunc
	ctx    context.Context

	successHandler func(imageRef string) error
}

// TODO just support serialImagePuller
func (sip *serialImagePuller) PullImage(ctx context.Context, imageRef string, pullSecret *v1.Secret) {
	logger := logging.FromContext(ctx)
	if pullRequest, ok := sip.getImagePullRequest(imageRef); ok && pullRequest.finishPull == false {
		logger.Infof("ImagePuller is pulling Image %s", imageRef)
		return
	}

	logger.Infof("ImagePuller start to pull  Image %s", imageRef)

	ctx, cancel := context.WithCancel(ctx)
	pullRequest := &imagePullRequest{
		imageRef:   imageRef,
		pullSecret: pullSecret,
		cancel:     cancel,
		ctx:        ctx,
	}

	sip.putImagePullRequest(pullRequest)

	// send to do realPull
	sip.pullRequests <- pullRequest
}

func (sip *serialImagePuller) processImagePullRequests() {

	for pullRequest := range sip.pullRequests {
		logger := logging.FromContext(pullRequest.ctx)
		logger.Infof("ImagePuller receive imagePull task,imageRef :%s", pullRequest.imageRef)

		func() {
			exist, _ := sip.ImageExists(pullRequest.ctx, pullRequest.imageRef)
			if exist {
				pullRequest.finishPull = true

				//if err := pullRequest.successHandler(pullRequest.imageRef); err != nil {
				//	logger.Infof("call handler err %s", err.Error())
				//} else {
				//
				logger.Infof("call handler success !")
				//}

			} else {
				err := sip.imageService.PullImage(pullRequest.ctx, pullRequest.imageRef, pullRequest.pullSecret)
				pullRequest.finishPull = true
				if err == nil {
					//if callErr := pullRequest.successHandler(pullRequest.imageRef); callErr != nil {
					//	logger.Errorf("call handler err %s", callErr.Error())
					//} else {
					logger.Infof("call handler success !")
					//}
				}

			}
		}()
	}
}

func (sip *serialImagePuller) ImageExists(ctx context.Context, imageRef string) (bool, error) {
	// Trim docker.io and index.docker.io
	if strings.Contains(imageRef,"docker.io"){
		splits:=strings.Split(imageRef,"/")
		imageRef = strings.Join(splits[1:],"/")
	}

	if strings.Contains(imageRef, "@sha256") {
		return sip.ImageHashExists(ctx, imageRef)
	} else {
		imageName, Tag := cri.ParseRepositoryTag(imageRef)
		return sip.ImageTagExists(ctx, imageName, Tag)
	}
}

func (sip *serialImagePuller) ImageTagExists(ctx context.Context, imageName, imageTag string) (bool, error) {
	logger := logging.FromContext(ctx)
	imageInfos, err := sip.imageService.ListImages(ctx)
	if err != nil {
		logger.Errorf("List images failed, err %v", err)
		return false, err
	}
	for _, info := range imageInfos {
		if info.ContainsImage(imageName, imageTag) {
			return true, nil
		}
	}
	return false, nil
}

func (sip *serialImagePuller) ImageHashExists(ctx context.Context, imageRef string) (bool, error) {
	logger := logging.FromContext(ctx)
	imageInfos, err := sip.imageService.ListImages(ctx)
	if err != nil {
		logger.Error("List images failed, err %v", err)
		return false, err
	}

	for _, info := range imageInfos {
		if info.ContainsImageHash(imageRef) {
			return true, nil
		}
	}
	return false, nil
}
