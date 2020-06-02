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
	"github.com/blackducksoftware/cerebros/go/pkg/blackduck/docker"
	"github.com/blackducksoftware/cerebros/go/pkg/blackduck/hubcli"
	"github.com/blackducksoftware/cerebros/go/pkg/polaris"
	"github.com/blackducksoftware/cerebros/go/pkg/polaris/api"
	"github.com/blackducksoftware/cerebros/go/pkg/util"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"os"
)

type Scanner struct {
	Polaris     polaris.ScannerInterface
	Blackduck   hubcli.ScanClientInterface
	ImagePuller docker.ImagePullerInterface
}

func initPolaris(config *PolarisConfig) (*polaris.Scanner, error) {
	polarisClient := api.NewClient(config.URL, config.Email, config.Password)

	err := polarisClient.Authenticate()
	if err != nil {
		return nil, errors.WithMessagef(err, "unable to authenticate")
	}
	log.Infof("successfully authenticated")

	log.Infof("about to make dir %s", config.CLIPath)
	err = os.MkdirAll(config.CLIPath, 0755)
	if err != nil {
		return nil, errors.WithMessagef(err, "unable to make dir %s for polaris cli download", config.CLIPath)
	}

	unzippedCLIPath, err := polarisClient.DownloadCli(config.CLIPath, config.OSType)
	if err != nil {
		return nil, errors.WithMessagef(err, "unable to download polaris cli")
	}

	cliPath := fmt.Sprintf("%s/bin", unzippedCLIPath)

	tokenName := fmt.Sprintf("containerized-cli-%d", rand.Int())
	scanToken, err := polarisClient.GetAccessToken(tokenName)
	if err != nil {
		return nil, errors.WithMessagef(err, "unable to get scan token")
	}

	return polaris.NewScanner(cliPath, config.URL, scanToken.Data.Attributes.AccessToken)
}

func initBlackduck(config *BlackduckConfig) (*hubcli.ScanClient, error) {
	log.Infof("instantiating Blackduck CLI with config %+v", config)
	return hubcli.NewScanClient(
		config.Host,
		config.Username,
		config.Password,
		config.Port,
		config.OSType)
}

func initImagePuller(config *ImageFacadeConfig) (docker.ImagePullerInterface, error) {
	imageDir := config.ImageDirectory
	if imageDir == "" {
		return nil, errors.New("empty image directory")
	}
	err := util.CreateIfNotExists(imageDir)
	if err != nil {
		return nil, errors.WithMessagef(err, "unable to create or set up image directory %s", imageDir)
	}

	var imagePuller docker.ImagePullerInterface
	ipType := config.ImagePullerType
	if ipType == "docker" {
		imagePuller = docker.NewImagePuller(config.ImageDirectory, config.PrivateDockerRegistries)
	} else if ipType == "skopeo" {
		imagePuller = docker.NewSkopeoImagePuller(config.ImageDirectory, config.PrivateDockerRegistries)
	} else {
		return nil, errors.New(fmt.Sprintf("invalid image puller type %s", ipType))
	}
	return imagePuller, nil
}

func NewScannerFromConfig(blackduckCfg *BlackduckConfig, polarisCfg *PolarisConfig, imagePuller *ImageFacadeConfig) (*Scanner, error) {
	var polarisScanner *polaris.Scanner
	var err error
	if polarisCfg != nil {
		polarisScanner, err = initPolaris(polarisCfg)
		if err != nil {
			return nil, errors.WithMessagef(err, "unable to instantiate polaris cli")
		}
	}

	var blackduckScanner *hubcli.ScanClient
	if blackduckCfg != nil {
		blackduckScanner, err = initBlackduck(blackduckCfg)
		if err != nil {
			return nil, errors.WithMessagef(err, "unable to instantiate blackduck cli")
		}
	}

	var imageFacade docker.ImagePullerInterface
	if imagePuller != nil {
		imageFacade, err = initImagePuller(imagePuller)
		if err != nil {
			return nil, errors.WithMessagef(err, "unable to instantiate image puller")
		}
	}

	return NewScanner(polarisScanner, blackduckScanner, imageFacade), nil
}

func NewScanner(polaris polaris.ScannerInterface, blackduck hubcli.ScanClientInterface, imagePuller docker.ImagePullerInterface) *Scanner {
	return &Scanner{
		Polaris:     polaris,
		Blackduck:   blackduck,
		ImagePuller: imagePuller,
	}
}

func (sc *Scanner) Scan(scanConfig *ScanConfig) error {
	path, err := sc.obtainScanFiles(scanConfig.CodeLocation)
	if err != nil {
		return errors.WithMessagef(err, "unable to obtain files for scanning")
	}

	scanType := scanConfig.ScanType
	if scanType.Polaris {
		return sc.Polaris.CaptureAndScan(path)
	} else if scanType.Blackduck != nil {
		return sc.Blackduck.Scan(path, scanType.Blackduck)
	} else {
		return errors.New("unrecognized scan type: everything was nil")
	}
}

func cleanUpFile(path string) {
	err := os.Remove(path)
	// TODO
	//recordCleanUpFile(err == nil)
	if err != nil {
		log.Errorf("unable to remove file %s: %s", path, err.Error())
	} else {
		log.Infof("successfully cleaned up file %s", path)
	}
}

func (sc *Scanner) obtainScanFiles(cl *CodeLocation) (string, error) {
	if cl.GitRepo != nil {
		tmpDir, dir, err := FetchGithubArchive(cl.GitRepo.Repo)
		log.Warnf("TODO at some point, we'll probably want to delete temp dir %s", tmpDir)
		//recordEvent("removeTempDir", os.RemoveAll(tmpDir))
		return dir, err
	} else if cl.FileSystem != nil {
		// blow up if files aren't there
		path := cl.FileSystem.Path
		exists, err := util.FileExists(path)
		if err != nil {
			return "", errors.WithMessagef(err, "unable to check if file %s exists", path)
		} else if !exists {
			return "", errors.New(fmt.Sprintf("file %s not found", path))
		}
		return path, nil
	} else if cl.DockerImage != nil {
		pullResult, err := sc.ImagePuller.PullImage(cl.DockerImage.PullSpec)
		if err != nil {
			return "", errors.WithMessagef(err, "unable to get docker image %s for scanning", cl.DockerImage.PullSpec)
		}
		// TODO do something with repo, digest, and tag?  like maybe use them to name the project/version/scan in blackduck?
		log.Infof("image pull result: %+v", pullResult)
		return pullResult.Path, nil
	} else if cl.None {
		return "", nil
	} else {
		return "", errors.New(fmt.Sprintf("invalid codelocation config: %+v", cl))
	}
}
