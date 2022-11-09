/*
Copyright 2022.

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
	"context"
	"flag"
	"os"
	"time"

	"github.com/AlonaKaplan/kubesecondarydns/controllers"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	v1 "kubevirt.io/api/core/v1"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(v1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	opts := zap.Options{}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	ctrlOptions := ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: "0", // disable metrics
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrlOptions)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	waitForKubevirt()

	// TODO zoneManager := manager.New()
	if err = (&controllers.VirtualMachineInstanceReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("VirtualMachineInstance"),
		Scheme: mgr.GetScheme(),
		// TODO ZoneMananger: zoneManager
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "VirtualMachineInstance")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func checkForKubevirt(kubeClient *kubernetes.Clientset) error {
	result := kubeClient.ExtensionsV1beta1().RESTClient().Get().RequestURI("/apis/apiextensions.k8s.io/v1/customresourcedefinitions/virtualmachineinstances.kubevirt.io").Do(context.TODO())
	return result.Error()
}

// Check for Kubevirt CRD to be available
func waitForKubevirt() {
	clientset, err := kubernetes.NewForConfig(ctrl.GetConfigOrDie())
	if err != nil {
		setupLog.Error(err, "unable to create a kubernetes client error")
		os.Exit(1)
	}
	for _ = range time.Tick(5 * time.Second) {
		err := checkForKubevirt(clientset)
		if err != nil {
			setupLog.Info("kubevirt doesn't exist in the cluster", "err", err)
		}
		if err == nil {
			setupLog.Info("kubevirt exists in the cluster")
			break
		}
	}
}
