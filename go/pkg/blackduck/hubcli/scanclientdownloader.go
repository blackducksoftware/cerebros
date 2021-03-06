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

import "time"

//import (
//	"fmt"
//	"github.com/blackducksoftware/cerebros/go/pkg/util"
//	"os"
//	"time"
//
//	"github.com/blackducksoftware/hub-client-go/hubclient"
//	"github.com/pkg/errors"
//	log "github.com/sirupsen/logrus"
//)

// DownloadScanClient ...
func DownloadScanClient(osType OSType, cliRootPath string, hubHost string, hubUser string, hubPassword string, hubPort int, timeout time.Duration) (*ScanClientInfo, error) {
	panic("TODO - fix up hub-client-go dependency to support go modules")
	//// 1. instantiate hub client
	//hubBaseURL := fmt.Sprintf("https://%s:%d", hubHost, hubPort)
	//hubClient, err := hubclient.NewWithSession(hubBaseURL, hubclient.HubClientDebugTimings, timeout)
	//if err != nil {
	//	return nil, errors.Wrapf(err, "unable to instantiate hub client")
	//}
	//
	//log.Infof("successfully instantiated hub client %s", hubBaseURL)
	//
	//// 2. get hub version
	//currentVersion, err := hubClient.CurrentVersion()
	//if err != nil {
	//	return nil, errors.Wrapf(err, "unable to get hub version")
	//}
	//log.Infof("got hub version: %s", currentVersion.Version)
	//
	//cliInfo := NewScanClientInfo(currentVersion.Version, cliRootPath, osType)
	//exists, err := util.FileExists(cliRootPath)
	//if err != nil {
	//	return nil, errors.WithMessagef(err, "unable to check if scanner file already present at %s", cliRootPath)
	//}
	//if exists {
	//	log.Infof("scan client found at %s, skipping download", cliRootPath)
	//	return cliInfo, nil
	//}
	//
	//// 3. log in to hub client
	//err = hubClient.Login(hubUser, hubPassword)
	//if err != nil {
	//	return nil, errors.Wrapf(err, "unable to log in to hub")
	//}
	//log.Info("successfully logged in to hub")
	//
	//// 4. create directory
	//err = os.MkdirAll(cliInfo.RootPath, 0755)
	//if err != nil {
	//	return nil, errors.Wrapf(err, "unable to make dir %s", cliInfo.RootPath)
	//}
	//
	//// 5. pull down scan client as .zip
	//log.Infof("downloading scan client type %s from %s", osType.String(), hubHost)
	//switch osType {
	//case OSTypeMac:
	//	err = hubClient.DownloadScanClientMac(cliInfo.ScanCliZipPath())
	//case OSTypeLinux:
	//	err = hubClient.DownloadScanClientLinux(cliInfo.ScanCliZipPath())
	//}
	//if err != nil {
	//	return nil, errors.Wrapf(err, "unable to download scan client")
	//}
	//
	//log.Infof("successfully downloaded scan client to %s", cliInfo.ScanCliZipPath())
	//
	//// 6. unzip scan client
	//_, err = util.Unzip(cliInfo.ScanCliZipPath(), cliInfo.RootPath)
	//if err != nil {
	//	return nil, errors.WithMessagef(err, "unable to unzip %s to %s", cliInfo.ScanCliZipPath(), cliInfo.RootPath)
	//}
	//log.Infof("successfully unzipped from %s to %s", cliInfo.ScanCliZipPath(), cliInfo.RootPath)
	//
	//// 7. we're done
	//return cliInfo, nil
}
