package database

import (
	"github.com/stretchr/testify/require"
	"testing"

	"tgj-bot/models"
	"tgj-bot/th"

	"github.com/stretchr/testify/assert"
)

func TestClient_SaveMR(t *testing.T) {
	t.Run("should save", func(t *testing.T) {
		f := newFixture(t)
		defer f.finish()
		u := f.createUser()
		mr := models.MR{
			URL:      th.String(),
			AuthorID: &u.ID,
		}
		mr, err := f.CreateMR(mr)
		assert.NoError(t, err)

		actMR := f.getMR(mr.URL)
		assert.Equal(t, mr.URL, actMR.URL)
	})
	t.Run("should return err if duplicate", func(t *testing.T) {
		f := newFixture(t)
		defer f.finish()
		u := f.createUser()
		mr := models.MR{
			URL:      th.String(),
			AuthorID: &u.ID,
		}
		mr, err := f.CreateMR(mr)
		assert.NoError(t, err)

		_, err = f.CreateMR(mr)
		assert.Error(t, err)
	})
}

func TestClient_GetOpenedMRs(t *testing.T) {
	t.Run("should return all created MRs", func(t *testing.T) {
		f := newFixture(t)
		defer f.finish()
		u := f.createUser()

		expMrs := f.createMRs(u.ID, 4)

		actMrs, err := f.GetOpenedMRs()
		assert.NoError(t, err)

		assert.Len(t, actMrs, len(expMrs))
	})
	t.Run("should return all created MRs except closed", func(t *testing.T) {
		f := newFixture(t)
		defer f.finish()
		u := f.createUser()

		expMrs := f.createMRs(u.ID, 4)
		closedID := expMrs[0].ID
		f.closeMR(closedID)

		actMrs, err := f.GetOpenedMRs()
		assert.NoError(t, err)

		for _, mr := range actMrs {
			assert.NotEqual(t, closedID, mr.ID)
		}
	})
}

func TestClient_GetMrByID(t *testing.T) {
	f := newFixture(t)
	defer f.finish()
	u := f.createUser()

	eMr := f.createMR(u.ID)

	aMr, err := f.GetMrByID(eMr.ID)
	assert.NoError(t, err)
	assert.Equal(t, eMr, aMr)
}

func TestClient_CloseMRs(t *testing.T) {
	t.Run("should close mr if no reviews for it", func(t *testing.T) {
		f := newFixture(t)
		defer f.finish()

		u := f.createUser()
		eMr := f.createMR(u.ID)
		eMr.IsClosed = true

		mrs, err := f.CloseMRs()
		assert.Len(t, mrs, 1)
		assert.Equal(t, eMr, mrs[0])
		assert.NoError(t, err)

		aMr := f.getMR(eMr.URL)
		assert.True(t, aMr.IsClosed)
	})

	t.Run("should not close mr with review", func(t *testing.T) {
		f := newFixture(t)
		defer f.finish()
		u := f.createUsersN(2)
		eMr := f.createMR(u[0].ID)

		reviews := make(map[int][]int)
		reviews[u[1].ID] = []int{eMr.ID}
		f.createReviews(reviews)

		mrs, err := f.CloseMRs()
		assert.Len(t, mrs, 0)
		assert.NoError(t, err)

		aMr := f.getMR(eMr.URL)
		assert.False(t, aMr.IsClosed)
	})
}

func TestClient_GetUserClosedMRs(t *testing.T) {
	f := newFixture(t)
	defer f.finish()

	user := f.createUser()
	user3 := f.createUser()

	items := []models.MR{
		{AuthorID: &user.ID, IsClosed: true, JiraStatus: 10, URL: th.String()},
		{AuthorID: &user.ID, IsClosed: false, JiraStatus: 10, URL: th.String()},
		{AuthorID: &user3.ID, IsClosed: true, JiraStatus: 10, URL: th.String()},
		{AuthorID: &user.ID, IsClosed: true, JiraStatus: 20, URL: th.String()},
		{AuthorID: &user.ID, IsClosed: true, JiraStatus: 10, URL: th.String()},
	}

	for index, item := range items {
		newMr, err := f.CreateMR(item)
		assert.NoError(t, err)

		items[index] = newMr
	}

	expValues := []models.MR{
		items[0],
		items[4],
	}

	values, err := f.GetUserClosedMRs(user.ID, 10)
	assert.NoError(t, err)
	assert.EqualValues(t, expValues, values)
}

func TestClient_GetNeedToUpdateFromJiraMRs(t *testing.T) {
	f := newFixture(t)
	defer f.finish()

	user := f.createUser()

	mr1 := models.MR{
		URL:            th.String(),
		AuthorID:       &user.ID,
		NeedJiraUpdate: false,
	}
	mr1, err := f.CreateMR(mr1)
	require.NoError(t, err)

	mr2 := models.MR{
		URL:            th.String(),
		AuthorID:       &user.ID,
		NeedJiraUpdate: true,
	}
	mr2, err = f.CreateMR(mr2)
	require.NoError(t, err)

	mr3 := models.MR{
		URL:            th.String(),
		AuthorID:       &user.ID,
		NeedJiraUpdate: true,
	}
	mr3, err = f.CreateMR(mr3)
	require.NoError(t, err)

	mr4 := models.MR{
		URL:            th.String(),
		AuthorID:       &user.ID,
		NeedJiraUpdate: false,
	}
	mr4, err = f.CreateMR(mr4)
	require.NoError(t, err)

	items, err := f.GetNeedToUpdateFromJiraMRs()
	require.NoError(t, err)

	assert.ElementsMatch(t, []models.MR{mr2, mr3}, items)
}

func TestClient_SetNoNeedToUpdateMRsFromJira(t *testing.T) {
	t.Run("zero items", func(t *testing.T) {
		f := newFixture(t)
		defer f.finish()

		var mrs []models.MR
		assert.NoError(t, f.SetNoNeedToUpdateMRsFromJira(mrs))
	})

	t.Run("set successfully", func(t *testing.T) {
		f := newFixture(t)
		defer f.finish()

		user := f.createUser()

		mr1 := models.MR{
			URL:            th.String(),
			AuthorID:       &user.ID,
			NeedJiraUpdate: true,
		}
		mr1, err := f.CreateMR(mr1)
		require.NoError(t, err)

		mr2 := models.MR{
			URL:            th.String(),
			AuthorID:       &user.ID,
			NeedJiraUpdate: true,
		}
		mr2, err = f.CreateMR(mr2)
		require.NoError(t, err)

		mr3 := models.MR{
			URL:            th.String(),
			AuthorID:       &user.ID,
			NeedJiraUpdate: true,
		}
		mr3, err = f.CreateMR(mr3)
		require.NoError(t, err)

		mr4 := models.MR{
			URL:            th.String(),
			AuthorID:       &user.ID,
			NeedJiraUpdate: false,
		}
		mr4, err = f.CreateMR(mr4)
		require.NoError(t, err)

		require.NoError(t, f.SetNoNeedToUpdateMRsFromJira([]models.MR{mr1, mr2}))

		items, err := f.GetNeedToUpdateFromJiraMRs()
		require.NoError(t, err)

		assert.ElementsMatch(t, []models.MR{mr3}, items)
	})
}
