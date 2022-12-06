package tests

import (
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "kubevirt.io/api/core/v1"

	networkv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
)

var clientInstance client.Client

func createClient() (client.Client, error) {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(v1.AddToScheme(scheme))
	utilruntime.Must(networkv1.AddToScheme(scheme))
	ctrlOptions := ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: "0", // disable metrics
	}
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrlOptions)
	if err != nil {
		return nil, err
	}
	return client.New(mgr.GetConfig(), client.Options{Scheme: mgr.GetScheme(), Mapper: mgr.GetRESTMapper()})
}

func getClient() client.Client {
	var err error
	if clientInstance == nil {
		clientInstance, err = createClient()
		if err != nil {
			panic(err)
		}
	}
	return clientInstance
}
