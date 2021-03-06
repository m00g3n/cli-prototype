/*
* CODE GENERATED AUTOMATICALLY WITH devops/internal/config
 */

package main

import (
	"encoding/json"
	"github.com/docopt/docopt-go"
	"gitops/internal/unstructured"
	"gitops/internal/workspace"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/retry"
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
    kyma function init --runtime=<RUNTIME> [--dir=<DIR>] [options]
    kyma function apply [--dir=<DIR>] [options]

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
	Dir        string `docopt:"--dir" json:"dir"`
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

	if cfg.Dir == "" {
		cfg.Dir, err = os.Getwd()
		if err != nil {
			return nil, err
		}
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

	if cfg.Name == "" {
		cfg.Name = path.Base(cfg.Dir)
	}

	configuration := workspace.Cfg{
		Runtime:    v1alpha1.Runtime(cfg.Runtime),
		Name:       cfg.Name,
		Namespace:  "default",
		SourcePath: cfg.Dir,
	}

	if err := workspace.Initialize(configuration, cfg.Dir); err != nil {
		entry.Fatal(err)
	}
	entry.Debug("workspace initialized")
}

func applyFunction(cfg *config) {
	entry := log.WithField("sourcePath", cfg.Dir)
	entry.Debug("opening project")

	file, err := os.Open(path.Join(cfg.Dir, workspace.CfgFilename))
	if err != nil {
		entry.Fatal(err)
	}

	// Load project configuration
	var configuration workspace.Cfg
	if err := json.NewDecoder(file).Decode(&configuration); err != nil {
		entry.Fatal(err)
	}
	configuration.SourcePath = cfg.Dir

	client := client(cfg)
	resourceInterface := client.Resource(groupResourceVersionFunction).Namespace(configuration.Namespace)

	// Check if object exists
	response, err := resourceInterface.Get(configuration.Name, v1.GetOptions{})
	fnFound := !errors.IsNotFound(err)
	if err != nil && fnFound {
		entry.Fatal(err)
	}

	obj, err := unstructured.NewFunction(configuration)
	if err != nil {
		entry.Fatal(err)
	}

	// If object is up to date return
	var equal bool
	if fnFound {
		equal = equality.Semantic.DeepDerivative(response.Object["spec"], obj.Object["spec"])
	}

	if fnFound && equal {
		entry.Debug("object already created and up to date")
		return
	}

	// If object needs update
	if fnFound && !equal {
		response.Object["spec"] = obj.Object["spec"]
		entry.Debug("updating object")
		err = retry.RetryOnConflict(retry.DefaultRetry, func() (err error) {
			_, err = resourceInterface.Update(response, v1.UpdateOptions{})
			return
		})

		if err != nil {
			entry.Fatal(err)
		}
		entry.Debug("object updated")
		return
	}

	if log.GetLevel() == log.DebugLevel {
		data, err := json.Marshal(&obj)
		if err != nil {
			entry.Error(err)
		}
		entry.Debug("Creating object:", string(data))
	}

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
