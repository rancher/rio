package v1

type Question struct {
	Variable          string        `json:"variable,omitempty"`
	Label             string        `json:"label,omitempty"`
	Description       string        `json:"description,omitempty"`
	Type              string        `json:"type,omitempty"`
	Required          bool          `json:"required,omitempty"`
	Default           string        `json:"default,omitempty"`
	Group             string        `json:"group,omitempty"`
	MinLength         int           `json:"minLength,omitempty"`
	MaxLength         int           `json:"maxLength,omitempty"`
	Min               int           `json:"min,omitempty"`
	Max               int           `json:"max,omitempty"`
	Options           []string      `json:"options,omitempty"`
	ValidChars        string        `json:"validChars,omitempty"`
	InvalidChars      string        `json:"invalidChars,omitempty"`
	Subquestions      []SubQuestion `json:"subquestions,omitempty"`
	ShowIf            string        `json:"showIf,omitempty"`
	ShowSubquestionIf string        `json:"showSubquestionIf,omitempty"`
}

type SubQuestion struct {
	Variable     string   `json:"variable,omitempty"`
	Label        string   `json:"label,omitempty"`
	Description  string   `json:"description,omitempty"`
	Type         string   `json:"type,omitempty"`
	Required     bool     `json:"required,omitempty"`
	Default      string   `json:"default,omitempty"`
	Group        string   `json:"group,omitempty"`
	MinLength    int      `json:"minLength,omitempty"`
	MaxLength    int      `json:"maxLength,omitempty"`
	Min          int      `json:"min,omitempty"`
	Max          int      `json:"max,omitempty"`
	Options      []string `json:"options,omitempty"`
	ValidChars   string   `json:"validChars,omitempty"`
	InvalidChars string   `json:"invalidChars,omitempty"`
	ShowIf       string   `json:"showIf,omitempty"`
}

type TemplateMeta struct {
	Name      string     `json:"name,omitempty"`
	Version   string     `json:"version,omitempty"`
	IconURL   string     `json:"iconUrl,omitempty"`
	Readme    string     `json:"readme,omitempty"`
	Questions []Question `json:"questions,omitempty"`
}
