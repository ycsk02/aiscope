package api

const (
	WorkspaceNone = ""
	ClusterNone   = ""
)

const (
	StatusOK = "ok"
)

type ListResult struct {
	Items      []interface{} `json:"items"`
	TotalItems int           `json:"totalItems"`
}
