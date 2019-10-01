package database

import (
	"testing"

	"tgj-bot/models"
	"tgj-bot/th"

	"github.com/stretchr/testify/assert"
)

func TestClient_SaveReview(t *testing.T) {
	f := newFixture(t)
	defer f.finish()
	u := f.createUser()
	mr := f.createMR(u.ID)
	r := models.Review{
		MrID:      mr.ID,
		UserID:    u.ID,
		UpdatedAt: th.Int64(),
	}

	assert.NoError(t, f.SaveReview(r))

	actR := f.getReviewsByMR(mr.ID)[0]
	assert.Equal(t, r, actR)
	actR = f.getReviewsByUser(u.ID)[0]
	assert.Equal(t, r, actR)
}

func TestClient_UpdateReviewApprove(t *testing.T) {
	f := newFixture(t)
	defer f.finish()
	u := f.createUser()
	eMr := f.createMR(u.ID)

	reviews := make(map[int][]int)
	reviews[u.ID] = []int{eMr.ID}
	r := f.createReviews(reviews)[0]

	r.IsApproved = !r.IsApproved

	assert.NoError(t, f.UpdateReviewApprove(r))

	actR := f.getReviewsByUser(u.ID)[0]
	assert.Equal(t, r, actR)
}

func TestClient_UpdateReviewComment(t *testing.T) {
	f := newFixture(t)
	defer f.finish()
	u := f.createUser()
	eMr := f.createMR(u.ID)

	reviews := make(map[int][]int)
	reviews[u.ID] = []int{eMr.ID}
	r := f.createReviews(reviews)[0]

	r.IsCommented = !r.IsCommented

	assert.NoError(t, f.UpdateReviewComment(r))

	actR := f.getReviewsByUser(u.ID)[0]
	assert.Equal(t, r, actR)
}

func TestClient_GetUserReviewMRs(t *testing.T) {
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

	ids, err := f.GetReviewMRsByUserID(u[1].ID)
	assert.NoError(t, err)
	assert.Len(t, ids, len(m0))
	for _, id := range ids {
		assert.True(t, isContain(m0Arr, id))
	}
}
