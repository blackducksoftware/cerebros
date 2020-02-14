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
package jobrunner

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
)

var polarisUserTimeout = "POLARIS_USER_INPUT_TIMEOUT_MINUTES"
var polarisServerURL = "POLARIS_SERVER_URL"
var polarisAccessToken = "POLARIS_ACCESS_TOKEN"

type PolarisConfig struct {
	PolarisURL   string
	PolarisToken string
}

type PolarisScanner struct {
	config PolarisConfig
	pathToPolarisCLI string
}

func NewPolarisScanner(pathToPolarisCLI string, config PolarisConfig) (*PolarisScanner, error) {
	err := os.Setenv(polarisUserTimeout, "1")
	if err != nil {
		return nil, errors.Wrapf(err, "unable to setenv for %s", polarisUserTimeout)
	}
	ps := &PolarisScanner{config: config, pathToPolarisCLI: pathToPolarisCLI}
	if err := ps.configurePolarisCliWithAccessToken(); err != nil {
		return nil, errors.Wrapf(err, "unable to configure polaris cli with access token")
	}
	return ps, nil
}

func (ps *PolarisScanner) Capture(capturePath string) (string, error) {
	log.Infof("Running Polaris Capture")
	if output, err := execSh("polaris capture", capturePath); err != nil {
		return "", errors.Wrapf(err,"unable to run polaris capture: %s", output)
	}

	log.Infof("Deleting polaris.yml")
	if err := os.Remove(path.Join(capturePath, "polaris.yml")); err != nil {
		return "", errors.Wrapf(err, "unable to remove polaris.yml")
	}

	return path.Join(capturePath, ".synopsys/polaris/data/coverity/2019.06-5/idir"), nil
}

func (ps *PolarisScanner) Scan(repoPath, idirPath string) error {
	shellCmd := "polaris setup"
	log.Infof("attempting to exec %s in %s", shellCmd, repoPath)
	if output, err := execSh(shellCmd, repoPath); err != nil {
		return errors.Wrapf(err, "unable to exec %s in %s: %s", shellCmd, repoPath, output)
	}

	polarisYmlPath := path.Join(repoPath, "polaris.yml")
	content, err := ioutil.ReadFile(polarisYmlPath)
	if err != nil {
		return errors.Wrapf(err, "unable to read file %s", polarisYmlPath)
	}

	CoverityConfig := make(map[string]interface{})
	if err := yaml.Unmarshal(content, CoverityConfig); err != nil {
		return errors.Wrapf(err, "unable to unmarshal yaml for coverity config")
	}
	// TODO - Need to get the CoverityConfig struct from somewhere :/
	CoverityConfig["capture"] = map[string]interface{}{
		"coverity": map[string]interface{}{
			"idirCapture": map[string]interface{}{
				"idirPath": idirPath,
			},
		},
	}
	CoverityConfigYaml, err := yaml.Marshal(CoverityConfig)
	if err != nil {
		return errors.Wrapf(err,"unable to marshal yaml for coverity config")
	}

	if err := ioutil.WriteFile(polarisYmlPath, CoverityConfigYaml, os.FileMode(700)); err != nil {
		return errors.Wrapf(err, "unable to write file %s", polarisYmlPath)
	}

	analyzeCmd := "polaris analyze"
	if output, err := execSh(analyzeCmd, repoPath); err != nil {
		return errors.Wrapf(err, "unable to exec %s in %s: %s", analyzeCmd, repoPath, output)
	}

	return nil
}

func execSh(shellCmd, workdir string) (string, error) {
	execCmd := exec.Command("sh", "-c", shellCmd)
	if len(workdir) > 0 {
		execCmd.Dir = path.Clean(workdir)
	}

	output, err := execCmd.CombinedOutput()
	if err != nil {
		return string(output), errors.Wrapf(errors.WithStack(err), "command failed: %s", output)
	}

	return string(output), nil
}

