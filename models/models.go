package models

import (
	"net/url"
	"strconv"
	"strings"

	ce "tgj-bot/custom_errors"
)

type UserBrief struct {
	ID               int
	TelegramID       string
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

const (
	Developer = Role("dev")
	Lead      = Role("lead")
)

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
	ID       int
	URL      string
	AuthorID *int
	IsClosed bool
	GitlabID int
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
	mrID, err := strconv.Atoi(pathArr[len(pathArr)-1])
	return mrID, err
}
