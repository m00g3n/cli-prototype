package workspace

import (
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
)

type workspace []File

func (ws workspace) build(cfg Cfg, dirPath string) error {
	workspaceFiles := append(ws, cfg)
	for _, fileTemplate := range workspaceFiles {
		if err := write(dirPath, fileTemplate, cfg); err != nil {
			return err
		}
	}
	return nil
}

func write(destinationPath string, fileTemplate File, cfg Cfg) error {
	outFilePath := path.Join(destinationPath, fileTemplate.Name())

	entry := log.WithFields(map[string]interface{}{
		"outputFileName": outFilePath,
		"workspaceName":  cfg.WorkspaceName,
	})

	entry.Debug("creating output file")
	file, err := os.Create(outFilePath)
	if err != nil {
		return err
	}
	defer func() {
		entry.Debug("closing file")
		err := file.Close()
		if err != nil {
			entry.Error(err)
		}
	}()
	entry.Debug("file created")

	entry.Debug("generating content")
	err = fileTemplate.Generate(file, cfg)
	if err != nil {
		return err
	}
	entry.Debug("file generated")

	return nil
}

var errUnsupportedRuntime = errors.New("unsupported runtime")

func Initialize(cfg Cfg, dirPath string) error {
	switch cfg.Runtime {
	case v1alpha1.Nodejs12:
		return nodeJS12.build(cfg, dirPath)
	default:
		return errUnsupportedRuntime
	}
}
