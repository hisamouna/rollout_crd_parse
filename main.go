package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"sigs.k8s.io/yaml"
)

type RolloutCrd struct {
	Spec struct {
		Validation struct {
			OpenAPIV3Schema OpenAPIV3Schema `yaml:"openAPIV3Schema"`
		} `yaml:"validation"`
	} `yaml:"spec"`
}

type OpenAPIV3Schema struct {
	Props    interface{} `json:"properties" yaml:"properties"`
	Required []string    `json:"required" yaml:"required"`
	Type     string      `json:"type" yaml:"type"`
}

type SchemaJson struct {
	Definitions Definitions `json:"definitions"`
}

type Definitions struct {
	Rollout Rollout `json:"v1alpha1.Rollout"`
}

type Rollout struct {
	Props        interface{}   `json:"properties"`
	Required     []string      `json:"required"`
	Type         string        `json:"type"`
	VersionKinds []VersionKind `json:"x-kubernetes-group-version-kind"`
}

type VersionKind struct {
	Group   string `json:"group"`
	Kind    string `json:"kind"`
	Version string `json:"version"`
}

func main() {
	fb, err := downloadFile()
	if err != nil {
		log.Fatalf("downloadFile: %s", err)
	}
	b, err := parseYaml(fb)
	if err != nil {
		log.Fatalf("parsYaml: %s", err)
	}

	j, err := yaml.YAMLToJSON(b)
	if err != nil {
		log.Fatalf("yaml.YAMLToJSON: %s", err.Error())
	}

	var sch OpenAPIV3Schema
	err = json.Unmarshal(j, &sch)
	if err != nil {
		log.Fatalf("json.Unmarshal: %s", err.Error())
	}

	sj := &SchemaJson{
		Definitions: Definitions{
			Rollout: Rollout{
				sch.Props,
				sch.Required,
				sch.Type,
				[]VersionKind{
					{
						"argoproj.io",
						"Rollout",
						"v1alpha1",
					},
				},
			},
		},
	}

	jsonFile, err := json.MarshalIndent(sj, "", " ")
	if err != nil {
		log.Fatalf("json.MarshalIndent: %s", err.Error())
	}

	err = ioutil.WriteFile("schema.json", jsonFile, 0644)
	if err != nil {
		log.Fatalf("ioutil.WriteFile: %s", err.Error())
	}
}

func downloadFile() ([]byte, error) {
	url := "https://raw.githubusercontent.com/argoproj/argo-rollouts/v0.10.2/manifests/crds/rollout-crd.yaml"

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf := &bytes.Buffer{}
	io.Copy(buf, resp.Body)

	return buf.Bytes(), nil
}

// 元のyamlからOpenAPIスキーマに必要な要素のみ抽出
func parseYaml(inputBytes []byte) ([]byte, error) {
	rcrd := &RolloutCrd{}
	err := yaml.Unmarshal(inputBytes, rcrd)
	if err != nil {
		return nil, err
	}

	return yaml.Marshal(rcrd.Spec.Validation.OpenAPIV3Schema)
}
