module aiscope

go 1.17

require (
	github.com/emicklei/go-restful v2.15.0+incompatible
	github.com/spf13/cobra v1.3.0
	k8s.io/klog/v2 v2.40.1
	sigs.k8s.io/controller-runtime v0.11.0
)

require (
	github.com/go-logr/logr v1.2.0
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.17.0
	github.com/pkg/errors v0.9.1
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5
	k8s.io/api v0.23.0
	k8s.io/apiextensions-apiserver v0.23.0
	k8s.io/apimachinery v0.23.0
	k8s.io/apiserver v0.23.0
	k8s.io/client-go v0.23.0
)

require k8s.io/klog v1.0.0

replace (
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.22.5
	k8s.io/apimachinery => k8s.io/apimachinery v0.22.5
	k8s.io/apiserver => k8s.io/apiserver v0.22.5
	k8s.io/client-go => k8s.io/client-go v0.22.5
	k8s.io/component-base => k8s.io/component-base v0.22.5
)
