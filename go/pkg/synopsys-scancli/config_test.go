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

package synopsys_scancli

import (
	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
)

var config = `
{
  "Blackduck": {
    "Host": "35.222.148.238",
    "Username": "sysadmin",
    "Password": "blackduck",
    "Port": 443,
    "OSType": "linux",
    "ClientTimeoutSeconds": 300
  },
  "ImageFacade": {
    "PrivateDockerRegistries": [],
    "ImagePullerType": "docker",
    "ImageDirectory": "/tmp/images",
    "DeleteImageAfterScan": false
  },
  "ScanQueue": {
    "Host": "localhost",
    "Port": 4100
  },
  "Port": 4102,
  "LogLevel": "debug"
}
`

func RunConfigTests() {
	Describe("Config", func() {
		It("should parse config correctly", func() {
			// setup
			file, err := ioutil.TempFile("/tmp", "config-test-*.json")
			Expect(err).To(Succeed())
			log.Infof("created file %s", file.Name())
			Expect(ioutil.WriteFile(file.Name(), []byte(config), 0644)).To(Succeed())

			// standard json
			standardJson := &Config{}
			Expect(json.Unmarshal([]byte(config), standardJson)).To(Succeed())

			// viper
			viperConfig, err := GetConfig(file.Name())
			Expect(err).To(Succeed())

			// comparison
			Expect(standardJson).To(Equal(viperConfig))
		})
	})
}
