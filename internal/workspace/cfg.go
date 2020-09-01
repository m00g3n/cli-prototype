package workspace

import (
	"encoding/json"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"io"
)

var _ file = &Cfg{}

const CfgFilename = "serverless.json"

type Cfg struct {
	Runtime    v1alpha1.Runtime
	Git        bool              `json:"git,omitempty"`
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	SourcePath string            `json:"-"`
	Labels     map[string]string `json:"labels,omitempty"`
}

func (cfg Cfg) write(writer io.Writer, _ interface{}) error {
	return json.NewEncoder(writer).Encode(&cfg)
}

func (cfg Cfg) fileName() string {
	return CfgFilename
}
