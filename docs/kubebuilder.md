
```shell
go mod init aiscope
kubebuilder init --domain aiscope
kubebuilder edit --multigroup=true
kubebuilder create api --group tenant --version v1alpha2 --kind Workspace
kubebuilder create api --group iam --version v1alpha2 --kind User
kubebuilder create api --group iam --version v1alpha2 --kind WorkspaceRole
kubebuilder create api --group iam --version v1alpha2 --kind WorkspaceRoleBinding
kubebuilder create api --group iam --version v1alpha2 --kind GlobalRole
kubebuilder create api --group iam --version v1alpha2 --kind GlobalRoleBinding
kubebuilder create api --group iam --version v1alpha2 --kind LoginRecord
kubebuilder create api --group experiment --version v1alpha2 --kind TrackingServer
kubebuilder create api --group experiment --version v1alpha2 --kind JupyterNotebook
kubebuilder create api --group experiment --version v1alpha2 --kind CodeServer

go mod tidy
make generate
make manifests



```

