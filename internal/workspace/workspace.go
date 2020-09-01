package workspace

import (
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
)

type FileName string

type workspace []file

func (ws workspace) build(cfg Cfg, dirPath string) error {
	workspaceFiles := append(ws, cfg)
	for _, fileTemplate := range workspaceFiles {
		if err := write(dirPath, fileTemplate, cfg); err != nil {
			return err
		}
	}
	return nil
}

func write(destinationDirPath string, fileTemplate file, cfg Cfg) error {
	outFilePath := path.Join(destinationDirPath, fileTemplate.fileName())

	entry := log.WithFields(map[string]interface{}{
		"outputFileName": outFilePath,
		"workspaceName":  cfg.fileName,
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
	err = fileTemplate.write(file, cfg)
	if err != nil {
		return err
	}
	entry.Debug("file generated")

	return nil
}

var errUnsupportedRuntime = errors.New("unsupported runtime")

func Initialize(cfg Cfg, dirPath string) error {
	ws, err := fromRuntime(cfg.Runtime)
	if err != nil {
		return err
	}
	return ws.build(cfg, dirPath)
}

func fromRuntime(runtime v1alpha1.Runtime) (workspace, error) {
	switch runtime {
	case v1alpha1.Nodejs12, v1alpha1.Nodejs10:
		return workspaceNodeJs, nil
	case v1alpha1.Python38:
		return workspacePython, nil
	default:
		return nil, errUnsupportedRuntime
	}
}
