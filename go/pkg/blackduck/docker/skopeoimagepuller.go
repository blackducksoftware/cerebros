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
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	skopeoCopyStage = "skopeo copy docker image"
	skopeoGetStage  = "skopeo get docker image"
)

type SkopeoImagePuller struct {
	directory  string
	registries []RegistryAuth
}

func NewSkopeoImagePuller(directory string, registries []RegistryAuth) *SkopeoImagePuller {
	log.Infof("creating Skopeo image puller")
	return &SkopeoImagePuller{directory: directory, registries: registries}
}

func (ip *SkopeoImagePuller) PullImage(pullSpec string) (*PullResult, error) {
	start := time.Now()
	img := image{PullSpec: pullSpec}
	log.Infof("Processing image: %s in %s", img.PullSpec, img.tarFilePath(ip.directory))

	err := ip.SaveImageToTar(img)
	if err != nil {
		return nil, errors.WithMessagef(err, "unable to save image %s to tar file", img.PullSpec)
	}

	recordDockerTotalDuration(time.Now().Sub(start))

	log.Infof("Ready to scan image %s at path %s", img.PullSpec, img.tarFilePath(ip.directory))
	inspect, err := ip.inspect(img.PullSpec)
	if err != nil {
		return nil, err
	}

	var tag string
	if len(inspect.RepoTags) > 0 {
		tag = inspect.RepoTags[0]
	}
	pr := &PullResult{
		Path: img.tarFilePath(ip.directory),
		Repo: inspect.Name,
		Tag:  tag,
		Sha:  inspect.Digest, // TODO probably have to peel off the sha256: prefix ?
	}
	return pr, nil
}

type SkopeoInspectResponse struct {
	Name         string
	Digest       string
	RepoTags     []string
	Architecture string
	Os           string
	Layers       []string
}

func (ip *SkopeoImagePuller) inspect(pullSpec string) (*SkopeoInspectResponse, error) {
	cmd := exec.Command("skopeo",
		"inspect",
		fmt.Sprintf("docker://%s", pullSpec))
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.Wrapf(err, "unable to run <%s>", cmd.String())
	}

	response := &SkopeoInspectResponse{}
	err = json.Unmarshal(stdoutStderr, response)
	if err != nil {
		errors.Wrapf(err, "unable to unmarshal output from <%s>", cmd.String())
	}
	return response, nil
}

// TODO not sure why this method is here, maybe remove it from the interface and delete it here too?
//func (ip *SkopeoImagePuller) CreateImageInLocalDocker(img image) error {
//	start := time.Now()
//	dockerPullSpec := img.PullSpec
//	log.Infof("Attempting to create %s ......", dockerPullSpec)
//
//	authHeader := ip.needAuthHeader(img)
//	var headerValue string
//	if strings.Compare(authHeader, "") != 0 {
//		headerValue = fmt.Sprintf("--src-creds=%s", authHeader)
//	}
//
//	var cmd *exec.Cmd
//	if len(headerValue) > 0 {
//		cmd = exec.Command("skopeo",
//			"--insecure-policy",
//			"--tls-verify=false",
//			"copy",
//			headerValue,
//			fmt.Sprintf("docker://%s", dockerPullSpec),
//			fmt.Sprintf("docker-daemon:%s", dockerPullSpec))
//	} else {
//		cmd = exec.Command("skopeo",
//			"--insecure-policy",
//			"--tls-verify=false",
//			"copy",
//			fmt.Sprintf("docker://%s", dockerPullSpec),
//			fmt.Sprintf("docker-daemon:%s", dockerPullSpec))
//	}
//
//	log.Infof("running skopeo copy command %+v", cmd)
//	stdoutStderr, err := cmd.CombinedOutput()
//
//	if err != nil {
//		recordDockerError(skopeoCopyStage, "skopeo copy failed", img, err)
//		log.Errorf("skopeo copy command failed for %s with error %s and output:\n%s\n", dockerPullSpec, err.Error(), string(stdoutStderr))
//		return errors.Wrapf(err, "Create failed for image %s", dockerPullSpec)
//	}
//
//	recordDockerCreateDuration(time.Now().Sub(start))
//
//	err = ip.recordTarFileSize(img)
//
//	return err
//}

