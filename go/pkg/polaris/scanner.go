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
package polaris

import (
	"github.com/blackducksoftware/cerebros/go/pkg/util"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path"
	"time"
)

var polarisUserTimeout = "POLARIS_USER_INPUT_TIMEOUT_MINUTES"
var polarisServerURL = "POLARIS_SERVER_URL"
var polarisAccessToken = "POLARIS_ACCESS_TOKEN"

var analyzeTimeout = 15 * time.Minute
var captureTimeout = 15 * time.Minute
var configureTimeout = 30 * time.Second
var setupTimeout = 15 * time.Minute

type Scanner struct {
	URL     string
	Token   string
	CLIPath string
}

func NewScanner(cliPath string, url string, token string) (*Scanner, error) {
	err := os.Setenv(polarisUserTimeout, "1")
	if err != nil {
		return nil, errors.WithMessagef(err, "unable to setenv for %s", polarisUserTimeout)
	}
	ps := &Scanner{
		URL:     url,
		Token:   token,
		CLIPath: cliPath,
	}
	if err := ps.configurePolarisCliWithAccessToken(); err != nil {
		return nil, errors.WithMessagef(err, "unable to configure polaris cli with access token")
	}
	return ps, nil
}

func (ps *Scanner) Capture(capturePath string) (string, error) {
	log.Infof("Running Polaris Capture on path %s", capturePath)
	shellCmd := ps.polarisBinaryPath() + " capture"
	start := time.Now()
	err := util.ExecShellWithTimeout(shellCmd, capturePath, captureTimeout)
	recordEventTime("polaris_capture", time.Now().Sub(start))
	recordEvent("polaris_capture", err)
	if err != nil {
		return "", errors.WithMessagef(err, "unable to run polaris capture")
	}

	log.Infof("Deleting polaris.yml")
	err = os.Remove(path.Join(capturePath, "polaris.yml"))
	recordEvent("delete_polaris_yml", err)
	if err != nil {
		return "", errors.Wrapf(err, "unable to remove polaris.yml")
	}

	coverityDirPath := path.Join(capturePath, ".synopsys/polaris/data/coverity")
	filenames, err := ioutil.ReadDir(coverityDirPath)
	if err != nil {
		return "", errors.Wrapf(err, "unable to read directory in order to find idir path")
	}
	if len(filenames) != 1 || !filenames[0].IsDir() {
		return "", errors.Wrapf(err, "expected 1 directory in %s, found %d files/dirs", coverityDirPath, len(filenames))
	}
	return path.Join(coverityDirPath, filenames[0].Name(), "idir"), nil
}

func (ps *Scanner) polarisBinaryPath() string {
	return path.Join(ps.CLIPath, "polaris")
}

func (ps *Scanner) CaptureAndScan(capturePath string, useLocalAnalysis bool) error {
	idirPath, err := ps.Capture(capturePath)
	if err != nil {
		return errors.WithMessagef(err, "unable to capture path %s", capturePath)
	}
	return ps.Scan(capturePath, idirPath, useLocalAnalysis)
}

func (ps *Scanner) Scan(repoPath string, idirPath string, useLocalAnalysis bool) error {
	shellCmd := ps.polarisBinaryPath() + " setup"
	log.Infof("attempting to exec %s in %s", shellCmd, repoPath)
	start := time.Now()
	err := util.ExecShellWithTimeout(shellCmd, repoPath, setupTimeout)
	recordEventTime("polaris_setup", time.Now().Sub(start))
	recordEvent("polaris_setup", err)
	if err != nil {
		return err
	}

	polarisYmlPath := path.Join(repoPath, "polaris.yml")
	content, err := ioutil.ReadFile(polarisYmlPath)
	recordEvent("read_polaris_yaml", err)
	if err != nil {
		return errors.Wrapf(err, "unable to read file %s", polarisYmlPath)
	}

	coverityConfig := make(map[string]interface{})
	err = yaml.Unmarshal(content, coverityConfig)
	recordEvent("unmarshal_coverity_yaml_config", err)
	if err != nil {
		return errors.Wrapf(err, "unable to unmarshal yaml for coverity config")
	}
	// TODO - Need to get the coverityConfig struct from somewhere :/
	coverityConfig["capture"] = map[string]interface{}{
		"coverity": map[string]interface{}{
			"idirCapture": map[string]interface{}{
				"idirPath": idirPath,
			},
		},
	}
	if useLocalAnalysis {
		coverityConfig["analyze"] = map[string]interface{}{
			"mode": "local",
		}
	}
	coverityConfigYaml, err := yaml.Marshal(coverityConfig)
	log.Debugf("using coverity yaml config: \n%s\n", coverityConfigYaml)
	recordEvent("marshal_coverity_yaml_config", err)
	if err != nil {
		return errors.Wrapf(err, "unable to marshal yaml for coverity config")
	}

	err = ioutil.WriteFile(polarisYmlPath, coverityConfigYaml, os.FileMode(700))
	recordEvent("write_coverity_config", err)
	if err != nil {
		return errors.Wrapf(err, "unable to write file %s", polarisYmlPath)
	}

	var analyzeCmd string
	if useLocalAnalysis {
		analyzeCmd = ps.polarisBinaryPath() + " analyze -w"
	} else {
		analyzeCmd = ps.polarisBinaryPath() + " analyze"
	}
	analyzeStart := time.Now()
	err = util.ExecShellWithTimeout(analyzeCmd, repoPath, analyzeTimeout)
	recordEventTime("polaris_analyze", time.Now().Sub(analyzeStart))
	recordEvent("polaris_analyze", err)
	return err
}

