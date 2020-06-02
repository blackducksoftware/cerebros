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
package api

import (
	"encoding/json"
	"fmt"
	"github.com/blackducksoftware/cerebros/go/pkg/util"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"path"
	"time"
)

type OSType string

// ...
const (
	OSTypeMac     OSType = "Mac"
	OSTypeLinux   OSType = "Linux"
	OSTypeWindows OSType = "Windows"
)

func (o OSType) String() string {
	switch o {
	case OSTypeMac:
		return "OSTypeMac"
	case OSTypeLinux:
		return "OSTypeLinux"
	case OSTypeWindows:
		return "OSTypeWindows"
	}
	panic(fmt.Errorf("invalid OSType value: %s", string(o)))
}

func (o OSType) PlatformType() string {
	switch o {
	case OSTypeMac:
		return "macosx"
	case OSTypeLinux:
		return "linux64"
	case OSTypeWindows:
		return "win64"
	}
	panic(fmt.Errorf("invalid OSType value: %s", string(o)))
}

func ParseOSType(osType string) (OSType, error) {
	switch osType {
	case "mac", "Mac":
		return OSTypeMac, nil
	case "linux", "Linux":
		return OSTypeLinux, nil
	case "windows", "Windows":
		return OSTypeWindows, nil
	}
	return "", errors.New(fmt.Sprintf("invalid ostype: %s", osType))
}

func (o *OSType) UnmarshalJSON(data []byte) error {
	var str string
	err := json.Unmarshal(data, &str)
	if err != nil {
		return err
	}
	osType, err := ParseOSType(str)
	if err != nil {
		return err
	}
	*o = osType
	return nil
}

func (o *OSType) UnmarshalText(text []byte) (err error) {
	osType, err := ParseOSType(string(text))
	if err != nil {
		return err
	}
	*o = osType
	return nil
}

func (o OSType) PolarisCLIUrl(host string) string {
	// example: //https://onprem-perf.dev.polaris.synopsys.com/api/tools/polaris_cli-win64.zip
	return fmt.Sprintf("%s/api/tools/polaris_cli-%s.zip", host, o.PlatformType())
}

func (client *Client) DownloadCli(dir string, osType OSType) (string, error) {
	url := osType.PolarisCLIUrl(client.URL)

	filename := fmt.Sprintf("polaris-cli-%s", osType.PlatformType())
	downloadPath := path.Join(dir, filename) + ".zip"
	log.Infof("fetching polaris cli at %s to file %s", url, downloadPath)
	restyClient := resty.New().SetTimeout(5 * time.Minute)

	// download
	resp, err := restyClient.R().SetOutput(downloadPath).Get(url)
	log.Debugf("response from resty: %s", resp.String())
	if err != nil {
		return "", errors.Wrapf(err, "http GET request to %s failed", url)
	}

	// unzip
	filenames, err := util.Unzip(downloadPath, dir)
	if err != nil {
		return "", errors.WithMessagef(err, "unable to unzip %s", downloadPath)
	}

	if len(filenames) == 0 {
		return "", errors.New(fmt.Sprintf("expected at least 1 unzipped file, found 0"))
	}

	cliDir := filenames[0]
	log.Warnf("assuming cli downloaded to %s, HOWEVER THIS IS NOT CERTAIN!", cliDir)
	return cliDir, nil
}

func getPolarisCliDirPath(baseDir string) (string, error) {
	log.Debugf("looking for polaris cli dir in %s", baseDir)
	files, err := ioutil.ReadDir(baseDir)
	if err != nil {
		return "", errors.Wrapf(err, "unable to read dir %s", baseDir)
	}
	if len(files) != 1 {
		return "", errors.New(fmt.Sprintf("expected 1 directory in %s, found %d files/dirs", baseDir, len(files)))
	}
	file := files[0]
	if !file.IsDir() {
		return "", errors.New(fmt.Sprintf("expected 1 directory in %s, found file %s", baseDir, file.Name()))
	}
	return file.Name(), nil
}
