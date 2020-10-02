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
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	dockerSocketPath = "/var/run/docker.sock"

	createStage = "create docker image"
	getStage    = "get docker image"
)

type ImagePuller struct {
	directory  string
	client     *http.Client
	registries []RegistryAuth
}

func NewImagePuller(directory string, registries []RegistryAuth) *ImagePuller {
	fd := func(proto, addr string) (conn net.Conn, err error) {
		return net.Dial("unix", dockerSocketPath)
	}
	tr := &http.Transport{Dial: fd}
	client := &http.Client{Transport: tr}
	return &ImagePuller{
		directory:  directory,
		client:     client,
		registries: registries}
}

// PullImage gives us access to a docker image by:
//   1. hitting a docker create endpoint (?)
//   2. pulling down the newly created image and saving as a tarball
// It does this by accessing the host's docker daemon, locally, over the docker
// socket.  This gives us a window into any images that are local.
func (ip *ImagePuller) PullImage(pullSpec string) (*PullResult, error) {
	start := time.Now()

	img := image{PullSpec: pullSpec}
	err := ip.CreateImageInLocalDocker(img)
	if err != nil {
		return nil, errors.WithMessagef(err, "unable to create image %s in locker docker", img.PullSpec)
	}
	log.Infof("Processing image: %s", img.PullSpec)

	err = ip.SaveImageToTar(img)
	if err != nil {
		return nil, errors.WithMessagef(err, "unable to save image %s to tar file", img.PullSpec)
	}

	recordDockerTotalDuration(time.Now().Sub(start))

	log.Infof("Ready to scan image %s at path %s", img.PullSpec, img.tarFilePath(ip.directory))

	return ip.inspectImage(img)
}

// CreateImageInLocalDocker could also be implemented using curl:
// this example hits ... ? the default registry?  docker hub?
//   curl --unix-socket /var/run/docker.sock -X POST http://localhost/images/create?fromImage=alpine
// this example hits the kipp registry:
//   curl --unix-socket /var/run/docker.sock -X POST http://localhost/images/create\?fromImage\=registry.kipp.blackducksoftware.com%2Fblackducksoftware%2Fhub-jobrunner%3A4.5.0
//
func (ip *ImagePuller) CreateImageInLocalDocker(img image) error {
	start := time.Now()
	imageURL := img.createURL()
	log.Infof("Attempting to create %s ......", imageURL)
	req, err := http.NewRequest("POST", imageURL, nil)
	if err != nil {
		recordDockerError(createStage, "unable to create POST request", img, err)
		return errors.Wrapf(err, "unable to create POST request for image %s", imageURL)
	}

	if registryAuth := needsAuthHeader(img.PullSpec, ip.registries); registryAuth != nil {
		headerValue := encodeAuthHeader(registryAuth.User, registryAuth.Password)
		// log.Infof("X-Registry-Auth value:\n%s\n", headerValue)
		req.Header.Add("X-Registry-Auth", headerValue)

		recordEvent("add auth header")
		log.Debugf("adding auth header for %s", img.PullSpec)

		// // the -n prevents echo from appending a newline
		// fmt.Printf("XRA=`echo -n \"{ \\\"username\\\": \\\"%s\\\", \\\"password\\\": \\\"%s\\\" }\" | base64 --wrap=0`\n", ip.dockerUser, ip.dockerPassword)
		// fmt.Printf("curl -i --unix-socket /var/run/docker.sock -X POST -d \"\" -H \"X-Registry-Auth: %s\" %s\n", headerValue, imageURL)
	} else {
		recordEvent("omit auth header")
		log.Debugf("omitting auth header for %s", img.PullSpec)
	}

	resp, err := ip.client.Do(req)
	if err != nil {
		recordDockerError(createStage, "POST request failed", img, err)
		return errors.Wrapf(err, "Create failed for image %s", imageURL)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		recordDockerError(createStage, "POST request failed", img, err)
		return errors.Errorf("Create may have failed for %s: status code %d, response %+v", imageURL, resp.StatusCode, resp)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		recordDockerError(createStage, "unable to read POST response body", img, err)
		log.Errorf("unable to read response body for %s: %s", imageURL, err.Error())
	}
	log.Debugf("body of POST response from %s: %s", imageURL, string(bodyBytes))

	recordDockerCreateDuration(time.Now().Sub(start))

	return err
}

type DockerInspectResponse struct {
	Id           string
	RepoTags     []string
	RepoDigests  []string
	Parent       string
	Os           string
	Architecture string
	Size         int
	VirtualSize  int
	RootFS       struct {
		Type   string
		Layers []string
	}
	// TODO lots of other stuff
}

