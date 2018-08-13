package client

const (
	PermissionType          = "permission"
	PermissionFieldAPIGroup = "apiGroup"
	PermissionFieldName     = "name"
	PermissionFieldResource = "resource"
	PermissionFieldRole     = "role"
	PermissionFieldURL      = "url"
	PermissionFieldVerbs    = "verbs"
)

type Permission struct {
	APIGroup string   `json:"apiGroup,omitempty" yaml:"apiGroup,omitempty"`
	Name     string   `json:"name,omitempty" yaml:"name,omitempty"`
	Resource string   `json:"resource,omitempty" yaml:"resource,omitempty"`
	Role     string   `json:"role,omitempty" yaml:"role,omitempty"`
	URL      string   `json:"url,omitempty" yaml:"url,omitempty"`
	Verbs    []string `json:"verbs,omitempty" yaml:"verbs,omitempty"`
}
