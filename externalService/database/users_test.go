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
		assert.NoError(t, f.SaveUser(u))

		actU := f.getUser(u.TelegramUsername)
		u.ID = actU.ID
		assert.Equal(t, u, actU)
	})
	t.Run("should update user", func(t *testing.T) {
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
		assert.NoError(t, f.SaveUser(u))
		u.GitlabID = th.String()
		assert.NoError(t, f.SaveUser(u))
		actU := f.getUser(u.TelegramUsername)
		u.ID = actU.ID
		assert.Equal(t, u, actU)
	})
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