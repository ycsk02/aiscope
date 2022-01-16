
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
go mod tidy
make generate
make manifests



```

