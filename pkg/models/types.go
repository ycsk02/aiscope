package models

type PageableResponse struct {
	Items      []interface{} `json:"items" description:"paging data"`
	TotalCount int           `json:"total_count" description:"total count"`
}

type Workspace struct {
	Group          `json:",inline"`
	Admin          string   `json:"admin,omitempty"`
	Namespaces     []string `json:"namespaces"`
	DevopsProjects []string `json:"devops_projects"`
}

type Group struct {
	Path        string   `json:"path"`
	Name        string   `json:"name"`
	Gid         string   `json:"gid"`
	Members     []string `json:"members"`
	Logo        string   `json:"logo"`
	ChildGroups []string `json:"child_groups"`
	Description string   `json:"description"`
}

type PodInfo struct {
	Namespace string `json:"namespace" description:"namespace"`
	Pod       string `json:"pod" description:"pod name"`
	Container string `json:"container" description:"container name"`
}
