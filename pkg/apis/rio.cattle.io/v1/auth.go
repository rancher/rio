package v1

type Permission struct {
	Role     string   `json:"role,omitempty"`
	Verbs    []string `json:"verbs,omitempty"`
	APIGroup string   `json:"apiGroup,omitempty"`
	Resource string   `json:"resource,omitempty"`
	URL      string   `json:"url,omitempty"`
	Name     string   `json:"name,omitempty"`
}
