package client

const (
	FeatureSpecType             = "featureSpec"
	FeatureSpecFieldAnswers     = "answers"
	FeatureSpecFieldDescription = "description"
	FeatureSpecFieldEnable      = "enable"
	FeatureSpecFieldQuestions   = "questions"
)

type FeatureSpec struct {
	Answers     map[string]string `json:"answers,omitempty" yaml:"answers,omitempty"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	Enable      bool              `json:"enable,omitempty" yaml:"enable,omitempty"`
	Questions   []Question        `json:"questions,omitempty" yaml:"questions,omitempty"`
}