func (ip *ImagePuller) inspectImage(img image) (*PullResult, error) {
	//start := time.Now()
	imageURL := img.inspectURL()
	log.Infof("Attempting to inspect %s ......", imageURL)
	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		//recordDockerError(inspectStage, "unable to create GET request", img, err)
		return nil, errors.Wrapf(err, "unable to create GET request for image %s", imageURL)
	}

	// TODO is this necessary?  presumably the image is already in local docker ... ?
	//if registryAuth := needsAuthHeader(img.PullSpec, ip.registries); registryAuth != nil {
	//	headerValue := encodeAuthHeader(registryAuth.User, registryAuth.Password)
	//	// log.Infof("X-Registry-Auth value:\n%s\n", headerValue)
	//	req.Header.Add("X-Registry-Auth", headerValue)
	//
	//	recordEvent("add auth header")
	//	log.Debugf("adding auth header for %s", img.PullSpec)
	//
	//	// // the -n prevents echo from appending a newline
	//	// fmt.Printf("XRA=`echo -n \"{ \\\"username\\\": \\\"%s\\\", \\\"password\\\": \\\"%s\\\" }\" | base64 --wrap=0`\n", ip.dockerUser, ip.dockerPassword)
	//	// fmt.Printf("curl -i --unix-socket /var/run/docker.sock -X POST -d \"\" -H \"X-Registry-Auth: %s\" %s\n", headerValue, imageURL)
	//} else {
	//	recordEvent("omit auth header")
	//	log.Debugf("omitting auth header for %s", img.PullSpec)
	//}

	resp, err := ip.client.Do(req)
	if err != nil {
		//recordDockerError(createStage, "GET request failed", img, err)
		return nil, errors.Wrapf(err, "Inspect failed for image %s", imageURL)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		//recordDockerError(createStage, "GET request failed", img, err)
		return nil, errors.Errorf("Inspect may have failed for %s: status code %d, response %+v", imageURL, resp.StatusCode, resp)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//recordDockerError(createStage, "unable to read GET response body", img, err)
		return nil, errors.Wrapf(err, "unable to read response body for %s: %s", imageURL, err.Error())
	}
	log.Debugf("body of GET response from %s: %s", imageURL, string(bodyBytes))

	response := &DockerInspectResponse{}
	err = json.Unmarshal(bodyBytes, response)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to parse docker inspect response")
	}

	//recordDockerCreateDuration(time.Now().Sub(start))

	if len(response.RepoDigests) == 0 {
		return nil, errors.Errorf("found 0 repo digests for %s", img.PullSpec)
	}
	repo, digest, err := parseRepoDigest(response.RepoDigests[0])
	if err != nil {
		return nil, errors.WithMessagef(err, "unable to parse repo digest from docker inspect")
	}

	if len(response.RepoTags) == 0 {
		return nil, errors.Errorf("found 0 repo tags for %s", img.PullSpec)
	}
	repo, tag, err := parseRepoTag(response.RepoTags[0])
	if err != nil {
		return nil, errors.WithMessagef(err, "unable to parse repo tag from docker inspect")
	}

	pr := &PullResult{
		Path: img.tarFilePath(ip.directory),
		Repo: repo,
		Tag:  tag,
		Sha:  digest,
	}
	return pr, nil
}

// SaveImageToTar -- part of what it does is to issue an http request similar to the following:
//   curl --unix-socket /var/run/docker.sock -X GET http://localhost/images/openshift%2Forigin-docker-registry%3Av3.6.1/get
func (ip *ImagePuller) SaveImageToTar(img image) error {
	start := time.Now()
	url := img.getURL()
	log.Infof("Making docker GET image request: %s", url)
	resp, err := ip.client.Get(url)
	if err != nil {
		recordDockerError(getStage, "GET request failed", img, err)
		return errors.Wrapf(err, "GET request to %s failed", url)
	} else if resp.StatusCode != http.StatusOK {
		err = errors.Errorf("docker GET failed: received status != 200 from %s: %s", url, resp.Status)
		recordDockerError(getStage, "GET request failed", img, err)
		return err
	}

	log.Infof("docker GET request for image %s successful", url)

	body := resp.Body
	defer func() {
		body.Close()
	}()
	tarFilePath := img.tarFilePath(ip.directory)
	log.Infof("Starting to write file contents to tar file %s", tarFilePath)

	f, err := os.OpenFile(tarFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		recordDockerError(getStage, "unable to create tar file", img, err)
		return errors.Wrapf(err, "unable to create tar file %s", tarFilePath)
	}
	if _, err = io.Copy(f, body); err != nil {
		recordDockerError(getStage, "unable to copy tar file", img, err)
		return errors.Wrapf(err, "unable to copy to tar file %s", tarFilePath)
	}

	recordDockerGetDuration(time.Now().Sub(start))

	// What's the right way to get the size of the file?
	//  1. resp.ContentLength
	//  2. check the size of the file after it's written
	// fileSizeInMBs := int(resp.ContentLength / (1024 * 1024))
	stats, err := os.Stat(tarFilePath)

	if err != nil {
		recordDockerError(getStage, "unable to get tar file stats", img, err)
		return errors.Wrapf(err, "unable to get tar file stats for %s", tarFilePath)
	}

	fileSizeInMBs := int(stats.Size() / (1024 * 1024))
	recordTarFileSize(fileSizeInMBs)

	return nil
}
