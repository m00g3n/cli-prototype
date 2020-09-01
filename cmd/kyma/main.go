/*
* CODE GENERATED AUTOMATICALLY WITH devops/internal/config
 */

package main

import (
	"encoding/json"
	"github.com/docopt/docopt-go"
	"gitops/internal/unstructured"
	"gitops/internal/workspace"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"path"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

const (
	usage = `kyma description

Usage:
    kyma [options]
    kyma function init --runtime=<RUNTIME> [options]
    kyma function apply [options]

Options:
    --kubeConfig			Path to kube config file.
	--debug                 Enable verbose output.
	-h --help               Show this screen.
	--version               Show version.`

	version = "0.0.1"
)

type config struct {
	KubeConfig string `docopt:"--kubeConfig" json:"kubeConfig"`
	Name       string `docopt:"--name" json:"name"`
	Debug      bool   `docopt:"--debug" json:"debug"`
	Runtime    string `docopt:"--runtime" json:"runtime"`
	Function   bool   `docopt:"function" json:"function"`
	Init       bool   `docopt:"init" json:"init"`
	Apply      bool   `docopt:"apply" json:"apply"`
}

func newConfig() (*config, error) {
	arguments, err := docopt.ParseArgs(usage, nil, version)
	if err != nil {
		return nil, err
	}
	var cfg config
	if err = arguments.Bind(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

var groupResourceVersionFunction = schema.GroupVersionResource{
	Group:    "serverless.kyma-project.io",
	Version:  "v1alpha1",
	Resource: "functions"}

func client(cfg *config) dynamic.Interface {
	home := homedir.HomeDir()

	if cfg.KubeConfig == "" && home == "" {
		log.Fatal("unable to find kubeconfig file")
	}

	if cfg.KubeConfig == "" {
		cfg.KubeConfig = filepath.Join(home, ".kube", "config")
	}

	entry := log.WithField("kubeConfig", cfg.KubeConfig)

	entry.Debug("building client from configuration")
	config, err := clientcmd.BuildConfigFromFlags("", cfg.KubeConfig)
	if err != nil {
		entry.Fatal(err)
	}

	forConfig, err := dynamic.NewForConfig(config)
	if err != nil {
		entry.Fatal(err)
	}
	entry.Debug("client built")
	return forConfig
}

func initializeWorkspace(cfg *config) {
	entry := log.WithField("runtime", cfg.Runtime)
	entry.Debug("initializing project")

	srcPath := "/tmp/testme"
	configuration := workspace.Cfg{
		Runtime:    v1alpha1.Nodejs12,
		Name:       "testme",
		Namespace:  "default",
		SourcePath: srcPath,
	}

	if err := workspace.Initialize(configuration, srcPath); err != nil {
		entry.Fatal(err)
	}
	entry.Debug("workspace initialized")
}

func applyFunction(cfg *config) {
	srcPath := "/tmp/testme"
	entry := log.WithField("sourcePath", srcPath)
	entry.Debug("opening project")

	file, err := os.Open(path.Join(srcPath, workspace.CfgFilename))
	if err != nil {
		entry.Fatal(err)
	}

	var configuration workspace.Cfg
	if err := json.NewDecoder(file).Decode(&configuration); err != nil {
		entry.Fatal(err)
	}

	client := client(cfg)
	resourceInterface := client.Resource(groupResourceVersionFunction).Namespace(configuration.Namespace)

	configuration.SourcePath = srcPath

	obj, err := unstructured.NewFunction(configuration)
	if err != nil {
		entry.Fatal(err)
	}

	data, err := json.Marshal(&obj)
	if err != nil {
		entry.Error(err)
	}
	entry.Debug("Creating object:", string(data))


	result, err := resourceInterface.Create(&obj, v1.CreateOptions{})

	if err == nil {
		entry.Debug("object created:", result)
		return
	}

	entry.Fatal(err)
}

func main() {
	// parse command arguments
	cfg, err := newConfig()
	if err != nil {
		log.Fatal(err)
	}

	if cfg.Debug {
		log.SetLevel(log.DebugLevel)
	}

	if cfg.Init {
		initializeWorkspace(cfg)
	}

	if cfg.Apply {
		applyFunction(cfg)
	}
}