func (ps *Scanner) configurePolarisCliWithAccessToken() error {
	err := os.Setenv(polarisServerURL, ps.URL)
	if err != nil {
		return errors.WithMessagef(err, "unable to set env var %s", polarisServerURL)
	}
	err = os.Setenv(polarisAccessToken, ps.Token)
	if err != nil {
		return errors.WithMessagef(err, "unable to set env var %s", polarisAccessToken)
	}

	start := time.Now()
	cmd := ps.polarisBinaryPath() + " configure"
	err = util.ExecShellWithTimeout(cmd, ps.CLIPath, configureTimeout)
	recordEventTime("polaris_configure", time.Now().Sub(start))
	recordEvent("polaris_configure", err)
	if err != nil {
		log.Errorf("failed to run polaris configure: %s", err)
		return err
	}

	return nil
}

//func captureBuildAndPushToBucket(fromBucket, fromBucketPath, polarisCliPath, bucketName, pathToBucketServiceAccount, storagePathInBucket string) error {
//	// Create a temporary directpry
//	// TODO change this. Use emptyDir
//	tmpDir, err := ioutil.TempDir("/tmp", "scan")
//	if err != nil {
//		return err
//	}
//	defer os.RemoveAll(tmpDir)
//
//	fmt.Printf("Downloading from GS Bucket\n")
//
//	// Download the zip archive
//	tmpFile := path.Join(tmpDir, "master.zip")
//	if err := copyFromGSBucket(fromBucket, pathToBucketServiceAccount, fromBucketPath, tmpFile); err != nil {
//		return err
//	}
//
//	// Unzip then remove
//	if err := util.Unzip(tmpFile, tmpDir); err != nil {
//		return err
//	}
//
//	fmt.Printf("Running Polaris Capture\n")
//	if err := polarisCliCapture(tmpDir, polarisCliPath); err != nil {
//		return err
//	}
//
//	pathToIdir := fmt.Sprintf("%s%s", tmpDir, cleanPath(".synopsys/polaris/data/coverity/2019.06-5/idir"))
//	pathToIdirZip := filepath.Join(tmpDir, "idir.zip")
//	fmt.Printf("Path to IDIR: %s\n", pathToIdir)
//
//	if err := util.Zipit(pathToIdir, pathToIdirZip); err != nil {
//		return err
//	}
//
//	fmt.Printf("Copying to GS Bucket\n")
//	if err := copyToGSBucket(bucketName, pathToBucketServiceAccount, storagePathInBucket, pathToIdirZip); err != nil {
//		return err
//	}
//	fmt.Printf("Scan Complete\n")
//	return nil
//}

//
//func scanRepoWithCoverityCli(covBuildCliPath, repoPath, buildToolCommand string) error {
//	// example: /Users/hammer/.synopsys/polaris/coverity-tools-macosx-2019.03/bin/cov-configure --java
//	cmd := fmt.Sprintf("cd %s;%s/cov-configure --java", repoPath, covBuildCliPath)
//	if output, err := util.ExecShellWithTimeout(cmd); err != nil {
//		fmt.Printf("error out: %s\n", output)
//		return err
//	}
//	// /Users/hammer/.synopsys/polaris/coverity-tools-macosx-2019.03/bin/cov-build --dir myidir /Users/hammer/go/src/github.com/blackducksoftware/hub-fortify-parser/gradlew
//	cmd = fmt.Sprintf("cd %s;%s/cov-build --dir myidirs %s", repoPath, covBuildCliPath, buildToolCommand)
//	if output, err := util.ExecShellWithTimeout(cmd); err != nil {
//		fmt.Printf("error out: %s\n", output)
//		return err
//	}
//	return nil
//}
