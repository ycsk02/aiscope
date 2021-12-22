package k8s

import (
	"k8s.io/client-go/util/homedir"
	"os"
	"os/user"
	"path"
)

type KubernetesOptions struct {
	// kubeconfig path, if not specified, will use
	// in cluster way to create clientset
	KubeConfig string `json:"kubeconfig" yaml:"kubeconfig"`

	// kubernetes apiserver public address, used to generate kubeconfig
	// for downloading, default to host defined in kubeconfig
	// +optional
	Master string `json:"master,omitempty" yaml:"master"`

	// kubernetes clientset qps
	// +optional
	QPS float32 `json:"qps,omitempty" yaml:"qps"`

	// kubernetes clientset burst
	// +optional
	Burst int `json:"burst,omitempty" yaml:"burst"`
}

// NewKubernetesOptions returns a `zero` instance
func NewKubernetesOptions() (option *KubernetesOptions) {
	option = &KubernetesOptions{
		QPS:   1e6,
		Burst: 1e6,
	}

	// make it be easier for those who wants to run api-server locally
	homePath := homedir.HomeDir()
	if homePath == "" {
		// try os/user.HomeDir when $HOME is unset.
		if u, err := user.Current(); err == nil {
			homePath = u.HomeDir
		}
	}

	userHomeConfig := path.Join(homePath, ".kube/config")
	if _, err := os.Stat(userHomeConfig); err == nil {
		option.KubeConfig = userHomeConfig
	}
	return
}
