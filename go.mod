module aiscope

go 1.17

require (
	github.com/emicklei/go-restful v2.15.0+incompatible
	github.com/spf13/cobra v1.3.0
	k8s.io/klog/v2 v2.40.1
	sigs.k8s.io/controller-runtime v0.11.0
)

require (
	github.com/emicklei/go-restful-openapi v1.4.1
	github.com/go-ldap/ldap v3.0.3+incompatible
	github.com/go-logr/logr v1.2.0
	github.com/google/go-cmp v0.5.6
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.17.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/pflag v1.0.5
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5
	k8s.io/api v0.23.0
	k8s.io/apiextensions-apiserver v0.23.0
	k8s.io/apimachinery v0.23.0
	k8s.io/apiserver v0.23.0
	k8s.io/client-go v0.23.0
	k8s.io/klog v1.0.0
)

require (
	github.com/traefik/traefik/v2 v2.5.7
	gopkg.in/asn1-ber.v1 v1.0.0-20181015200546-f715ec2f112d // indirect
)

require (
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/mitchellh/mapstructure v1.4.3
)

replace (
	github.com/abbot/go-http-auth => github.com/containous/go-http-auth v0.4.1-0.20200324110947-a37a7636d23e
	github.com/go-check/check => github.com/containous/check v0.0.0-20170915194414-ca0bf163426a
	github.com/gorilla/mux => github.com/containous/mux v0.0.0-20220113180107-8ffa4f6d063c
	github.com/mailgun/minheap => github.com/containous/minheap v0.0.0-20190809180810-6e71eb837595
	github.com/mailgun/multibuf => github.com/containous/multibuf v0.0.0-20190809014333-8b6c9a7e6bba
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.22.5
	k8s.io/apimachinery => k8s.io/apimachinery v0.22.5
	k8s.io/apiserver => k8s.io/apiserver v0.22.5
	k8s.io/client-go => k8s.io/client-go v0.22.5
	k8s.io/component-base => k8s.io/component-base v0.22.5
)
