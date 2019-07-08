package database

import (
	ce "tgj-bot/customErrors"

	"tgj-bot/models"
)

func (c *Client) SaveReview(r models.Review) (err error) {
	q := `INSERT INTO reviews (mr_id, user_id, updated_at) VALUES (?, ?, ?)`
	_, err = c.db.Exec(q, r.MrID, r.UserID, r.UpdatedAt)
	if err != nil {
		err = ce.WrapWithLog(err, ce.ErrCreateUser.Error())
		return
	}

	return
}

func (c *Client) UpdateReviewApprove(r models.Review) error {
	q := `UPDATE reviews 
			SET is_approved = ?,
				updated_at = ?
		  WHERE user_id = ? 
  			AND mr_id = ?`
	_, err := c.db.Exec(q, r.IsApproved, r.UpdatedAt, r.UserID, r.MrID)
	if err != nil {
		err = ce.WrapWithLog(err, "update review approve")
		return err
	}

	return nil
}

func (c *Client) UpdateReviewComment(r models.Review) (err error) {
	q := `UPDATE reviews 
			SET is_commented = ?,
				updated_at = ?
		  WHERE user_id = ? 
  			AND mr_id = ?`
	_, err = c.db.Exec(q, r.IsCommented, r.UpdatedAt, r.UserID, r.MrID)
	if err != nil {
		err = ce.WrapWithLog(err, "update review comment")
	}

	return
}

func (c *Client) GetReviewMRsByUserID(uID int) (ids []int, err error) {
	q := `SELECT mr_id FROM reviews WHERE is_approved = FALSE AND user_id = ?`
	rows, err := c.db.Query(q, uID)
	if err != nil {
		err = ce.WrapWithLog(err, "get user review mrs")
		return
	}
	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(&uID); err!= nil {
			err = ce.WrapWithLog(err, "get user review mrs scan")
			return
		}
		ids = append(ids, uID)
	}
	return
}

func (c *Client) DeleteReview(r models.Review) (err error) {
	q := `DELETE FROM reviews WHERE mr_id = ? AND user_id = ?`
	_, err = c.db.Exec(q, r.MrID, r.UserID)
	if err != nil {
		err = ce.WrapWithLog(err, "delete user review")
	}
	return
}