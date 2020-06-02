/*
Copyright (C) 2020 Synopsys, Inc.

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
package synopsys_scancli

import (
	"fmt"
	"github.com/blackducksoftware/cerebros/go/pkg/util"
	resty "github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"path"
	"strings"
	"time"
)

func GitClone(repo string) (string, error) {
	cloneDirectory, err := ioutil.TempDir("/tmp", repo)
	if err != nil {
		return "", errors.Wrapf(err, "unable to create tmp dir")
	}
	command := fmt.Sprintf("git clone git@github.com:%s %s", repo, cloneDirectory)
	output, err := util.ExecShell(command, ".", 2*time.Minute)
	if err != nil {
		log.Debugf("unable to git clone, output: %s", output)
		return "", errors.WithMessagef(err, "unable to clone repo %s to %s", repo, cloneDirectory)
	}
	return cloneDirectory, nil
}

// FetchGithubArchive returns tempDir, unzippedDir, error
// the caller is responsible for removing tempDir
func FetchGithubArchive(repo string) (string, string, error) {
	cleanedName := strings.ReplaceAll(repo, "/", "-")

	// 1. create tmp dir to put files in
	tempDir, err := ioutil.TempDir("", "gh-archive")
	recordEvent("create_temp_dir", err)
	if err != nil {
		return tempDir, "", errors.Wrapf(err, "unable to create temp dir")
	}

	// 2. download archive
	url := fmt.Sprintf("https://codeload.github.com/%s/zip/master", repo)
	downloadPath := path.Join(tempDir, strings.ReplaceAll(repo, "/", "-")) + ".zip"
	log.Infof("fetching github archive with url %s to file %s", url, downloadPath)
	restyClient := resty.New()
	restyClient.SetTimeout(1 * time.Minute)
	downloadStart := time.Now()
	resp, err := restyClient.R().SetOutput(downloadPath).Get(url)
	recordEventTime("fetch_github_archive", time.Now().Sub(downloadStart))
	recordEvent("fetch_github_archive", err)
	log.Debugf("response from resty: %s", resp.String())
	if err != nil {
		return tempDir, "", errors.Wrapf(err, "http GET request to %s failed", url)
	}

	// 3. unzip archive
	unzipPath := path.Join(tempDir, cleanedName)
	unzipStart := time.Now()
	_, err = util.Unzip(downloadPath, unzipPath)
	recordEventTime("unzip_github_archive", time.Now().Sub(unzipStart))
	recordEvent("unzip_github_archive", err)
	if err != nil {
		return tempDir, "", errors.WithMessagef(err, "unable to unzip %s for repo %s", downloadPath, repo)
	}

	// 4. done
	return tempDir, unzipPath, nil
}
