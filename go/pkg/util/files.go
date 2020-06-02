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
package util

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"os"
)

func CreateIfNotExists(path string) error {
	ok, err := FileExists(path)
	if err != nil {
		return errors.Wrapf(err, "unable to stat path %s", path)
	}
	if !ok {
		log.Infof("directory %s does not exist, creating ...", path)
		err = os.MkdirAll(path, 0755)
		if err != nil {
			return errors.Wrapf(err, "path %s does not exist and unable to create", path)
		}
		log.Infof("successfully created directory %s", path)
	} else {
		log.Infof("directory %s already exists, skipping creation", path)
	}
	return nil
}

// FileExists returns whether the given file or directory exists
func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
