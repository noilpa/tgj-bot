package models

import (
	ce "tgj-bot/customErrors"
)

type UserBrief struct {
	ID               int
	TelegramID       string
	TelegramUsername string
	Role             Role
}

type User struct {
	UserBrief
	GitlabID string
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
	AuthorID int
	IsClosed bool
}

type Review struct {
	MrID        int
	UserID      int
	IsApproved  bool
	IsCommented bool
	UpdatedAt   int64
}
