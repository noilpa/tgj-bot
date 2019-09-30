package gitlab_

import (
	"log"

	"github.com/davecgh/go-spew/spew"
	"github.com/xanzy/go-gitlab"
)

const (
	opened = "opened"
	closed = "closed"
	locked = "locked"
	merged = "merged"
)

type GitlabService interface {
	CheckMrLikes(mrID int) (users map[int]struct{}, err error)
	CheckMrComments(mrID int) (users map[int]struct{}, err error)
	GetMrAuthorID(mrID int) (int, error)
	MrIsOpen(mrID int) (bool, error)
}

type GitlabConfig struct {
	Token     string `json:"token"`
	ProjectID string `json:"project_id"`
}

type Client struct {
	Gitlab  *gitlab.Client
	Project *gitlab.Project
}

func RunGitlab(cfg GitlabConfig) (client Client, err error) {
	client.Gitlab = gitlab.NewClient(nil, cfg.Token)

	if err = client.Gitlab.SetBaseURL("https://git.itv.restr.im/"); err != nil {
		return
	}
	log.Printf("Gitlab BaseURL: %v", client.Gitlab.BaseURL().String())
	if client.Project, _, err = client.Gitlab.Projects.GetProject(cfg.ProjectID, nil); err != nil {
		return
	}
	log.Printf("Gitlab Project: %v", client.Project.String())
	return
}

//
// посмотреть комменты к МРу -> список юзеров
// посмотреть лайки -> список юзеров
// если все комменты закрыты но лайка нет -> нотификация ревьюеру
// если есть открытые комменты -> нотификация автору МРа
// добавить в таблицу МР колонку автор_ид
//

// return list of users with emoji on mr
func (c *Client) CheckMrLikes(mrID int) (users map[int]struct{}, err error) {
	emojies, _, err := c.Gitlab.AwardEmoji.ListMergeRequestAwardEmoji(c.Project.ID, mrID, nil)
	if err != nil {
		return
	}
	users = make(map[int]struct{})
	for _, e := range emojies {
		if _, found := users[e.User.ID]; !found {
			users[e.User.ID] = struct{}{}
		}
	}
	return
}

// если есть открытые комметны то нотификацию получает хост МРа
// return list of users with open comment flag
func (c *Client) CheckMrComments(mrID int) (users map[int]struct{}, err error) {
	comments, _, err := c.Gitlab.MergeRequests.GetIssuesClosedOnMerge(c.Project.ID, mrID, nil)
	if err != nil {
		return nil, err
	}
	users = make(map[int]struct{})
	for _, comment := range comments {
		if _, found := users[comment.Author.ID]; !found {
			users[comment.Author.ID] = struct{}{}
		}
	}
	return
}

func (c *Client) GetMrAuthorID(mrID int) (int, error) {
	spew.Dump(c)
	mr, _, err := c.Gitlab.MergeRequests.GetMergeRequest(c.Project.ID, mrID, nil)
	if err != nil {
		return 0, err
	}
	return mr.Author.ID, nil
}

func (c *Client) MrIsOpen(mrID int) (bool, error) {
	mr, _, err := c.Gitlab.MergeRequests.GetMergeRequest(c.Project.ID, mrID, nil)
	if err != nil {
		return false, err
	}
	log.Printf("Get MR from gitlab %d: %v\n", mrID, mr)
	return mr.State == opened, nil
}
