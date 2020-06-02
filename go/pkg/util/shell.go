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
	"context"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"os/exec"
	"path"
	"time"
)

func ExecShell(shellCmd string, directory string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	execCmd := exec.CommandContext(ctx, "sh", "-c", shellCmd)
	if len(directory) > 0 {
		execCmd.Dir = path.Clean(directory)
	}

	log.Infof("about to run %s in directory %s", execCmd.String(), directory)
	output, err := execCmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		log.Errorf("command %s in %s timed out", execCmd.String(), directory)
		return "", errors.Wrapf(ctx.Err(), "command timed out")
	}

	if err != nil {
		log.Errorf("failed to run %s in directory %s: %s", execCmd.String(), directory, err)
		return string(output), errors.Wrapf(err, "command failed: %s", output)
	}

	log.Infof("successfully ran %s in directory %s", execCmd.String(), directory)
	return string(output), nil
}
