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
	//	. "github.com/onsi/gomega"
	"encoding/json"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type TestStruct struct {
	A string
	B struct {
		C OSType
	}
	D OSType
}

func RunOSTypeTests() {
	Describe("OSType", func() {
		It("should parse from json", func() {
			text := []byte(`["linux", "Linux", "mac", "Mac", "windows", "Windows"]`)
			var arr []OSType
			Expect(json.Unmarshal(text, &arr)).To(Succeed())
			Expect(arr).To(Equal([]OSType{OSTypeLinux, OSTypeLinux, OSTypeMac, OSTypeMac, OSTypeWindows, OSTypeWindows}))
		})
		It("should parse from json within a struct", func() {
			text := []byte(`
				{
				  "A": "aaa",
				  "B": {
					"C": "linux"
				  },
				  "D": "linux"
				}
			`)
			var ts *TestStruct
			Expect(json.Unmarshal(text, &ts)).To(Succeed())
			Expect(ts).To(Equal(&TestStruct{A: "aaa", B: struct{ C OSType }{C: OSTypeLinux}, D: OSTypeLinux}))
			Expect(fmt.Sprintf("%+v", ts)).To(Equal("&{A:aaa B:{C:OSTypeLinux} D:OSTypeLinux}"))
		})
	})
}
