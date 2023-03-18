// Copyright © 2021-2023 The Gomon Project.

package gocore

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"
)

// YamlMap reads a YAML configuration into a MapSlice.
func YamlMap(yml string) (*yaml.MapSlice, error) {
	ms := yaml.MapSlice{}
	err := yaml.Unmarshal([]byte(yml), &ms)
	return &ms, err
}

// YamlValue uses the keypath to extract the nested element from a yaml configuration.
func YamlValue(keypath []string, yml any) string {
	return strings.TrimSpace(yamlValue(keypath, yml))
}

// yamlValue uses the keypath to extract the nested element from a yaml configuration.
func yamlValue(keypath []string, yml any) string {
	if len(keypath) == 0 {
		return yamlDecode("", yml)
	}
	var is []any
	switch yml := yml.(type) {
	default:
		return fmt.Sprint(yml)
	case yaml.MapItem:
		if keypath[0] == yml.Key {
			return yamlDecode("", yml.Value)
		}
	case []any:
		is = yml
	case yaml.MapSlice:
		for _, m := range yml {
			is = append(is, m)
		}
	}

	for _, i := range is {
		switch i := i.(type) {
		case yaml.MapSlice:
			if keypath[0] == i[0].Key {
				if len(keypath) == 1 {
					return yamlDecode("", is)
				}
				if v, ok := i[0].Value.(string); ok {
					if keypath[1] == v {
						return yamlValue(keypath[2:], i)
					}
				} else {
					return yamlValue(keypath[1:], i)
				}
			}
		case yaml.MapItem:
			if keypath[0] == i.Key {
				if _, ok := i.Value.(string); ok && len(keypath) > 1 {
					return yamlValue(keypath[1:], i)
				}
				return yamlValue(keypath[1:], i.Value)
			}
		}
	}
	return ""
}

// yamlDecode decodes the yaml object value
func yamlDecode(indent string, yml any) string {
	var s string
	switch yml := yml.(type) {
	case yaml.MapSlice:
		s = "\n"
		for _, i := range yml {
			if len(indent) > 1 && indent[len(indent)-2] == '-' {
				s += indent[2:]
				indent = indent[:len(indent)-2]
			} else {
				s += indent
			}
			s += i.Key.(string) + ":"
			s += yamlDecode(indent+"  ", i.Value)
		}
	case []any:
		if len(yml) == 0 {
			return " []\n"
		}
		for _, i := range yml {
			s += yamlDecode(indent+"- ", i)
		}
	default:
		if len(indent) > 1 && indent[len(indent)-2] == '-' {
			return fmt.Sprintf("\n%s%s", indent[2:], yml)
		}
		return fmt.Sprintf(" %v\n", yml)
	}
	return s
}

// yml is an example prometheus.yml file to test decoding.
/*
var yml = []byte(`
global:
  scrape_interval: 15s
  scrape_timeout: 10s
  evaluation_interval: 15s
alerting:
  alertmanagers:
  - static_configs:
    - targets: []
    scheme: http
    timeout: 10s
    api_version: v1
scrape_configs:
- job_name: prometheus
  honor_timestamps: true
  scrape_interval: 15s
  scrape_timeout: 10s
  metrics_path: /metrics
  scheme: http
  static_configs:
  - targets:
    - localhost:9090
- job_name: grafana
  honor_timestamps: true
  scrape_interval: 15s
  scrape_timeout: 10s
  metrics_path: /metrics
  scheme: http
  static_configs:
  - targets:
    - localhost:3000
    - localhost:7777
- job_name: gomon
  honor_timestamps: true
  scrape_interval: 15s
  scrape_timeout: 10s
  metrics_path: /metrics
  scheme: http
  static_configs:
  - targets:
    - localhost:1234
`)

// test searches various keypaths for values in prometheus.yml.
func test() {
	var keypath []string
	var val string
	var dur time.Duration

	keypath = []string{"scrape_configs"}
	val = ValueYaml(keypath, yml)
	fmt.Fprintf(os.Stderr, "keypath %v value %s\n", keypath, val)

	keypath = []string{"scrape_configs", "job_name"}
	val = ValueYaml(keypath, yml)
	fmt.Fprintf(os.Stderr, "keypath %v value %s\n", keypath, val)

	keypath = []string{"scrape_configs", "job_name", "gomon"}
	val = ValueYaml(keypath, yml)
	fmt.Fprintf(os.Stderr, "keypath %v value %s\n", keypath, val)

	keypath = []string{"scrape_configs", "job_name", "grafana", "static_configs"}
	val = ValueYaml(keypath, yml)
	fmt.Fprintf(os.Stderr, "keypath %v value %s\n", keypath, val)

	keypath = []string{"alerting", "alertmanagers", "static_configs", "scheme"}
	val = ValueYaml(keypath, yml)
	fmt.Fprintf(os.Stderr, "keypath %v value %s\n", keypath, val)

	keypath = []string{"global", "scrape_interval"}
	val = ValueYaml(keypath, yml)
	fmt.Fprintf(os.Stderr, "keypath %v value %s\n", keypath, val)
	dur, _ = time.ParseDuration(strings.TrimSpace(val))
	fmt.Fprintf(os.Stderr, "default scrape_interval is %v\n", dur)

	keypath = []string{"scrape_configs", "job_name", "gomon", "scrape_interval"}
	val = ValueYaml(keypath, yml)
	fmt.Fprintf(os.Stderr, "keypath %v value %s\n", keypath, val)
	dur, _ = time.ParseDuration(strings.TrimSpace(val))
	fmt.Fprintf(os.Stderr, "gomon scrape_interval is %v\n", dur)
}
*/
