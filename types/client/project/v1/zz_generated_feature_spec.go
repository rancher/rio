package client

const (
	FeatureSpecType             = "featureSpec"
	FeatureSpecFieldAnswers     = "answers"
	FeatureSpecFieldDescription = "description"
	FeatureSpecFieldEnabled     = "enable"
	FeatureSpecFieldQuestions   = "questions"
)

type FeatureSpec struct {
	Answers     map[string]string `json:"answers,omitempty" yaml:"answers,omitempty"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	Enabled     bool              `json:"enable,omitempty" yaml:"enable,omitempty"`
	Questions   []Question        `json:"questions,omitempty" yaml:"questions,omitempty"`
}
