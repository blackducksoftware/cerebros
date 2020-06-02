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

package hubcli

import (
	"fmt"
	"github.com/blackducksoftware/cerebros/go/pkg/util"
	"github.com/go-resty/resty/v2"
	"os/exec"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	hubScheme   = "https"
	cliRootPath = "/tmp/blackduck-cli"
)

// ScanClient implements ScanClientInterface using
// the Black Duck hub and scan client programs.
type ScanClient struct {
	host           string
	username       string
	password       string
	port           int
	os             OSType
	scanClientInfo *ScanClientInfo
	detectPath     string
}

// NewScanClient requires hub login credentials
func NewScanClient(host string, username string, password string, port int, os OSType) (*ScanClient, error) {
	sc := ScanClient{
		host:           host,
		username:       username,
		password:       password,
		port:           port,
		os:             os,
		scanClientInfo: nil,
		detectPath:     "/tmp/detect.sh"}
	err := sc.downloadDetect()
	if err != nil {
		return nil, errors.WithMessagef(err, "unable to download detect")
	}
	return &sc, nil
}

func (sc *ScanClient) ensureScanClientIsDownloaded() error {
	if sc.scanClientInfo != nil {
		return nil
	}
	scanClientInfo, err := DownloadScanClient(
		sc.os,
		cliRootPath,
		sc.host,
		sc.username,
		sc.password,
		sc.port,
		time.Duration(300)*time.Second)
	if err != nil {
		return errors.WithMessagef(err, "unable to download scan client")
	}
	sc.scanClientInfo = scanClientInfo
	return nil
}

// Scan ...
func (sc *ScanClient) IScan(path string, projectName string, versionName string, scanName string) error {
	if err := sc.ensureScanClientIsDownloaded(); err != nil {
		return errors.WithMessagef(err, "cannot run scan cli: not downloaded")
	}
	startTotal := time.Now()

	scanCliImplJarPath := sc.scanClientInfo.ScanCliImplJarPath()
	scanCliJarPath := sc.scanClientInfo.ScanCliJarPath()
	scanCliJavaPath := sc.scanClientInfo.ScanCliJavaPath()
	cmd := exec.Command(scanCliJavaPath,
		"-Xms512m",
		"-Xmx4096m",
		"-Dblackduck.scan.cli.benice=true",
		"-Dblackduck.scan.skipUpdate=true",
		"-Done-jar.silent=true",
		"-Done-jar.jar.path="+scanCliImplJarPath,
		"-jar", scanCliJarPath,
		"--host", sc.host,
		"--port", fmt.Sprintf("%d", sc.port),
		"--scheme", hubScheme,
		"--project", projectName,
		"--release", versionName,
		"--username", sc.username,
		"--name", scanName,
		"--insecure",
		"-v",
		path)
	cmd.Env = append(cmd.Env, fmt.Sprintf("BD_HUB_PASSWORD=%s", sc.password))

	log.Infof("running command %+v for path %s\n", cmd, path)
	startScanClient := time.Now()
	stdoutStderr, err := cmd.CombinedOutput()

	recordScanClientDuration(time.Now().Sub(startScanClient), err == nil)
	recordTotalScannerDuration(time.Now().Sub(startTotal), err == nil)

	if err != nil {
		recordScannerError("scan client failed")
		log.Errorf("java scanner failed for path %s with error %s and output:\n%s\n", path, err.Error(), string(stdoutStderr))
		return errors.WithMessagef(err, "java scanner failed for path %s", path)
	}
	log.Infof("successfully completed java scanner for path %s", path)
	log.Debugf("output from path %s: %s", path, stdoutStderr)
	return nil
}

