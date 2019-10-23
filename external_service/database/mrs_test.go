package database

import (
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
		mr, err := f.SaveMR(mr)
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
		mr, err := f.SaveMR(mr)
		assert.NoError(t, err)

		_, err = f.SaveMR(mr)
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
