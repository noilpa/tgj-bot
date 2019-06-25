package gitlab_

import (
	"github.com/xanzy/go-gitlab"
)

const (
	openIssue = "opened"
)

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
	if client.Project, _, err = client.Gitlab.Projects.GetProject(cfg.ProjectID, nil); err != nil {
		return
	}

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
func (c *Client) CheckMrLikes(mrID int) (users []int, err error) {
	emojies, _, err := c.Gitlab.AwardEmoji.ListMergeRequestAwardEmoji(c.Project.ID, mrID, nil)
	if err != nil {
		return
	}
	ids := make(map[int]bool)
	for _, e := range emojies {
		if _, value := ids[e.User.ID]; !value {
			ids[e.User.ID] = true
			users = append(users, e.User.ID)
		}
	}
	return
}

// если есть открытые комметны то нотификацию получает хост МРа
// return list of users with open comment flag
func (c *Client) CheckMrComments(mrID int) (users map[int]bool, err error) {
	comments, _, err := c.Gitlab.MergeRequests.GetIssuesClosedOnMerge(c.Project.ID, mrID, nil)
	if err != nil {
		return nil, err
	}
	users = make(map[int]bool)
	for _, comment := range comments {
		users[comment.Author.ID] = users[comment.Author.ID] || isOpenIssue(comment.State)
	}
	return
}

func isOpenIssue(state string) bool {
	return state == openIssue
}
