package workspace

import (
	"encoding/json"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"io"
)

var _ File = &Cfg{}

type Cfg struct {
	Runtime       v1alpha1.Runtime
	WorkspaceName string `json:"name"`
}

func (c Cfg) Generate(writer io.Writer, cfg Cfg) error {
	return json.NewEncoder(writer).Encode(&cfg)
}

func (c Cfg) Name() string {
	return "serverless.yaml"
}
