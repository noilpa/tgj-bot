package gitlab_

import (
	"github.com/xanzy/go-gitlab"
)

type GitlabConfig struct {
	Token string `json:"token"`
}

type Client struct {
	Gitlab *gitlab.Client
}

func RunGitlab(cfg GitlabConfig) (client Client, err error) {
	client.Gitlab = gitlab.NewClient(nil, "yourtokengoeshere")
	return
}