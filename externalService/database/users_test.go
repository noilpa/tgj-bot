package database

import (
	"testing"

	"tgj-bot/models"
	th "tgj-bot/testHelpers"

	"github.com/stretchr/testify/assert"
)



func TestClient_SaveUser(t *testing.T) {
	t.Run("should create user", func(t *testing.T) {
		f := newFixture(t)
		defer f.finish()
		u := models.User{
			UserBrief: models.UserBrief{
				TelegramID:       th.String(),
				TelegramUsername: th.String(),
				Role:             th.Role(),
			},
			GitlabID: th.String(),
			JiraID:   th.String(),
			IsActive: th.Bool(),
		}
		var err error
		u.ID, err = f.SaveUser(u)
		assert.NoError(t, err)

		actU := f.getUser(u.TelegramUsername)
		assert.Equal(t, u, actU)
	})
}

func TestClient_UpdateUser(t *testing.T) {
	f := newFixture(t)
	defer f.finish()
	u := f.createUser()
	u.TelegramUsername = "new name"

	assert.NoError(t, f.UpdateUser(u))
	actU := f.getUser(u.TelegramUsername)

	assert.Equal(t, u, actU)
}

func TestClient_ChangeIsActiveUser(t *testing.T) {
	t.Run("should change is_active", func(t *testing.T) {
		f := newFixture(t)
		defer f.finish()
		expUsers := f.createUsersN(4)

		for _, u := range expUsers {
			assert.NoError(t, f.ChangeIsActiveUser(u.TelegramUsername, !u.IsActive))
			actUser := f.getUser(u.TelegramUsername)
			u.ID = actUser.ID
			assert.Equal(t, u.IsActive, !actUser.IsActive)
		}
	})
	t.Run("should change is_active to same value", func(t *testing.T) {
		f := newFixture(t)
		defer f.finish()
		expUsers := f.createUsersN(4)

		for _, u := range expUsers {
			assert.NoError(t, f.ChangeIsActiveUser(u.TelegramUsername, u.IsActive))
			actUser := f.getUser(u.TelegramUsername)
			u.ID = actUser.ID
			assert.Equal(t, u, actUser)
		}
	})
}

func TestClient_GetUserByTgUsername(t *testing.T) {
	f := newFixture(t)
	defer f.finish()

	expU := f.createUser()
	actU, err := f.GetUserByTgUsername(expU.TelegramUsername)
	assert.NoError(t, err)
	expU.ID = actU.ID
	assert.Equal(t, expU, actU)
}

func TestClient_GetUserByGitlabID(t *testing.T) {
	t.Run("should get by string gitlab id", func(t *testing.T) {
		f := newFixture(t)
		defer f.finish()

		expU := f.createUser()
		actU, err := f.GetUserByGitlabID(expU.GitlabID)
		assert.NoError(t, err)
		expU.ID = actU.ID
		assert.Equal(t, expU, actU)
	})
}

func TestClient_GetUsersByMrID(t *testing.T) {
	f := newFixture(t)
	defer f.finish()

	u := f.createUsersN(3)
	m := f.createMR(u[0].ID)
	reviews := make(map[int][]int)
	reviews[u[1].ID] = []int{m.ID}
	reviews[u[2].ID] = []int{m.ID}
	f.createReviews(reviews)

	ubs, err := f.GetUsersByMrID(m.ID)
	assert.NoError(t, err)
	assert.Len(t, ubs, len(reviews))
	for _, ub := range ubs {
		assert.True(t, isContain([]int{u[1].ID, u[2].ID}, ub.ID))
	}
}