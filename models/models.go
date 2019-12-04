package models

import (
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	ce "tgj-bot/custom_errors"
	"tgj-bot/external_service/jira"
)

const (
	Developer = Role("dev")
	Lead      = Role("lead")
)

const (
	ReviewedLabel = "reviewed"
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
	ID           int
	URL          string
	AuthorID     *int
	IsClosed     bool
	GitlabID     int
	JiraID       int
	JiraPriority int
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
