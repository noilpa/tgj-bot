package database

import (
	"database/sql"
	"testing"

	"tgj-bot/models"
	"tgj-bot/th"

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
	t.Run("should return err if user not found", func(t *testing.T) {
		f := newFixture(t)
		defer f.finish()

		_, err := f.GetUserByGitlabID(th.Int())
		assert.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)

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

func TestClient_GetUsersWithPayload(t *testing.T) {
	f := newFixture(t)
	defer f.finish()
	u := f.createUsersN(2)
	m0 := f.createMRs(u[0].ID, 3)
	m1 := f.createMR(u[1].ID)
	m0Arr := []int{m0[0].ID, m0[1].ID, m0[2].ID}

	reviews := make(map[int][]int)
	reviews[u[0].ID] = []int{m1.ID}
	reviews[u[1].ID] = m0Arr
	f.createReviews(reviews)

	ups, err := f.GetUsersWithPayload(u[0].TelegramID)
	assert.NoError(t, err)
	assert.Equal(t, len(m0Arr), ups[0].Payload)
}

func TestClient_GetUserForReallocateMR(t *testing.T) {
	f := newFixture(t)
	defer f.finish()
	u := f.createUsersN(3)
	m0 := f.createMRs(u[0].ID, 3)
	m1 := f.createMR(u[1].ID)
	m0Arr := []int{m0[0].ID, m0[1].ID, m0[2].ID}

	reviews := make(map[int][]int)
	reviews[u[0].ID] = []int{m1.ID}
	reviews[u[1].ID] = m0Arr
	f.createReviews(reviews)

	up, err := f.GetUserForReallocateMR(u[0].UserBrief, m1.ID)
	assert.NoError(t, err)
	assert.Equal(t, u[2].ID, up.ID)
}

func TestClient_GetUsersByMrURL(t *testing.T) {
	f := newFixture(t)
	defer f.finish()

	u := f.createUsersN(3)
	m := f.createMR(u[0].ID)
	reviews := make(map[int][]int)
	reviews[u[1].ID] = []int{m.ID}
	reviews[u[2].ID] = []int{m.ID}
	f.createReviews(reviews)

	us, err := f.GetUsersByMrURL(m.URL)
	assert.NoError(t, err)
	assert.Len(t, us, 2)
}