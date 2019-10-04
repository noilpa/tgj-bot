package gitlab_

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/davecgh/go-spew/spew"
	"github.com/xanzy/go-gitlab"

	ce "tgj-bot/custom_errors"
	"tgj-bot/models"
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
	emojies, resp, err := c.Gitlab.AwardEmoji.ListMergeRequestAwardEmoji(c.Project.ID, mrID, nil)
	respBody, _ := ioutil.ReadAll(resp.Body)
	log.Println("Emojies resp:", string(respBody))
	if err != nil {
		return
	}
	log.Println("Emojies:", emojies)

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
	comments, resp, err := c.Gitlab.MergeRequests.GetIssuesClosedOnMerge(c.Project.ID, mrID, nil)
	respBody, _ := ioutil.ReadAll(resp.Body)
	log.Println("Comments resp:", string(respBody))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println("Comments:", comments)

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

func (c *Client) GetUserByID(gitlabID int) (name string, err error) {
	user, _, err := c.Gitlab.Users.GetUser(gitlabID)
	log.Println("Get user by gitlab id:", user)
	if err != nil {
		return
	}
	return user.Username, nil
}

func (c *Client) GetUserByName(gitlabName string) (id int, err error){
	options := &gitlab.ListUsersOptions{Username: &gitlabName}
	userList, _, err := c.Gitlab.Users.ListUsers(options)
	log.Println("Get user by gitlab name:", userList)
	if err != nil {
		return
	}
	if len(userList) > 1 {
		return 0, errors.New(fmt.Sprintf("more than one matches found by name=%s\ntry to use gitlab id instead", gitlabName))
	}
	return userList[0].ID, nil
}

func (c *Client) WriteReviewers(mrID int, reviewers models.UsersPayload) error {
	description, err := c.getMrDescription(mrID)
	spew.Dump(description, err)
	if err != nil {
		ce.WrapWithLog(err, "get mr description fail")
		return err
	}
	description += "\n\n// Reviewers: "
	for _, r := range reviewers {
		description += fmt.Sprintf("@%s ", r.GitlabName)
	}
	description += "//"
	opt := &gitlab.UpdateMergeRequestOptions{Description: &description}
	_, _, err = c.Gitlab.MergeRequests.UpdateMergeRequest(c.Project.ID, mrID, opt)
	spew.Dump(opt, c.Project.ID, mrID)
	spew.Dump(description, err)

	return err
}

func (c *Client) getMrDescription(mrID int) (description string, err error) {
	mr, resp, err := c.Gitlab.MergeRequests.GetMergeRequest(c.Project.ID, mrID, nil)
	respBody, _ := ioutil.ReadAll(resp.Body)
	log.Println("Get mr description:", string(respBody))
	if err != nil {
		return
	}

	return mr.Description, nil
}
