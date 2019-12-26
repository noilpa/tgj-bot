package models

import (
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	ce "tgj-bot/custom_errors"
	"tgj-bot/external_service/jira"
)

const (
	Developer = Role("dev")
	Lead      = Role("lead")
)

const (
	GitlabLabelWIP      = "WIP"
	GitlabLabelServices = "services"
	GitlabLabelReviewed = "reviewed"
)

var jiraRegExp = regexp.MustCompile(`\[NC-([0-9]+)\]*`)

type UserBrief struct {
	ID         int
	TelegramID string
	// always lowercase
	TelegramUsername string
	Role             Role
	GitlabName       string
	GitlabID         int
	JiraID           string
}

type User struct {
	UserBrief
	JiraID   string
	IsActive bool
}

type UserList []User

type UserPayload struct {
	UserBrief
	Payload int
}

type UsersPayload []UserPayload

func (ups UsersPayload) GetN(num int, role Role) (res UsersPayload, err error) {
	if len(ups) == 0 {
		return res, ce.ErrUsersForReviewNotFound
	}
	// users already are sorted from db
	for _, up := range ups {
		if num == 0 {
			break
		}
		if up.Role != role {
			continue
		}
		res = append(res, up)
		num--
	}
	return
}

type Role string

var ValidRoles = [...]Role{Developer, Lead}

func IsValidRole(r Role) bool {
	for _, role := range ValidRoles {
		if role == r {
			return true
		}
	}
	return false
}

type MR struct {
	ID             int
	URL            string
	AuthorID       *int
	IsClosed       bool
	GitlabID       int
	JiraID         int
	JiraPriority   int
	JiraStatus     int
	NeedJiraUpdate bool
	NeedQANotify   bool
	GitlabLabels   []string
}

func (mr *MR) ExtractJiraID(title string) {
	matches := jiraRegExp.FindStringSubmatch(title)
	if len(matches) == 2 {
		jiraID, err := strconv.Atoi(matches[1])
		if err == nil {
			mr.JiraID = jiraID
		}
	}
}

func (mr *MR) IsHighest() bool {
	return jira.PriorityHighest == mr.JiraPriority ||
		jira.PriorityHigh == mr.JiraPriority
}

func (mr *MR) IsOnReview() bool {
	return mr.JiraStatus == jira.StatusOnReview
}

func (mr *MR) CheckNoNeedUpdateFromJira() {
	if mr.JiraStatus == jira.StatusMerged ||
		mr.JiraStatus == jira.StatusApproved ||
		mr.JiraStatus == jira.StatusReady ||
		mr.JiraStatus == jira.StatusDone {
		mr.NeedJiraUpdate = false
	}
}

func (mr *MR) IsWIP() bool {
	for _, label := range mr.GitlabLabels {
		if GitlabLabelWIP == label {
			return true
		}
	}

	return false
}

type Review struct {
	MrID        int
	UserID      int
	IsApproved  bool
	IsCommented bool
	UpdatedAt   int64
}

func GetGitlabID(mrURL string) (int, error) {
	url_, err := url.Parse(mrURL)
	if err != nil {
		return 0, err
	}
	pathArr := strings.Split(url_.Path, "/")
	if len(pathArr) < 4 {
		return 0, errors.New("wrong url format")
	}
	return strconv.Atoi(pathArr[4])
}

const (
	OptionLastSendNotify = "last_send_notify"
)

type LastSendNotifyOption struct {
	Stamp int64 `json:"value"`
}

type Option struct {
	ID        int
	Name      string
	Item      string `json:"item"`
	UpdatedAt *time.Time
}
