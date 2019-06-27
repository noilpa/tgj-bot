package database

import (
	ce "tgj-bot/customErrors"

	"tgj-bot/models"
)

func (c *Client) SaveReview(r models.Review) (err error) {
	q := `INSERT INTO reviews (mr_id, user_id, updated_at) VALUES (?, ?, ?)`
	res, err := c.db.Exec(q, r.MrID, r.UserID, r.UpdatedAt)
	if err != nil {
		return ce.ErrCreateReview
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ce.Wrap(ce.ErrCreateReview, "no affected rows")
	}
	return
}

func (c *Client) UpdateReviewApprove(r models.Review) error {
	q := `UPDATE reviews 
			SET is_approved = ?,
				updated_at = ?
		  WHERE user_id = ? 
  			AND mr_id = ?`
	res, err := c.db.Exec(q, r.IsApproved, r.UpdatedAt, r.UserID, r.MrID)
	if err != nil {
		return ce.Wrap(ce.ErrChangeReviewApprove, err.Error())
	}
	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) UpdateReviewComment(r models.Review) error {
	q := `UPDATE reviews 
			SET is_commented = ?,
				updated_at = ?
		  WHERE user_id = ? 
  			AND mr_id = ?`
	res, err := c.db.Exec(q, r.IsCommented, r.UpdatedAt, r.UserID, r.MrID)
	if err != nil {
		return ce.Wrap(ce.ErrChangeReviewComment, err.Error())
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetUserReviewMRs(uID int) (ids []int, err error) {
	q := `SELECT mr_id FROM reviews WHERE is_approved = FALSE AND user_id = ?`
	rows, err := c.db.Query(q, uID)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(&uID); err!= nil {
			return
		}
		ids = append(ids, uID)
	}
	return
}