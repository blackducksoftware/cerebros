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
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
)

type OSType string

const (
	OSTypeLinux   OSType = "Linux"
	OSTypeMac     OSType = "Mac"
	OSTypeWindows OSType = "Windows"
)

func (t OSType) String() string {
	switch t {
	case OSTypeLinux:
		return "OSTypeLinux"
	case OSTypeMac:
		return "OSTypeMac"
	case OSTypeWindows:
		return "OSTypeWindows"
	}
	panic(fmt.Errorf("invalid OSType value: %s", string(t)))
}

func (t OSType) MarshalJSON() ([]byte, error) {
	jsonString := fmt.Sprintf(`"%s"`, t.String())
	return []byte(jsonString), nil
}

func (t OSType) MarshalText() (text []byte, err error) {
	return []byte(t.String()), nil
}

func parseOSType(str string) (OSType, error) {
	switch str {
	case "Linux", "linux":
		return OSTypeLinux, nil
	case "Mac", "mac":
		return OSTypeMac, nil
	case "Windows", "windows":
		return OSTypeWindows, nil
	}
	return OSTypeLinux, errors.New(fmt.Sprintf("invalid ostype: %s", str))
}

func (t *OSType) UnmarshalJSON(data []byte) error {
	var str string
	err := json.Unmarshal(data, &str)
	if err != nil {
		return err
	}
	osType, err := parseOSType(str)
	if err != nil {
		return err
	}
	*t = osType
	return nil
}

func (t *OSType) UnmarshalText(text []byte) (err error) {
	osType, err := parseOSType(string(text))
	if err != nil {
		return err
	}
	*t = osType
	return nil
}
