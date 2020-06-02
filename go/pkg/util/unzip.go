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

package util

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type closable interface {
	Close() error
}

// This code is from https://golangcode.com/unzip-files-in-go/ and
// https://stackoverflow.com/a/24792688/894284
func Unzip(source string, destination string) ([]string, error) {
	log.Infof("unzipping %s to %s", source, destination)
	filenames := []string{}
	r, err := zip.OpenReader(source)
	if err != nil {
		return filenames, errors.Wrapf(err, "unable to open reader")
	}
	defer log.Tracef("about to close %s", source)
	defer r.Close()

	openFiles := []closable{}
	addFileHandle := func(rc closable) {
		openFiles = append(openFiles, rc)
	}
	closeFiles := func() {
		log.Tracef("closing %d files", len(openFiles))
		for _, rc := range openFiles {
			err := rc.Close()
			if err != nil {
				log.Errorf("unable to close file: %s", err.Error())
			} else {
				log.Tracef("closed file")
			}
		}
		openFiles = []closable{}
	}
	defer closeFiles()

	for _, f := range r.File {
		log.Tracef("looking at %s", f.Name)
		rc, err := f.Open()
		if err != nil {
			return filenames, errors.Wrapf(err, "unable to open file")
		}
		addFileHandle(rc)

		fpath := filepath.Join(destination, f.Name)
		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
		} else {
			var fdir string
			if lastIndex := strings.LastIndex(fpath, string(os.PathSeparator)); lastIndex > -1 {
				fdir = fpath[:lastIndex]
			}

			err = os.MkdirAll(fdir, os.ModePerm)
			if err != nil {
				return filenames, errors.Wrapf(err, "unable to make directory")
			}
			log.Tracef("looking at %s", fpath)
			f, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return filenames, errors.Wrapf(err, "unable to open file")
			}
			addFileHandle(f)

			_, err = io.Copy(f, rc)
			if err != nil {
				return filenames, errors.Wrapf(err, "unable to copy file")
			}
		}
		closeFiles()
	}
	return filenames, nil
}
