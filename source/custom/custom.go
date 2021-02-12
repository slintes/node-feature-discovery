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
	"log"
	"reflect"

	"sigs.k8s.io/node-feature-discovery/source"
	"sigs.k8s.io/node-feature-discovery/source/custom/rules"
)

// Custom Features Configurations
type MatchRule struct {
	PciID      *rules.PciIDRule      `json:"pciId,omitempty"`
	UsbID      *rules.UsbIDRule      `json:"usbId,omitempty"`
	LoadedKMod *rules.LoadedKModRule `json:"loadedKMod,omitempty"`
	CpuID      *rules.CpuIDRule      `json:"cpuId,omitempty"`
	Kconfig    *rules.KconfigRule    `json:"kConfig,omitempty"`
	Hostname   *rules.HostnameRule   `json:"hostname,omitempty"`
}

type FeatureSpec struct {
	Name    string      `json:"name"`
	Value   *string     `json:"value"`
	MatchOn []MatchRule `json:"matchOn"`
}

type config []FeatureSpec

// newDefaultConfig returns a new config with pre-populated defaults
func newDefaultConfig() *config {
	return &config{}
}

// Implements FeatureSource Interface
type Source struct {
	config *config
}

// Return name of the feature source
func (s Source) Name() string { return "custom" }

// NewConfig method of the FeatureSource interface
func (s *Source) NewConfig() source.Config { return newDefaultConfig() }

// GetConfig method of the FeatureSource interface
func (s *Source) GetConfig() source.Config { return s.config }

// SetConfig method of the FeatureSource interface
func (s *Source) SetConfig(conf source.Config) {
	switch v := conf.(type) {
	case *config:
		s.config = v
	default:
		log.Printf("PANIC: invalid config type: %T", conf)
	}
}

// Discover features
func (s Source) Discover() (source.Features, error) {
	features := source.Features{}
	allFeatureConfig := append(getStaticFeatureConfig(), *s.config...)
	allFeatureConfig = append(allFeatureConfig, getConfigMapFeatureConfig()...)
	log.Printf("INFO: Custom features: %+v\n", allFeatureConfig)
	// Iterate over features
	for _, customFeature := range allFeatureConfig {
		featureExist, err := s.discoverFeature(customFeature)
		if err != nil {
			log.Printf("ERROR: failed to discover feature: %q: %s\n", customFeature.Name, err.Error())
			continue
		}
		if featureExist {
			var value interface{} = true
			if customFeature.Value != nil {
				value = *customFeature.Value
			}
			features[customFeature.Name] = value
		}
	}
	return features, nil
}

// Process a single feature by Matching on the defined rules.
// A feature is present if all defined Rules in a MatchRule return a match.
func (s Source) discoverFeature(feature FeatureSpec) (bool, error) {
	for _, rule := range feature.MatchOn {

		match := true

		// iterate all fields of the MatchRule
		r := reflect.ValueOf(rule)
		for i := 0; i < r.NumField(); i++ {
			// if we have a public field which isn't nil and implements Rule...
			f := r.Field(i)
			if f.CanInterface() {
				if _, ok := f.Interface().(rules.Rule); ok && !f.IsNil() {
					log.Printf("DEBUG: evaluating rule %s\n", r.Type().Field(i).Name)
					// .. call the Match func
					result := f.MethodByName("Match").Call([]reflect.Value{})
					// check error
					if !result[1].IsNil() {
						log.Printf("DEBUG: error")
						return false, result[1].Interface().(error)
					}
					log.Println("DEBUG: no error")
					// check match result
					if !result[0].Bool() {
						log.Println("DEBUG: no match")
						match = false
						break
					}
					log.Println("DEBUG: match!")
				}
			}
		}

		if !match {
			continue
		}

		return true, nil

	}
	return false, nil
}
