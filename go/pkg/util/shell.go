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
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
)

func CommandSetDirectory(cmd *exec.Cmd, directory string) {
	if len(directory) == 0 {
		currDirectory, _ := os.Executable()
		cmd.Dir = currDirectory
	} else {
		cmd.Dir = directory
	}
}

func CommandAddEnvironment(cmd *exec.Cmd, env map[string]string) {
	cmd.Env = os.Environ()
	for key, val := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, val))
	}
}

func CommandRunAndPrint(cmd *exec.Cmd) error {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// can't use RunCommand(cmd) here -- attaching to os pipes interferes with cmd.CombinedOutput()
	log.Infof("running command '%s' with pipes attached in directory: '%s'", cmd.String(), cmd.Dir)
	return errors.Wrapf(cmd.Run(), "unable to run command '%s'", cmd.String())
}
