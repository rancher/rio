package client

const (
	StackSpecType                           = "stackSpec"
	StackSpecFieldAdditionalFiles           = "additionalFiles"
	StackSpecFieldAnswers                   = "answers"
	StackSpecFieldDescription               = "description"
	StackSpecFieldDisableMesh               = "disableMesh"
	StackSpecFieldEnableKubernetesResources = "enableKubernetesResources"
	StackSpecFieldQuestions                 = "questions"
	StackSpecFieldRepoTags                  = "repoTags"
	StackSpecFieldTemplate                  = "template"
)

type StackSpec struct {
	AdditionalFiles           map[string]string `json:"additionalFiles,omitempty" yaml:"additionalFiles,omitempty"`
	Answers                   map[string]string `json:"answers,omitempty" yaml:"answers,omitempty"`
	Description               string            `json:"description,omitempty" yaml:"description,omitempty"`
	DisableMesh               bool              `json:"disableMesh,omitempty" yaml:"disableMesh,omitempty"`
	EnableKubernetesResources bool              `json:"enableKubernetesResources,omitempty" yaml:"enableKubernetesResources,omitempty"`
	Questions                 []Question        `json:"questions,omitempty" yaml:"questions,omitempty"`
	RepoTags                  []string          `json:"repoTags,omitempty" yaml:"repoTags,omitempty"`
	Template                  string            `json:"template,omitempty" yaml:"template,omitempty"`
}
