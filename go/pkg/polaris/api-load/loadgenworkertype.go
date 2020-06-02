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
package api_load

//
//import (
//	"encoding/json"
//	"fmt"
//	"github.com/pkg/errors"
//)
//
//type LoadGenWorkerType int
//
//const (
//	LoadGenWorkerTypeEntitlements LoadGenWorkerType = iota
//	LoadGenWorkerTypeGroups       LoadGenWorkerType = iota
//	LoadGenWorkerTypeJobs         LoadGenWorkerType = iota
//	LoadGenWorkerTypeProjects     LoadGenWorkerType = iota
//	LoadGenWorkerTypeTaxonomies   LoadGenWorkerType = iota
//)
//
//// String .....
//func (l LoadGenWorkerType) String() string {
//	switch l {
//	case LoadGenWorkerTypeEntitlements:
//		return "LoadGenWorkerTypeEntitlements"
//	case LoadGenWorkerTypeGroups:
//		return "LoadGenWorkerTypeGroups"
//	case LoadGenWorkerTypeJobs:
//		return "LoadGenWorkerTypeJobs"
//	case LoadGenWorkerTypeProjects:
//		return "LoadGenWorkerTypeProjects"
//	case LoadGenWorkerTypeTaxonomies:
//		return "LoadGenWorkerTypeTaxonomies"
//	}
//	panic(fmt.Errorf("invalid LoadGenWorkerType value: %d", l))
//}
//
//func (l LoadGenWorkerType) MarshalJSON() ([]byte, error) {
//	jsonString := fmt.Sprintf(`"%s"`, l.String())
//	return []byte(jsonString), nil
//}
//
//func (l LoadGenWorkerType) MarshalText() (text []byte, err error) {
//	return []byte(l.String()), nil
//}
//
//func (l *LoadGenWorkerType) UnmarshalText(text []byte) (err error) {
//	status, err := parseLoadGenWorkerType(string(text))
//	if err != nil {
//		return err
//	}
//	*l = status
//	return nil
//}
//
//func (l *LoadGenWorkerType) UnmarshalJSON(b []byte) error {
//	var str string
//	err := json.Unmarshal(b, &str)
//	if err != nil {
//		return errors.Wrapf(err, "unable to UnmarshalJSON for value %s", string(b))
//	}
//	lType, err := parseLoadGenWorkerType(str)
//	if err != nil {
//		return errors.Wrapf(err, "unable to parse load gen worker type for value %s", str)
//	}
//	*l = lType
//	return nil
//}
//
//func parseLoadGenWorkerType(text string) (LoadGenWorkerType, error) {
//	switch text {
//	case "Entitlements":
//		return LoadGenWorkerTypeEntitlements, nil
//	case "Groups":
//		return LoadGenWorkerTypeGroups, nil
//	case "Jobs":
//		return LoadGenWorkerTypeJobs, nil
//	case "Projects":
//		return LoadGenWorkerTypeProjects, nil
//	case "Taxonomies":
//		return LoadGenWorkerTypeTaxonomies, nil
//	default:
//		return LoadGenWorkerTypeEntitlements, errors.New(fmt.Sprintf("invalid LoadGenWorkerType name: %s", text))
//	}
//}
