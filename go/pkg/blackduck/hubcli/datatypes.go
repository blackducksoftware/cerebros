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
package hubcli

//import (
//	"encoding/json"
//	"fmt"
//	"github.com/pkg/errors"
//)
//
//// ScanType describes the different types of scans possible with a Hub
//type ScanType string
//
//// .....
//const (
//	ScanTypeIScan     ScanType = "iscan"
//	ScanTypeDocker    ScanType = "docker"
//	ScanTypeBinary    ScanType = "binary"
//	ScanTypeSignature ScanType = "signature"
//)
//
//func ParseScanType(str string) (ScanType, error) {
//	switch str {
//	case "iscan":
//		return ScanTypeIScan, nil
//	case "docker":
//		return ScanTypeDocker, nil
//	case "binary":
//		return ScanTypeBinary, nil
//	case "signature":
//		return ScanTypeSignature, nil
//	}
//	return "", errors.New(fmt.Sprintf("invalid scantype: %s", str))
//}
//
//// String .....
//func (st ScanType) String() string {
//	switch st {
//	case ScanTypeIScan:
//		return "ScanTypeIScan"
//	case ScanTypeDocker:
//		return "ScanTypeDocker"
//	case ScanTypeBinary:
//		return "ScanTypeBinary"
//	case ScanTypeSignature:
//		return "ScanTypeSignature"
//	}
//	panic(fmt.Errorf("invalid ScanType value: %s", string(st)))
//}
//
//// MarshalJSON .....
//func (st ScanType) MarshalJSON() ([]byte, error) {
//	return []byte(st), nil
//}
//
//// MarshalText .....
//func (st ScanType) MarshalText() (text []byte, err error) {
//	return []byte(st), nil
//}
//
//func (st *ScanType) UnmarshalJSON(data []byte) error {
//	var str string
//	err := json.Unmarshal(data, &str)
//	if err != nil {
//		return err
//	}
//	status, err := ParseScanType(str)
//	if err != nil {
//		return err
//	}
//	*st = status
//	return nil
//}
//
//func (st *ScanType) UnmarshalText(text []byte) (err error) {
//	status, err := ParseScanType(string(text))
//	if err != nil {
//		return err
//	}
//	*st = status
//	return nil
//}
//
//type ImageScanConfig struct {
//	//Repository            string
//	//Sha                   string
//	HubProjectName        string
//	HubProjectVersionName string
//	HubScanName           string
//	ScanType              ScanType
//	ImageTag              string
//}

//func (hisr *ImageScanConfig) pullSpec() string {
//	return fmt.Sprintf("%s@sha256:%s", hisr.Repository, hisr.Sha)
//}
