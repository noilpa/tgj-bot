package database

import (
	ce "tgj-bot/custom_errors"

	"tgj-bot/models"
)

func (c *Client) SaveMR(mr models.MR) (models.MR, error) {
	q := `INSERT INTO mrs (url, author_id) VALUES (?, ?)`
	res, err := c.db.Exec(q, mr.URL, mr.AuthorID, mr.URL)
	if err != nil {
		err = ce.WrapWithLog(err, "save mr")
		return mr, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		err = ce.WrapWithLog(err, "save mr")
		return mr, err
	}
	mr.ID = int(id)
	return mr, nil
}

func (c *Client) GetOpenedMRs() (mrs []models.MR, err error) {
	q := `SELECT id, url, author_id FROM mrs WHERE is_closed = FALSE`
	rows, err := c.db.Query(q)
	if err != nil {
		err = ce.WrapWithLog(err, "get opened mrs")
		return
	}
	defer rows.Close()

	var mr models.MR
	for rows.Next() {
		if err = rows.Scan(&mr.ID, &mr.URL, &mr.AuthorID); err != nil {
			err = ce.WrapWithLog(err, "get opened mrs")
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
		err = ce.WrapWithLog(ce.ErrCloseMRs, err.Error())
		return err
	}

	return nil
}

func (c *Client) CloseMR(id int) error {
	q := `UPDATE mrs SET is_closed=True
		  WHERE  id = ?`
	_, err := c.db.Exec(q)
	if err != nil {
		err = ce.WrapWithLog(ce.ErrCloseMRs, err.Error())
		return err
	}
	return nil
}

func (c *Client) GetMrByID(id int) (mr models.MR, err error) {
	q := `SELECT id, url, author_id, is_closed FROM mrs WHERE id = ?`
	err = c.db.QueryRow(q, id).Scan(&mr.ID, &mr.URL, &mr.AuthorID, &mr.IsClosed)
	if err != nil {
		err = ce.WrapWithLog(err, "get mr by id")
	}
	return
}

func (c *Client) GetMRbyURL(url string) (mr models.MR, err error) {
	q := `SELECT id, url, author_id, is_closed FROM main.mrs WHERE url = ?`
	err = c.db.QueryRow(q, url).Scan(&mr.ID, &mr.URL, &mr.AuthorID, &mr.IsClosed)
	return
}


