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

package cri

import (
	"context"

	v1 "k8s.io/api/core/v1"
)

const ImageClassDocker = "docker"
const ImageClassContainerd = "containerd"
const IamgeClassCriO = "crio"

type ImageInfo struct {
	// ID of an image.
	ID string `json:"Id,omitempty"`
	// repository with digest.
	RepoDigests []string `json:"RepoDigests"`
	// repository with tag.
	RepoTags []string `json:"RepoTags"`
	// size of image's taking disk space.
	Size int64 `json:"Size,omitempty"`
}

type ImageService interface {
	// PullImage pulls an image with the authentication config.
	PullImage(ctx context.Context, imageRef string, pullSecret *v1.Secret) error
	// ListImages lists the existing images.
	ListImages(ctx context.Context) ([]ImageInfo, error)
	// RemoveImage removes the image.
	RemoveImage(imageRef string) error
}

func (c ImageInfo) ContainsImage(name string, tag string) bool {
	for _, repoTag := range c.RepoTags {
		imageRepo, imageTag := ParseRepositoryTag(repoTag)
		if imageRepo == name && imageTag == tag {
			return true
		}
	}
	return false
}

func (c ImageInfo) ContainsImageHash(imageHash string) bool {
	for _, repoDigest := range c.RepoDigests {
		if imageHash == repoDigest {
			return true
		}
	}

	return false
}
