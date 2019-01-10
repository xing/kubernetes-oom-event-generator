package util

import (
	"os"

	"github.com/golang/glog"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // needed for local development with .kube/config
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Clientset abstracts the cluster config loading both locally and on Kubernetes
func Clientset() kubernetes.Interface {
	// Try to load in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		glog.V(3).Infof("Could not load in-cluster config: %v", err)

		// Fall back to local config
		config, err = clientcmd.BuildConfigFromFlags("", os.Getenv("HOME")+"/.kube/config")
		if err != nil {
			glog.Fatalf("Failed to load client config: %v", err)
		}
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatalf("Failed to create kubernetes client: %v", err)
	}

	return client
}
