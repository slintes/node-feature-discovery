/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package rules

import (
	"log"
	"os"
	"regexp"
)

var (
	// TODO is it safe to assume nodeName == hostname??
	nodeName = os.Getenv("NODE_NAME")
)

// Rule that matches on hostnames configured in a ConfigMap
type HostnameRule []string

// Force implementation of Rule
var _ Rule = HostnameRule{}

func (h HostnameRule) Match() (bool, error) {

	for _, hostnamePattern := range h {
		log.Printf("DEBUG: matchHostname %s\n", hostnamePattern)

		match, err := regexp.MatchString(hostnamePattern, nodeName)
		if err != nil {
			log.Printf("ERROR: error testing regex: pattern %s, nodename %s, %v", hostnamePattern, nodeName, err)
			continue
		}
		if !match {
			log.Printf("DEBUG: no match")
			continue
		}
		log.Printf("DEBUG: match!")
		return true, nil
	}

	return false, nil
}
