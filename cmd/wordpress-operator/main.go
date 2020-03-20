/*
Copyright 2018 Pressinfra SRL.

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

package main

import (
	"os"

	logf "github.com/presslabs/controller-util/log"
	flag "github.com/spf13/pflag"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"

	"github.com/presslabs/wordpress-operator/pkg/apis"
	"github.com/presslabs/wordpress-operator/pkg/cmd/options"
	"github.com/presslabs/wordpress-operator/pkg/controller"
)

const genericErrorExitCode = 1

var setupLog = logf.Log.WithName("wordpress-operator")

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	options.AddToFlagSet(fs)
	err := fs.Parse(os.Args)
	if err != nil {
		setupLog.Error(err, "unable to parse args")
		os.Exit(genericErrorExitCode)
	}

	development := os.Getenv("DEV") == "true"
	logf.SetLogger(logf.ZapLogger(development))

	setupLog.Info("Starting wordpress-operator...")

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		setupLog.Error(err, "unable to get configuration")
		os.Exit(genericErrorExitCode)
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{})
	if err != nil {
		setupLog.Error(err, "unable to create a new manager")
		os.Exit(genericErrorExitCode)
	}

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		setupLog.Error(err, "unable to register types to scheme")
		os.Exit(genericErrorExitCode)
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		setupLog.Error(err, "unable to setup controllers")
		os.Exit(genericErrorExitCode)
	}

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "unable to start the manager")
		os.Exit(genericErrorExitCode)
	}
}
