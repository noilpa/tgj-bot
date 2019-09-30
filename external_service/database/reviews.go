package database

import (
	ce "tgj-bot/custom_errors"

	"tgj-bot/models"
)

func (c *Client) SaveReview(r models.Review) (err error) {
	q := `INSERT INTO reviews (mr_id, user_id, updated_at) VALUES ($1, $2, $3)
		  ON CONFLICT ON CONSTRAINT reviews_pkey
		  DO UPDATE SET user_id = $2, updated_at = $3`
	_, err = c.db.Exec(q, r.MrID, r.UserID, r.UpdatedAt)
	if err != nil {
		err = ce.WrapWithLog(err, ce.ErrCreateUser.Error())
		return
	}

	return
}

func (c *Client) UpdateReviewApprove(r models.Review) error {
	q := `UPDATE reviews 
			SET is_approved = $1,
				updated_at = $2
		  WHERE user_id = $3 
  			AND mr_id = $4`
	_, err := c.db.Exec(q, r.IsApproved, r.UpdatedAt, r.UserID, r.MrID)
	if err != nil {
		err = ce.WrapWithLog(err, "update review approve")
		return err
	}

	return nil
}

func (c *Client) UpdateReviewComment(r models.Review) (err error) {
	q := `UPDATE reviews 
			SET is_commented = $1,
				updated_at = $2
		  WHERE user_id = $3
  			AND mr_id = $4`
	_, err = c.db.Exec(q, r.IsCommented, r.UpdatedAt, r.UserID, r.MrID)
	if err != nil {
		err = ce.WrapWithLog(err, "update review comment")
	}

	return
}

func (c *Client) GetReviewMRsByUserID(uID int) (ids []int, err error) {
	q := `SELECT mr_id FROM reviews WHERE is_approved = FALSE AND user_id = $1`
	rows, err := c.db.Query(q, uID)
	if err != nil {
		err = ce.WrapWithLog(err, "get user review mrs")
		return
	}
	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(&uID); err != nil {
			err = ce.WrapWithLog(err, "get user review mrs scan")
			return
		}
		ids = append(ids, uID)
	}
	return
}

func (c *Client) DeleteReview(r models.Review) (err error) {
	q := `DELETE FROM reviews WHERE mr_id = $1 AND user_id = $2`
	_, err = c.db.Exec(q, r.MrID, r.UserID)
	if err != nil {
		err = ce.WrapWithLog(err, "delete user review")
	}
	return
}

func (c *Client) GetOpenedReviewsByUserID(uID int) (rs []models.Review, err error) {
	q := `SELECT mr_id, user_id, is_commented, updated_at 
		  FROM reviews
		  JOIN mrs m on reviews.mr_id = m.id
		  WHERE user_id = $1 
		    AND is_approved = FALSE
		    AND m.is_closed = FALSE`
	rows, err := c.db.Query(q, uID)
	if err != nil {
		return
	}
	defer rows.Close()

	var r models.Review
	for rows.Next() {
		if err = rows.Scan(&r.MrID, &r.UserID, &r.IsCommented, &r.UpdatedAt); err != nil {
			return
		}
		rs = append(rs, r)
	}
	return
}