// ScanSh invokes scan.cli.sh
// example:
// 	BD_HUB_PASSWORD=??? ./bin/scan.cli.sh --host ??? --port 443 --scheme https --username sysadmin --insecure --name ??? --release ??? --project ??? ???.tar
func (sc *ScanClient) ScanSh(path string, projectName string, versionName string, scanName string) error {
	if err := sc.ensureScanClientIsDownloaded(); err != nil {
		return errors.WithMessagef(err, "cannot run scan.cli.sh: not downloaded")
	}
	startTotal := time.Now()

	cmd := exec.Command(sc.scanClientInfo.ScanCliShPath(),
		"-Xms512m",
		"-Xmx4096m",
		"-Dblackduck.scan.cli.benice=true",
		"-Dblackduck.scan.skipUpdate=true",
		"-Done-jar.silent=true",
		//		"-Done-jar.jar.path="+scanCliImplJarPath,
		//		"-jar", scanCliJarPath,
		"--host", sc.host,
		"--port", fmt.Sprintf("%d", sc.port),
		"--scheme", hubScheme,
		"--project", projectName,
		"--release", versionName,
		"--username", sc.username,
		"--name", scanName,
		"--insecure",
		"-v",
		path)
	cmd.Env = append(cmd.Env, fmt.Sprintf("BD_HUB_PASSWORD=%s", sc.password))

	log.Infof("running command %+v for path %s\n", cmd, path)
	startScanClient := time.Now()
	stdoutStderr, err := cmd.CombinedOutput()

	recordScanClientDuration(time.Now().Sub(startScanClient), err == nil)
	recordTotalScannerDuration(time.Now().Sub(startTotal), err == nil)

	if err != nil {
		recordScannerError("scan.cli.sh failed")
		log.Errorf("scan.cli.sh failed for path %s with error %s and output:\n%s\n", path, err.Error(), string(stdoutStderr))
		return errors.WithMessagef(err, "scan.cli.sh failed for path %s", path)
	}
	log.Infof("successfully completed scan.cli.sh for path %s", path)
	log.Debugf("output from path %s: %s", path, stdoutStderr)
	return nil
}

// Detect

func (sc *ScanClient) downloadDetect() error {
	url := "https://detect.synopsys.com/detect.sh"
	downloadPath := sc.detectPath
	exists, err := util.FileExists(downloadPath)
	if err != nil {
		return errors.WithMessagef(err, "unable to stat %s", downloadPath)
	}
	if exists {
		log.Infof("file %s already found, skipping download", downloadPath)
		return nil
	}
	log.Infof("about to download detect from %s to %s", url, downloadPath)
	client := resty.New()
	resp, err := client.R().SetOutput(downloadPath).Get(url)
	log.Debugf("response: %s", resp.String())
	if err != nil {
		return errors.Wrapf(err, "unable to download detect from %s", url)
	}

	cmd := exec.Command("chmod", "u+x", downloadPath)
	log.Infof("about to run %s", cmd.String())
	stdoutStderr, err := cmd.CombinedOutput()
	log.Debugf("output from %s: %s", cmd.String(), stdoutStderr)
	if err != nil {
		return errors.Wrapf(err, "unable to run %s", cmd.String())
	}

	return nil
}

// DetectOffline TODO is this method even necessary?
func (sc *ScanClient) DetectOffline(path string) error {
	cmd := exec.Command(sc.detectPath,
		"-de",
		"--logging.level.com.synopsys.integration=\"TRACE\"",
		"--blackduck.offline.mode=true",
		fmt.Sprintf("--detect.docker.image=%s", path))
	log.Infof("about to run detect offline: <%s>", cmd.String())
	stdoutStderr, err := cmd.CombinedOutput()
	log.Debugf("command output: %s", stdoutStderr)
	return errors.Wrapf(err, "unable to run %s", cmd.String())
}

func (sc *ScanClient) DetectDocker(projectName string, versionName string, scanName string, image string) error {
	return errors.New(fmt.Sprintf("unable to scan using detect tool DOCKER: not implemented (see https://raw.githubusercontent.com/blackducksoftware/blackduck-docker-inspector/master/deployment/kubernetes/setup.txt for more info on how to implement)"))
	//args := []string{
	//	"-de",
	//	"--logging.level.com.synopsys.integration=TRACE",
	//	"--detect.tools=\"DOCKER\"",
	//	"--blackduck.trust.cert=true",
	//	fmt.Sprintf("--detect.project.name=%s", projectName),
	//	fmt.Sprintf("--detect.project.version.name=%s", versionName),
	//	fmt.Sprintf("--detect.code.location.name=%s", scanName),
	//	fmt.Sprintf("--detect.docker.image=\"%s\"", image),
	//	fmt.Sprintf("--detect.docker.tar=\"%s\"", path),
	//}
	//return sc.runDetect(host, args)
}

func (sc *ScanClient) DetectSignatureScan(path string, projectName string, versionName string, scanName string) error {
	args := []string{
		"-de",
		"--logging.level.com.synopsys.integration=TRACE",
		"--detect.tools=\"SIGNATURE_SCAN\"",
		"--blackduck.trust.cert=true",
		fmt.Sprintf("--detect.project.name=%s", projectName),
		fmt.Sprintf("--detect.project.version.name=%s", versionName),
		fmt.Sprintf("--detect.code.location.name=%s", scanName),
		fmt.Sprintf("--detect.blackduck.signature.scanner.paths=\"%s\"", path),
	}
	return sc.runDetect(args)
}

