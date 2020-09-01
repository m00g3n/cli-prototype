package unstructured

import (
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"gitops/internal/workspace"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"path"
)

const (
	functionApiVersion = "serverless.kyma-project.io/v1alpha1"
)

func NewFunction(cfg workspace.Cfg) (unstructured.Unstructured, error) {
	out := unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": functionApiVersion,
		"kind":       "Function",
		"metadata": map[string]interface{}{
			"name":   cfg.Name,
			"labels": cfg.Labels,
		},
		"spec": map[string]string{},
	}}

	spec := out.Object["spec"].(map[string]string)
	for key, value := range runtimeMappings[cfg.Runtime] {
		filePath := path.Join(cfg.SourcePath, string(value))
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			return unstructured.Unstructured{}, err
		}
		if len(data) == 0 {
			continue
		}
		spec[string(key)] = string(data)
	}

	return out, nil
}

type property string

const (
	propertySource = "source"
	propertyDeps   = "deps"
)

var (
	runtimeMappings = map[v1alpha1.Runtime]map[property]workspace.FileName{
		v1alpha1.Nodejs12: {
			propertySource: workspace.FileNameHandlerJs,
			propertyDeps:   workspace.FileNameHandlerJs,
		},
		v1alpha1.Nodejs10: {
			propertySource: workspace.FileNameHandlerJs,
			propertyDeps:   workspace.FileNamePackageJSON,
		},
		v1alpha1.Python38: {
			propertySource: workspace.FileNameHandlerPy,
			propertyDeps:   workspace.FileNameRequirementsTxt,
		},
	}
)
