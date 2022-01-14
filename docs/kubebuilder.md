
```shell
1851  go mod init aiscope
1852  kubebuilder init --domain aiscope
1853  kubebuilder edit --multigroup=true
1854  kubebuilder create api --group tenant --version v1alpha2 --kind Workspace
1855  kubebuilder create api --group iam --version v1alpha2 --kind User
1856  kubebuilder create api --group iam --version v1alpha2 --kind WorkspaceRole
1857  kubebuilder create api --group iam --version v1alpha2 --kind WorkspaceRoleBinding
1858  kubebuilder create api --group iam --version v1alpha2 --kind GlobalRole
1859  kubebuilder create api --group iam --version v1alpha2 --kind GlobalRoleBinding
      kubebuilder create api --group iam --version v1alpha2 --kind LoginRecord
1860  go mod tidy
1861  make generate
1862  make manifests



```