func (sc *ScanClient) DetectBinaryScan(path string, projectName string, versionName string, scanName string) error {
	args := []string{
		"-de",
		"--logging.level.com.synopsys.integration=TRACE",
		"--detect.tools=\"BINARY_SCAN\"",
		"--blackduck.trust.cert=true",
		fmt.Sprintf("--detect.project.name=%s", projectName),
		fmt.Sprintf("--detect.project.version.name=%s", versionName),
		fmt.Sprintf("--detect.code.location.name=%s", scanName),
		fmt.Sprintf("--detect.binary.scan.file.path=\"%s\"", path),
	}
	return sc.runDetect(args)
}

//func (sc *ScanClient) Detect(tools []string, host string, path string, projectName string, versionName string, scanName string, image string) error {
//	toolString := strings.Join(tools, ",")
//	args := []string{
//		"-de",
//		"--logging.level.com.synopsys.integration=TRACE",
//		fmt.Sprintf("--detect.tools=\"%s\"", toolString),
//		fmt.Sprintf("--blackduck.url=https://%s", host),
//		//"--blackduck.api.token=\"XXX\"",
//		fmt.Sprintf("--blackduck.password=%s", sc.password),
//		fmt.Sprintf("--blackduck.username=%s", sc.username),
//		"--blackduck.trust.cert=true",
//		fmt.Sprintf("--detect.project.name=%s", projectName),
//		fmt.Sprintf("--detect.project.version.name=%s", versionName),
//		fmt.Sprintf("--detect.code.location.name=%s", scanName),
//		fmt.Sprintf("--detect.blackduck.signature.scanner.paths=%s", path),
//		fmt.Sprintf("--detect.binary.scan.file.path=%s", path)}
//	if image != "" {
//		args = append(args, fmt.Sprintf("--detect.docker.image=\"%s\"", image))
//	}
//	cmd := exec.Command(sc.detectPath, args...)
//	log.Infof("about to run Detect with invocation: <%s>", cmd.String())
//	stdoutStderr, err := cmd.CombinedOutput()
//	log.Debugf("command output: %s", stdoutStderr)
//	return errors.Wrapf(err, "unable to run %s", cmd.String())
//}

func (sc *ScanClient) runDetect(args []string) error {
	if err := sc.ensureScanClientIsDownloaded(); err != nil {
		return errors.WithMessagef(err, "cannot run detect.sh: scan client not downloaded")
	}

	args = append(args,
		fmt.Sprintf("--blackduck.url=https://%s", sc.host),
		fmt.Sprintf("--blackduck.password=%s", sc.password),
		fmt.Sprintf("--blackduck.username=%s", sc.username),
	)
	cmd := exec.Command(sc.detectPath, args...)
	//cmd.Env = append(cmd.Env, fmt.Sprintf("JAVA_HOME=%s", sc.scanClientInfo.ScanCliJavaHomePath()))
	cmd.Env = append(cmd.Env, fmt.Sprintf("DETECT_JAVA_PATH=%s", sc.scanClientInfo.ScanCliJavaPath()))
	log.Infof("about to run Detect with invocation: <%s> in environment [%+v]", cmd.String(), cmd.Env)
	stdoutStderr, err := cmd.CombinedOutput()
	log.Debugf("command output: %s", stdoutStderr)
	return errors.Wrapf(err, "unable to run %s", cmd.String())
}

func (sc *ScanClient) Scan(path string, cfg *ScanConfig) error {
	if cfg.DetectDocker != nil {
		d := cfg.DetectDocker
		return sc.DetectDocker(cfg.Names.ProjectName, cfg.Names.VersionName, cfg.Names.ScanName, d.ImageTag)
	} else if cfg.DetectSignatureScan {
		return sc.DetectSignatureScan(path, cfg.Names.ProjectName, cfg.Names.VersionName, cfg.Names.ScanName)
	} else if cfg.DetectBinaryScan {
		return sc.DetectBinaryScan(path, cfg.Names.ProjectName, cfg.Names.VersionName, cfg.Names.ScanName)
	} else if cfg.IScan {
		return sc.IScan(path, cfg.Names.ProjectName, cfg.Names.VersionName, cfg.Names.ScanName)
	} else {
		return errors.New("invalid scan config: everything was nil")
	}
}