func (ip *SkopeoImagePuller) SaveImageToTar(img image) error {
	start := time.Now()
	dockerPullSpec := img.PullSpec
	log.Infof("Attempting to create %s ......", dockerPullSpec)

	authHeader := ip.needAuthHeader(img)
	var headerValue string
	if strings.Compare(authHeader, "") != 0 {
		headerValue = fmt.Sprintf("--src-creds=%s", authHeader)
	}

	tarFilePath := img.tarFilePath(ip.directory)

	var cmd *exec.Cmd
	if len(headerValue) > 0 {
		cmd = exec.Command("skopeo",
			"--insecure-policy",
			"--tls-verify=false",
			"copy",
			headerValue,
			fmt.Sprintf("docker://%s", dockerPullSpec),
			fmt.Sprintf("docker-archive:%s", tarFilePath))
	} else {
		cmd = exec.Command("skopeo",
			"--insecure-policy",
			"--tls-verify=false",
			"copy",
			fmt.Sprintf("docker://%s", dockerPullSpec),
			fmt.Sprintf("docker-archive:%s", tarFilePath))
	}

	log.Infof("running skopeo copy command %+v", cmd)

	stdoutStderr, err := cmd.CombinedOutput()

	if err != nil {
		recordDockerError(skopeoCopyStage, "skopeo copy failed", img, err)
		log.Errorf("skopeo copy command failed for %s with error: %s, stdouterr: %s", dockerPullSpec, err.Error(), string(stdoutStderr))
		return errors.Wrapf(err, "Create failed for image %s", dockerPullSpec)
	}

	recordDockerGetDuration(time.Now().Sub(start))

	err = ip.recordTarFileSize(img)

	return err
}

// needAuthHeader will determine whether the secured registry credentials to be passed to the skopeo client for docker pull
func (ip *SkopeoImagePuller) needAuthHeader(img image) string {
	var headerValue string
	if registryAuth := needsAuthHeader(img.PullSpec, ip.registries); registryAuth != nil {
		headerValue = fmt.Sprintf("%s:%s", registryAuth.User, registryAuth.Password)

		recordEvent("add auth header")
		log.Debugf("adding auth header for %s", img.PullSpec)

		// // the -n prevents echo from appending a newline
		// fmt.Printf("XRA=`echo -n \"{ \\\"username\\\": \\\"%s\\\", \\\"password\\\": \\\"%s\\\" }\" | base64 --wrap=0`\n", ip.dockerUser, ip.dockerPassword)
		// fmt.Printf("curl -i --unix-socket /var/run/docker.sock -X POST -d \"\" -H \"X-Registry-Auth: %s\" %s\n", headerValue, imageURL)
	} else {
		recordEvent("omit auth header")
		log.Debugf("omitting auth header for %s", img.PullSpec)
	}
	return headerValue
}

// recordTarFileSize will record the TAR file size
func (ip *SkopeoImagePuller) recordTarFileSize(img image) error {
	// What's the right way to get the size of the file?
	//  1. resp.ContentLength
	//  2. check the size of the file after it's written
	// fileSizeInMBs := int(resp.ContentLength / (1024 * 1024))
	stats, err := os.Stat(img.tarFilePath(ip.directory))

	if err != nil {
		recordDockerError(skopeoGetStage, "unable to get tar file stats", img, err)
		return errors.Wrapf(err, "unable to get tar file stats from %s", img.tarFilePath(ip.directory))
	}

	fileSizeInMBs := int(stats.Size() / (1024 * 1024))
	recordTarFileSize(fileSizeInMBs)
	return nil
}
