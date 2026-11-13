package collector

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Client struct {
	Clientset *kubernetes.Clientset
	RestConfig *rest.Config
	Logger    *zap.Logger
}

func NewClient(kubeconfigPath, contextName string) (*Client, error) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	var config *rest.Config
	var err error

	if kubeconfigPath != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to build config from kubeconfig: %w", err)
		}
	} else if contextName != "" {
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		configOverrides := &clientcmd.ConfigOverrides{
			CurrentContext: contextName,
		}
		kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
		config, err = kubeConfig.ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to build config for context %s: %w", contextName, err)
		}
	} else {
		kubeconfig := ""
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
		if _, err := os.Stat(kubeconfig); err == nil {
			config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
			if err != nil {
				return nil, fmt.Errorf("failed to build config from default kubeconfig: %w", err)
			}
		} else {
			config, err = rest.InClusterConfig()
			if err != nil {
				return nil, fmt.Errorf("failed to build in-cluster config: %w", err)
			}
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	return &Client{
		Clientset:  clientset,
		RestConfig: config,
		Logger:     logger,
	}, nil
}

func (c *Client) GetClientset() *kubernetes.Clientset {
	return c.Clientset
}

func (c *Client) GetRestConfig() *rest.Config {
	return c.RestConfig
}

func (c *Client) TestConnection(ctx context.Context) error {
	_, err := c.Clientset.Discovery().ServerVersion()
	return err
}