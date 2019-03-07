package client

const (
	ImageBuildType            = "imageBuild"
	ImageBuildFieldBranch     = "branch"
	ImageBuildFieldCommit     = "commit"
	ImageBuildFieldDockerFile = "dockerFile"
	ImageBuildFieldHook       = "hook"
	ImageBuildFieldSecret     = "secret"
	ImageBuildFieldTag        = "tag"
	ImageBuildFieldTemplate   = "template"
	ImageBuildFieldUrl        = "url"
)

type ImageBuild struct {
	Branch     string `json:"branch,omitempty" yaml:"branch,omitempty"`
	Commit     string `json:"commit,omitempty" yaml:"commit,omitempty"`
	DockerFile string `json:"dockerFile,omitempty" yaml:"dockerFile,omitempty"`
	Hook       bool   `json:"hook,omitempty" yaml:"hook,omitempty"`
	Secret     string `json:"secret,omitempty" yaml:"secret,omitempty"`
	Tag        string `json:"tag,omitempty" yaml:"tag,omitempty"`
	Template   string `json:"template,omitempty" yaml:"template,omitempty"`
	Url        string `json:"url,omitempty" yaml:"url,omitempty"`
}
