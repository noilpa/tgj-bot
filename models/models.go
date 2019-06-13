package models


type UserBrief struct {
	ID int
	TelegramID string
	Role       Role
}

type User struct {
	UserBrief
	GitlabID   string
	JiraID     string
	IsActive   bool
}

type UserPayload struct {
	UserBrief
	Payload int
}

type UsersPayload []UserPayload

func (ups *UsersPayload) GetN(num int, role Role) UsersPayload {

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
