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

package custom

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"sigs.k8s.io/yaml"
)

const ConfigMapMountDir = "/etc/kubernetes/node-feature-discovery/custom.d"

// getConfigMapFeatureConfig returns features configured in additional ConfigMaps, which are mounted to
// /etc/kubernetes/node-feature-discovery/custom.d
func getConfigMapFeatureConfig() []FeatureSpec {

	features := make([]FeatureSpec, 0)

	log.Printf("DEBUG: getting files in %s\n", ConfigMapMountDir)
	files, err := ioutil.ReadDir(ConfigMapMountDir)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("DEBUG: hostname config directory %v does not exist\n", ConfigMapMountDir)
		} else {
			log.Printf("ERROR: unable to access hostname config directory %v\n", ConfigMapMountDir)
		}
		return features
	}

	for _, file := range files {
		configFile := path.Join(ConfigMapMountDir, file.Name())
		log.Printf("DEBUG: processing file %s\n", configFile)

		if file.IsDir() {
			log.Printf("DEBUG: skipping dir %s\n", configFile)
			continue
		}
		if strings.HasPrefix(file.Name(), ".") {
			log.Printf("DEBUG: skipping hidden file %s\n", configFile)
			continue
		}

		bytes, err := ioutil.ReadFile(configFile)
		if err != nil {
			log.Printf("ERROR: could not read custom config file %s,\n%v\n", configFile, err)
			continue
		}
		log.Printf("DEBUG: custom config rules raw: %s\n", string(bytes))

		config := &[]FeatureSpec{}
		err = yaml.UnmarshalStrict(bytes, config)
		if err != nil {
			log.Printf("ERROR: could not parse custom config file %s, %v\n", configFile, err)
			continue
		}

		features = append(features, *config...)
	}

	log.Printf("DEBUG: all configmap based custom feature specs: %+v\n", features)

	return features
}
