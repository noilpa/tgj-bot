package database

import (
	ce "tgj-bot/customErrors"

	"tgj-bot/models"
)

func (c *Client) SaveMR(mr models.MR) (models.MR, error) {
	q := `INSERT INTO mrs (url, author_id) VALUES (?, ?);
		  SELECT id FROM mrs WHERE url = ?`
	if err := c.db.QueryRow(q, mr.URL, mr.AuthorID).Scan(&mr.ID); err != nil {
		return models.MR{}, err
	}
	return mr, nil
}

func (c *Client) GetOpenedMRs() (mrs []models.MR, err error) {
	q := `SELECT id, url, author_id FROM mrs WHERE is_closed = FALSE`
	rows, err := c.db.Query(q)
	if err != nil {
		return
	}
	defer rows.Close()

	var mr models.MR
	for rows.Next() {
		if err = rows.Scan(&mr.ID, &mr.URL, &mr.AuthorID); err != nil {
			return
		}
		mrs = append(mrs, mr)
	}
	return
}

func (c *Client) CloseMRs() error {
	q := `UPDATE mrs SET is_closed=True
		  WHERE  id NOT IN (SELECT mr_id 
						    FROM reviews 
							WHERE is_approved= FALSE);`
	_, err := c.db.Exec(q)
	if err != nil {
		return ce.Wrap(ce.ErrCloseMRs, err.Error())
	}

	return nil
}

func (c *Client) GetMrByID(id int) (mr models.MR, err error) {
	q := `SELECT id, url, author_id, is_closed FROM mrs WHERE id = ?`
	err = c.db.QueryRow(q, id).Scan(&mr.ID, &mr.URL, &mr.AuthorID, &mr.IsClosed)
	return
}
