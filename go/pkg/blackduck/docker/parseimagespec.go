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

package docker

import (
	"fmt"
	"github.com/pkg/errors"
	"regexp"
)

//var dockerPullableRegexp = regexp.MustCompile("^docker-pullable://(.+)@sha256:([a-zA-Z0-9]+)$")
//var dockerRegexp = regexp.MustCompile("^docker://sha256:([a-zA-Z0-9]+)$")

var repoDigestRegexp = regexp.MustCompile("^(.+)@sha256:([a-zA-Z0-9]+)$")
var repoTagRegexp = regexp.MustCompile("^(.+?)(:[^/]+)?$")

// ParseImageIDString parses an ImageID that can pull an image from docker
// Example image id:
//   docker-pullable://registry.kipp.blackducksoftware.com/blackducksoftware/hub-registration@sha256:cb4983d8399a59bb5ee6e68b6177d878966a8fe41abe18a45c3b1d8809f1d043
//func ParseImageIDString(imageID string) (string, string, error) {
//	// Since the GO doesn't support for "don't start with string regex", had an ugly fix with HasPrefix in below code
//	if strings.HasPrefix(imageID, "docker-pullable://") {
//		return parseDockerPullableImageString(imageID)
//	}
//	if strings.HasPrefix(imageID, "docker://") {
//		return "", "", fmt.Errorf("scanning of unscheduled images (%s) is not supported, ", imageID)
//	}
//	return parseImageString(imageID)
//}

func parseRepoDigest(imageID string) (string, string, error) {
	match := repoDigestRegexp.FindStringSubmatch(imageID)
	if len(match) != 3 {
		return "", "", errors.New(fmt.Sprintf("unable to match imageRegexp regex <%s> to input <%s>", repoDigestRegexp.String(), imageID))
	}
	name := match[1]
	digest := match[2]
	return name, digest, nil
}

func parseRepoTag(image string) (string, string, error) {
	tagMatch := repoTagRegexp.FindStringSubmatch(image)
	if len(tagMatch) == 3 {
		tag := tagMatch[2]
		// drop the leading ':'  (e.g. ':latest' -> 'latest')
		if len(tag) > 0 {
			tag = tag[1:]
		}
		return tagMatch[1], tag, nil
	}
	return "", "", errors.New(fmt.Sprintf("unable to parse repo-tag from %s, match was %+v", image, tagMatch))
}

// func parseDockerImageString(imageID string) (string, string, error) {
// 	match := dockerRegexp.FindStringSubmatch(imageID)
// 	if len(match) != 2 {
// 		return "", "", fmt.Errorf("unable to match dockerRegexp regex <%s> to input <%s>", dockerRegexp.String(), imageID)
// 	}
// 	digest := match[1]
// 	return "", digest, nil
// }

//func parseDockerPullableImageString(imageID string) (string, string, error) {
//	match := dockerPullableRegexp.FindStringSubmatch(imageID)
//	if len(match) != 3 {
//		return "", "", fmt.Errorf("unable to match dockerPullableRegexp regex <%s> to input <%s>", dockerPullableRegexp.String(), imageID)
//	}
//	name := match[1]
//	digest := match[2]
//	return name, digest, nil
//}

//// ParseImageString will take a docker image string and return the repo and tag parts
//func ParseImageString(image string) (string, string) {
//	match := imageShaRegexp.FindStringSubmatch(image)
//	if len(match) == 3 {
//		return match[1], ""
//	}
//
//	tagMatch := imageTagRegexp.FindStringSubmatch(image)
//	if len(tagMatch) == 3 {
//		tag := tagMatch[2]
//		// drop the leading ':'  (e.g. ':latest' -> 'latest')
//		if len(tag) > 0 {
//			tag = tag[1:]
//		}
//		return tagMatch[1], tag
//	}
//	// TODO should we return an err here?
//	return "", ""
//}
