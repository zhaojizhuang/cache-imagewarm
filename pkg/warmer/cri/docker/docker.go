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

package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	dockertypes "github.com/docker/docker/api/types"
	dockerapi "github.com/docker/docker/client"
	dockermessage "github.com/docker/docker/pkg/jsonmessage"
	v1 "k8s.io/api/core/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"knative.dev/pkg/logging"

	"knative.dev/cache-imagewarm/pkg/warmer/cri"
	"knative.dev/cache-imagewarm/pkg/warmer/utils"
)

// NewDockerImageService create a docker runtime
func NewDockerImageService(runtimeURI string) (cri.ImageService, error) {
	r := &dockerImageService{runtimeURI: runtimeURI}
	if err := r.createRuntimeClientIfNecessary(); err != nil {
		return nil, err
	}
	r.imagePullProgressDeadline = 5 * time.Minute
	return r, nil
}

type dockerImageService struct {
	sync.Mutex
	runtimeURI string
	//accountManager utils.ImagePullAccountManager

	client *dockerapi.Client

	// timeout is the timeout of short running docker operations.
	timeout time.Duration
	// If no pulling progress is made before imagePullProgressDeadline, the image pulling will be cancelled.
	// Docker reports image progress for every 512kB block, so normally there shouldn't be too long interval
	// between progress updates.
	imagePullProgressDeadline time.Duration
}

// getTimeoutContext returns a new context with default request timeout
func (d *dockerImageService) getTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), d.timeout)
}

// operationTimeout is the error returned when the docker operations are timeout.
type operationTimeout struct {
	err error
}

func (e operationTimeout) Error() string {
	return fmt.Sprintf("operation timeout: %v", e.err)
}

// contextError checks the context, and returns error if the context is timeout.
func contextError(ctx context.Context) error {
	if ctx.Err() == context.DeadlineExceeded {
		return operationTimeout{err: ctx.Err()}
	}
	return ctx.Err()
}

func (d *dockerImageService) RemoveImage(imageRef string) error {
	ctx, cancel := d.getTimeoutContext()
	defer cancel()

	_, err := d.client.ImageRemove(ctx, imageRef, dockertypes.ImageRemoveOptions{Force: true, PruneChildren: true})
	if ctxErr := contextError(ctx); ctxErr != nil {
		return ctxErr
	}
	if dockerapi.IsErrNotFound(err) {
		return fmt.Errorf("no such image: %s", imageRef)
	}
	return err
}

func (d *dockerImageService) createRuntimeClientIfNecessary() error {
	d.Lock()
	defer d.Unlock()
	if d.client != nil {
		return nil
	}
	c, err := dockerapi.NewClientWithOpts(dockerapi.WithVersion("1.19"))
	if err != nil {
		return err
	}
	d.client = c
	return nil
}

// getCancelableContext returns a new cancelable context. For long running requests without timeout, we use cancelable
// context to avoid potential resource leak, although the current implementation shouldn't leak resource.
func (d *dockerImageService) getCancelableContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithCancel(ctx)
}

func (d *dockerImageService) PullImage(ctx context.Context, imageRef string, pullSecret *v1.Secret) (err error) {
	logger := logging.FromContext(ctx)

	ctx, cancel := d.getCancelableContext(ctx)
	defer cancel()
	if err = d.createRuntimeClientIfNecessary(); err != nil {
		return err
	}

	logger.Infof("Docker image service is starting to pull image :%s ", imageRef)

	resp, err := d.doPullImage(ctx, imageRef, pullSecret)
	if err != nil {
		return err
	}
	if resp != nil {
		defer resp.Close()
	}
	reporter := newProgressReporter(ctx, imageRef, cancel, d.imagePullProgressDeadline)
	reporter.start()
	defer reporter.stop()
	decoder := json.NewDecoder(resp)
	for {
		var msg dockermessage.JSONMessage
		err := decoder.Decode(&msg)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if msg.Error != nil {
			return msg.Error
		}
		reporter.set(&msg)
	}
	return nil
}
func (d *dockerImageService) doPullImage(ctx context.Context, imageRef string, pullSecret *v1.Secret) (resp io.ReadCloser, err error) {
	logger := logging.FromContext(ctx)
	registry := utils.ParseRegistry(imageRef)

	if pullSecret == nil {
		// Anonymous pull
		logger.Infof("Pull image %s anonymous", imageRef)
		resp, err = d.client.ImagePull(ctx, imageRef, dockertypes.ImagePullOptions{})

		return resp, err

	}

	var authInfos []utils.AuthInfo
	authInfos, err = cri.ConvertToRegistryAuths(*pullSecret, registry)
	if err == nil {
		var pullErrs []error
		for _, authInfo := range authInfos {
			var pullErr error
			logger.Infof("Pull image :%v with user %v", imageRef, authInfo.Username)
			resp, pullErr = d.client.ImagePull(ctx, imageRef, dockertypes.ImagePullOptions{RegistryAuth: authInfo.EncodeToString()})
			if pullErr == nil {
				return resp, nil
			}

			logger.Errorw("Failed to pull image :%v with user %v, err %v", imageRef, authInfo.Username, pullErr)
			pullErrs = append(pullErrs, pullErr)
		}
		if len(pullErrs) > 0 {
			err = utilerrors.NewAggregate(pullErrs)
		}
	}

	return resp, err
}

func (d *dockerImageService) ListImages(ctx context.Context) ([]cri.ImageInfo, error) {
	if err := d.createRuntimeClientIfNecessary(); err != nil {
		return nil, err
	}
	infos, err := d.client.ImageList(ctx, dockertypes.ImageListOptions{All: true})
	if err != nil {
		//d.handleRuntimeError(err)
		return nil, err
	}
	return newImageCollectionDocker(infos), nil
}

func newImageCollectionDocker(infos []dockertypes.ImageSummary) []cri.ImageInfo {
	collection := make([]cri.ImageInfo, 0, len(infos))
	for _, info := range infos {
		collection = append(collection, cri.ImageInfo{
			ID:          info.ID,
			RepoTags:    info.RepoTags,
			RepoDigests: info.RepoDigests,
			Size:        info.Size,
		})
	}
	return collection
}
