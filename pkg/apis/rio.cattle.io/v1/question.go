package v1

type Question struct {
	// The variable name to reference using ${...} syntax
	Variable string `json:"variable,omitempty"`

	// A friend name for the question
	Label string `json:"label,omitempty"`

	// A longer description of the question
	Description string `json:"description,omitempty"`

	// The field type: string, int, bool, enum. default is string
	Type string `json:"type,omitempty"`

	// The answer can not be blank
	Required bool `json:"required,omitempty"`

	// Default value of the answer if not specified by the user
	Default string `json:"default,omitempty"`

	// Group the question with questions in the same group (Most used by UI)
	Group string `json:"group,omitempty"`

	// Minimum length of the answer
	MinLength int `json:"minLength,omitempty"`

	// Maximum length of the answer
	MaxLength int `json:"maxLength,omitempty"`

	// Minimum value of an int answer
	Min int `json:"min,omitempty"`

	// Maximum value of an int answer
	Max int `json:"max,omitempty"`

	// An array of valid answers for type enum questions
	Options []string `json:"options,omitempty"`

	// Answer must be composed of only these characters
	ValidChars string `json:"validChars,omitempty"`

	// Answer must not have any of these characters
	InvalidChars string `json:"invalidChars,omitempty"`

	// A list of questions that are considered child questions
	Subquestions []SubQuestion `json:"subquestions,omitempty"`

	// Ask question only if this evaluates to true, more info on syntax below
	ShowIf string `json:"showIf,omitempty"`

	// Ask subquestions if this evaluates to true
	ShowSubquestionIf string `json:"showSubquestionIf,omitempty"`
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
	Name       string     `json:"name,omitempty"`
	Version    string     `json:"version,omitempty"`
	IconURL    string     `json:"iconUrl,omitempty"`
	Readme     string     `json:"readme,omitempty"`
	Questions  []Question `json:"questions,omitempty"`
	GoTemplate bool       `json:"goTemplate,omitempty"`
	EnvSubst   bool       `json:"envSubst,omitempty"`
}