func (ps *PolarisScanner)configurePolarisCliWithAccessToken() error {
	err := os.Setenv(polarisServerURL, ps.config.PolarisURL)
	if err != nil {
		return errors.Wrapf(err, "unable to set env var %s", polarisServerURL)
	}
	err = os.Setenv(polarisAccessToken, ps.config.PolarisToken)
	if err != nil {
		return errors.Wrapf(err, "unable to set env var %s", polarisAccessToken)
	}

	cmd := "polaris configure"
	if output, err := execSh(cmd, ps.pathToPolarisCLI); err != nil {
		log.Errorf("error Output: %s\n", output)
		return errors.Wrapf(err, "unable to run shell cmd: %s", output)
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

//func downloadPolarisCli(envURL string, authHeader map[string]string) (string, error) {
//	// TODO change this to linux
//	apiURL := fmt.Sprintf("%s/api/tools/polaris_cli-linux64.zip", envURL)
//	req, err := http.NewRequest("GET", apiURL, nil)
//	if err != nil {
//		return "", err
//	}
//	for key, val := range authHeader {
//		req.Header.Set(key, val)
//	}
//
//	client := http.Client{
//		Timeout: time.Second * 10,
//	}
//
//	resp, err := client.Do(req)
//	if err != nil {
//		return "", err
//	}
//
//	respBody, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		return "", err
//	}
//
//	zipReader, err := zip.NewReader(bytes.NewReader(respBody), int64(len(respBody)))
//	if err != nil {
//		return "", err
//	}
//
//	subDir := ""
//
//	for _, f := range zipReader.File {
//		if f.FileInfo().IsDir() {
//			// Make Folder
//			os.MkdirAll(f.Name, os.ModePerm)
//			subDir = f.Name
//			continue
//		}
//		// Make File
//		if err = os.MkdirAll(filepath.Dir(f.Name), os.ModePerm); err != nil {
//			return "", err
//		}
//		outFile, err := os.OpenFile(f.Name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
//		if err != nil {
//			return "", err
//		}
//		rc, err := f.Open()
//		if err != nil {
//			return "", err
//		}
//		_, err = io.Copy(outFile, rc)
//		outFile.Close()
//		rc.Close()
//	}
//
//	resp.Body.Close()
//
//	srcDir := cleanPath(subDir)
//
//	cmd := fmt.Sprintf("chmod +x %s/polaris*", srcDir[1:])
//	if output, err := execSh(cmd); err != nil {
//		return "", fmt.Errorf("%s - %s", output, err)
//	}
//
//	absolutePath, err := os.Getwd()
//	if err != nil {
//		return "", err
//	}
//	polarisCliPath := fmt.Sprintf("%s%s", absolutePath, srcDir)
//
//	fmt.Printf("polarisCliPath: %s\n", polarisCliPath)
//
//	return polarisCliPath, nil
//}

/*func authenticateUserAndGetCookie(envURL, emailID, password string) (string, error) {
	apiURL := fmt.Sprintf("%s/api/auth/authenticate", envURL)
	form := url.Values{}
	form.Add("email", emailID)
	form.Add("password", password)

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create NewRequest: %s", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("POST request failed: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return resp.Header["Set-Cookie"][0], nil
	}
	return "", fmt.Errorf("bad status code: %+v - %+v - %+v", resp.StatusCode, resp.Status, resp.Body)
}
*/

//
//func scanRepoWithCoverityCli(covBuildCliPath, repoPath, buildToolCommand string) error {
//	// example: /Users/hammer/.synopsys/polaris/coverity-tools-macosx-2019.03/bin/cov-configure --java
//	cmd := fmt.Sprintf("cd %s;%s/cov-configure --java", repoPath, covBuildCliPath)
//	if output, err := execSh(cmd); err != nil {
//		fmt.Printf("error out: %s\n", output)
//		return err
//	}
//	// /Users/hammer/.synopsys/polaris/coverity-tools-macosx-2019.03/bin/cov-build --dir myidir /Users/hammer/go/src/github.com/blackducksoftware/hub-fortify-parser/gradlew
//	cmd = fmt.Sprintf("cd %s;%s/cov-build --dir myidirs %s", repoPath, covBuildCliPath, buildToolCommand)
//	if output, err := execSh(cmd); err != nil {
//		fmt.Printf("error out: %s\n", output)
//		return err
//	}
//	return nil
//}
