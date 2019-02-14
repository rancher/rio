package client

const (
	ImageBuildType                    = "imageBuild"
	ImageBuildFieldBranch             = "branch"
	ImageBuildFieldCommit             = "commit"
	ImageBuildFieldDockerFile         = "dockerFile"
	ImageBuildFieldImageName          = "imageName"
	ImageBuildFieldImageTag           = "imageTag"
	ImageBuildFieldTag                = "tag"
	ImageBuildFieldTemplate           = "template"
	ImageBuildFieldUrl                = "url"
	ImageBuildFieldWebhookAccessToken = "webhookAccessToken"
)

type ImageBuild struct {
	Branch             string `json:"branch,omitempty" yaml:"branch,omitempty"`
	Commit             string `json:"commit,omitempty" yaml:"commit,omitempty"`
	DockerFile         string `json:"dockerFile,omitempty" yaml:"dockerFile,omitempty"`
	ImageName          string `json:"imageName,omitempty" yaml:"imageName,omitempty"`
	ImageTag           string `json:"imageTag,omitempty" yaml:"imageTag,omitempty"`
	Tag                string `json:"tag,omitempty" yaml:"tag,omitempty"`
	Template           string `json:"template,omitempty" yaml:"template,omitempty"`
	Url                string `json:"url,omitempty" yaml:"url,omitempty"`
	WebhookAccessToken string `json:"webhookAccessToken,omitempty" yaml:"webhookAccessToken,omitempty"`
}
