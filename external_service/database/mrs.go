package database

import (
	ce "tgj-bot/custom_errors"

	"tgj-bot/models"
)

func (c *Client) CreateMR(mr models.MR) (models.MR, error) {
	q := `INSERT INTO mrs (url, author_id) VALUES ($1, $2) RETURNING id`
	err := c.db.QueryRow(q, mr.URL, mr.AuthorID).Scan(&mr.ID)
	if err != nil {
		err = ce.WrapWithLog(err, "create mr")
		return mr, err
	}
	return mr, nil
}

func (c *Client) SaveMR(mr models.MR) (models.MR, error) {
	q := `UPDATE mrs SET is_closed=$2, jira_id=$3, jira_priority=$4 WHERE id=$1`
	_, err := c.db.Exec(q, mr.ID, mr.IsClosed, mr.JiraID, mr.JiraPriority)
	if err != nil {
		err = ce.WrapWithLog(err, "save mr")
		return mr, err
	}
	return mr, nil
}

func (c *Client) GetOpenedMRs() (mrs []models.MR, err error) {
	q := `SELECT id, url, author_id, is_closed, jira_id, jira_priority FROM mrs WHERE is_closed = FALSE`
	rows, err := c.db.Query(q)
	if err != nil {
		err = ce.WrapWithLog(err, "get opened mrs")
		return
	}
	defer rows.Close()

	var mr models.MR
	for rows.Next() {
		if err = rows.Scan(&mr.ID, &mr.URL, &mr.AuthorID, &mr.IsClosed, &mr.JiraID, &mr.JiraPriority); err != nil {
			err = ce.WrapWithLog(err, "get opened mrs")
			return
		}
		mrs = append(mrs, mr)
	}
	return
}

func (c *Client) CloseMRs() (mrs []models.MR, err error) {
	q := `UPDATE mrs SET is_closed=True
		  WHERE  id NOT IN (SELECT DISTINCT(mr_id) 
						    FROM reviews 
							WHERE is_approved= FALSE)
			AND id NOT IN (SELECT id 
						   FROM mrs 
			    		   WHERE is_closed=True)			
		  RETURNING id, url, author_id, is_closed;`
	rows, err := c.db.Query(q)
	if err != nil {
		err = ce.WrapWithLog(ce.ErrCloseMRs, err.Error())
		return
	}
	defer rows.Close()

	var mr models.MR
	for rows.Next() {
		if err = rows.Scan(&mr.ID, &mr.URL, &mr.AuthorID, &mr.IsClosed); err != nil {
			err = ce.WrapWithLog(err, "get closed mrs id")
			return nil, err
		}
		mrs = append(mrs, mr)
	}

	return
}

func (c *Client) CloseMR(id int) error {
	q := `UPDATE mrs SET is_closed=True
		  WHERE  id = $1`
	_, err := c.db.Exec(q, id)
	if err != nil {
		err = ce.WrapWithLog(ce.ErrCloseMRs, err.Error())
		return err
	}
	return nil
}

func (c *Client) GetMrByID(id int) (mr models.MR, err error) {
	q := `SELECT id, url, author_id, jira_id, jira_priority, is_closed FROM mrs WHERE id = $1`
	err = c.db.QueryRow(q, id).Scan(&mr.ID, &mr.URL, &mr.AuthorID, &mr.JiraID, &mr.JiraPriority, &mr.IsClosed)
	if err != nil {
		err = ce.WrapWithLog(err, "get mr by id")
	}
	return
}

func (c *Client) GetMRbyURL(url string) (mr models.MR, err error) {
	q := `SELECT id, url, author_id, is_closed FROM mrs WHERE url = $1`
	err = c.db.QueryRow(q, url).Scan(&mr.ID, &mr.URL, &mr.AuthorID, &mr.IsClosed)
	return
}
