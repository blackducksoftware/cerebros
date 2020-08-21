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
package synopsys_scancli

import (
	"github.com/blackducksoftware/cerebros/go/pkg/blackduck/hubcli"
)

type CodeLocation struct {
	GitRepo     *GitRepo
	FileSystem  *FileSystem
	DockerImage *DockerImage
	None        bool
}

type GitRepo struct {
	Repo string
}

type FileSystem struct {
	Path string
}

type DockerImage struct {
	PullSpec string
	//Registries []docker.RegistryAuth // TODO handle this elsewhere?
}

type PolarisScanConfig struct {
	UseLocalAnalysis bool
}

type ScanTypeConfig struct {
	Polaris   *PolarisScanConfig
	Blackduck *hubcli.ScanConfig
}

type ScanConfig struct {
	Key          string
	ScanType     *ScanTypeConfig
	CodeLocation *CodeLocation
}
