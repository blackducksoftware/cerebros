/*
Copyright (C) 2018 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownership. The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied. See the License for the
specific language governing permissions and limitations
under the License.
*/

package docker

import (
	"fmt"
	"net/url"
	"strings"
)

// ImageInterface -- is this really necessary?
type ImageInterface interface {
	DockerPullSpec() string
	DockerTarFilePath() string
}

// Image ...
type Image struct {
	Directory string
	PullSpec  string
}

// NewImage ...
func NewImage(directory string, pullSpec string) *Image {
	return &Image{Directory: directory, PullSpec: pullSpec}
}

// DockerPullSpec ...
func (image *Image) DockerPullSpec() string {
	return image.PullSpec
}

// DockerTarFilePath ...
func (image *Image) DockerTarFilePath() string {
	return fmt.Sprintf("%s/%s.tar", image.Directory, strings.Replace(image.PullSpec, "/", "_", -1))
}

func (image *Image) urlEncodedName() string {
	return url.QueryEscape(image.DockerPullSpec())
}

// createURL returns the URL used for hitting the docker daemon's create endpoint
func (image *Image) createURL() string {
	// TODO v1.24 refers to the docker version.  figure out how to avoid hard-coding this
	// TODO can probably use the docker api code for this
	return fmt.Sprintf("http://localhost/v1.24/images/create?fromImage=%s", image.urlEncodedName())
}

// getURL returns the URL used for hitting the docker daemon's get endpoint
func (image *Image) getURL() string {
	return fmt.Sprintf("http://localhost/v1.24/images/%s/get", image.urlEncodedName())
}

func (image *Image) inspectURL() string {
	return fmt.Sprintf("http://localhost/v1.24/images/%s/json", image.urlEncodedName())
}
