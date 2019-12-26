package database

import (
	ce "tgj-bot/custom_errors"
	"tgj-bot/models"

	"github.com/lib/pq"
)

func (c *Client) GetAllMRs() (mrs []models.MR, err error) {
	q := `SELECT id, url, author_id, is_closed, jira_id, jira_priority, jira_status, 
       gitlab_id, need_jira_update, need_qa_notify, gitlab_labels FROM mrs`
	rows, err := c.db.Query(q)
	if err != nil {
		err = ce.WrapWithLog(err, "get all mrs")
		return
	}
	defer rows.Close()

	var mr models.MR
	for rows.Next() {
		if err = rows.Scan(&mr.ID, &mr.URL, &mr.AuthorID, &mr.IsClosed, &mr.JiraID, &mr.JiraPriority,
			&mr.JiraStatus, &mr.GitlabID, &mr.NeedJiraUpdate, &mr.NeedQANotify, pq.Array(&mr.GitlabLabels)); err != nil {
			err = ce.WrapWithLog(err, "get all mrs")
			return
		}
		mrs = append(mrs, mr)
	}
	return
}

func (c *Client) GetNeedToUpdateFromJiraMRs() (mrs []models.MR, err error) {
	q := `SELECT id, url, author_id, is_closed, jira_id, jira_priority, jira_status, gitlab_id, 
       need_jira_update, need_qa_notify, gitlab_labels
			FROM mrs WHERE need_jira_update=True`
	rows, err := c.db.Query(q)
	if err != nil {
		err = ce.WrapWithLog(err, "get need to update from jira mrs")
		return
	}
	defer rows.Close()

	var mr models.MR
	for rows.Next() {
		if err = rows.Scan(&mr.ID, &mr.URL, &mr.AuthorID, &mr.IsClosed, &mr.JiraID, &mr.JiraPriority,
			&mr.JiraStatus, &mr.GitlabID, &mr.NeedJiraUpdate, &mr.NeedQANotify, pq.Array(&mr.GitlabLabels)); err != nil {
			err = ce.WrapWithLog(err, "get need to update from jira mrs")
			return
		}
		mrs = append(mrs, mr)
	}
	return
}

func (c *Client) CreateMR(mr models.MR) (models.MR, error) {
	q := `INSERT INTO mrs (url, author_id, gitlab_id, is_closed, jira_id, jira_priority, jira_status, 
                 need_jira_update, need_qa_notify, gitlab_labels) 
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id`
	err := c.db.QueryRow(q, mr.URL, mr.AuthorID, mr.GitlabID, mr.IsClosed, mr.JiraID, mr.JiraPriority,
		mr.JiraStatus, mr.NeedJiraUpdate, mr.NeedQANotify, pq.Array(mr.GitlabLabels)).Scan(&mr.ID)
	if err != nil {
		err = ce.WrapWithLog(err, "create mr")
		return mr, err
	}
	return mr, nil
}

func (c *Client) SaveMR(mr models.MR) (models.MR, error) {
	q := `UPDATE mrs SET is_closed=$2, jira_id=$3, jira_priority=$4, jira_status=$5, gitlab_id=$6, 
               need_jira_update=$7, need_qa_notify=$8, gitlab_labels=$9 WHERE id=$1`
	_, err := c.db.Exec(q, mr.ID, mr.IsClosed, mr.JiraID, mr.JiraPriority, mr.JiraStatus,
		mr.GitlabID, mr.NeedJiraUpdate, mr.NeedQANotify, pq.Array(mr.GitlabLabels))
	if err != nil {
		err = ce.WrapWithLog(err, "save mr")
		return mr, err
	}
	return mr, nil
}

func (c *Client) GetOpenedMRs() (mrs []models.MR, err error) {
	q := `SELECT id, url, author_id, is_closed, jira_id, jira_priority, jira_status, gitlab_id, 
       need_jira_update, need_qa_notify, gitlab_labels 
			FROM mrs WHERE is_closed = FALSE`
	rows, err := c.db.Query(q)
	if err != nil {
		err = ce.WrapWithLog(err, "get opened mrs")
		return
	}
	defer rows.Close()

	var mr models.MR
	for rows.Next() {
		if err = rows.Scan(&mr.ID, &mr.URL, &mr.AuthorID, &mr.IsClosed, &mr.JiraID, &mr.JiraPriority,
			&mr.JiraStatus, &mr.GitlabID, &mr.NeedJiraUpdate, &mr.NeedQANotify, pq.Array(&mr.GitlabLabels)); err != nil {
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
		  RETURNING id, url, author_id, is_closed, gitlab_id, jira_status, need_jira_update, need_qa_notify, gitlab_labels;`
	rows, err := c.db.Query(q)
	if err != nil {
		err = ce.WrapWithLog(ce.ErrCloseMRs, err.Error())
		return
	}
	defer rows.Close()

	var mr models.MR
	for rows.Next() {
		if err = rows.Scan(&mr.ID, &mr.URL, &mr.AuthorID, &mr.IsClosed, &mr.GitlabID, &mr.JiraStatus,
			&mr.NeedJiraUpdate, &mr.NeedQANotify, pq.Array(&mr.GitlabLabels)); err != nil {
			err = ce.WrapWithLog(err, "get closed mrs id")
			return nil, err
		}
		mrs = append(mrs, mr)
	}

	return
}

func (c *Client) CloseMR(id int) error {
	q := `UPDATE mrs SET is_closed=True WHERE id = $1`
	_, err := c.db.Exec(q, id)
	if err != nil {
		err = ce.WrapWithLog(ce.ErrCloseMRs, err.Error())
		return err
	}
	return nil
}

func (c *Client) GetMrByID(id int) (mr models.MR, err error) {
	q := `SELECT id, url, author_id, jira_id, jira_priority, jira_status, is_closed, 
       gitlab_id, need_jira_update, need_qa_notify, gitlab_labels FROM mrs WHERE id = $1`
	err = c.db.QueryRow(q, id).Scan(&mr.ID, &mr.URL, &mr.AuthorID, &mr.JiraID, &mr.JiraPriority,
		&mr.JiraStatus, &mr.IsClosed, &mr.GitlabID, &mr.NeedJiraUpdate, &mr.NeedQANotify, pq.Array(&mr.GitlabLabels))
	if err != nil {
		err = ce.WrapWithLog(err, "get mr by id")
	}
	return
}

func (c *Client) GetMRbyURL(url string) (mr models.MR, err error) {
	q := `SELECT id, url, author_id, is_closed, gitlab_id FROM mrs WHERE url = $1`
	err = c.db.QueryRow(q, url).Scan(&mr.ID, &mr.URL, &mr.AuthorID, &mr.IsClosed, &mr.GitlabID)
	return
}

func (c *Client) GetUserClosedMRs(uID int, jiraStatus int) (mrs []models.MR, err error) {
	q := `SELECT id, url, author_id, is_closed, gitlab_id, jira_id, jira_priority, jira_status
		  FROM mrs WHERE author_id=$1 AND is_closed=True AND jira_status=$2 ORDER by jira_priority DESC`
	rows, err := c.db.Query(q, uID, jiraStatus)
	if err != nil {
		return
	}
	defer rows.Close()

	var mr models.MR
	for rows.Next() {
		if err = rows.Scan(&mr.ID, &mr.URL, &mr.AuthorID, &mr.IsClosed, &mr.GitlabID, &mr.JiraID, &mr.JiraPriority, &mr.JiraStatus); err != nil {
			return
		}
		mrs = append(mrs, mr)
	}
	return
}
