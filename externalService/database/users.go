package database

import (
	"fmt"
	"strconv"

	ce "tgj-bot/customErrors"
	"tgj-bot/models"
)

func (c *Client) SaveUser(user models.User) (err error) {
	q := `INSERT INTO users (telegram_id, telegram_username, gitlab_id, jira_id, is_active, role)
		  VALUES (?, ?, ?, ?, ?, ?)`
	res, err := c.db.Exec(q, user.TelegramID, user.TelegramUsername, user.GitlabID, user.JiraID, user.IsActive, user.Role)
	if err != nil {
		return ce.ErrCreateUser
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ce.Wrap(ce.ErrCreateUser, "no affected rows")
	}
	return
}

func (c *Client) ChangeIsActiveUser(telegramUsername string, isActive bool) (err error) {
	q := `UPDATE users SET is_active = ? WHERE telegram_username = ?`
	res, err := c.db.Exec(q, isActive, telegramUsername)
	if err != nil {
		return ce.Wrap(ce.ErrChangeUserActivity, err.Error())
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ce.Wrap(ce.ErrChangeUserActivity, "no affected rows")
	}
	return
}

func (c *Client) GetUsersWithPayload(telegramID string) (ups models.UsersPayload, err error) {
	q := `SELECT id, 
       			 telegram_id,
       			 telegram_username,
       			 role, 
                 (SELECT count(*) 
                  FROM reviews r 
                  WHERE r.user_id = id
                    AND r.is_approved = FALSE) AS payload
          FROM users
		  WHERE telegram_id != ?
			AND is_active = TRUE
		  ORDER BY payload;`

	rows, err := c.db.Query(q, telegramID)
	if err != nil {
		return
	}
	defer rows.Close()

	var up models.UserPayload
	for rows.Next() {
		if err = rows.Scan(&up.ID, &up.TelegramID, &up.TelegramUsername, &up.Role, &up.Payload); err != nil {
			return
		}
		ups = append(ups, up)
	}
	return
}

func (c *Client) GetUsers() (users models.Users, err error) {
	q := `SELECT id, telegram_id, telegram_username, gitlab_id, jira_id, is_active, role 
		  FROM users`

	rows, err := c.db.Query(q)
	if err != nil {
		return
	}
	defer rows.Close()

	var u models.User
	for rows.Next() {
		if err = rows.Scan(&u.ID, &u.TelegramID, &u.TelegramUsername, &u.GitlabID, &u.JiraID, &u.IsActive, &u.Role); err != nil {
			return
		}
		users = append(users, u)
	}
	return
}

func (c *Client) GetUserByTgUsername(id string) (u models.User, err error) {
	q := `SELECT id, telegram_id, telegram_username, gitlab_id, jira_id, is_active, role 
		  FROM users 
          WHERE telegram_username = ?`
	err = c.db.QueryRow(q, id).Scan(&u.ID, &u.TelegramID, &u.TelegramUsername, &u.GitlabID, &u.JiraID, &u.IsActive, &u.Role)
	if err != nil {
		return
	}
	return
}

func (c *Client) GetUserByGitlabID(id interface{}) (u models.User, err error) {
	switch id.(type) {
	case int:
		id = strconv.Itoa(id.(int))
	case string:
		id = id.(string)
	default:
		err = ce.Wrap(ce.ErrInvalidVariableType, fmt.Sprintf("%T", id))
		return
	}

	q := `SELECT id, telegram_id, telegram_username, gitlab_id, jira_id, is_active, role 
		  FROM users 
          WHERE gitlab_id = ?`
	err = c.db.QueryRow(q, id).Scan(&u.ID, &u.TelegramID, &u.TelegramUsername, &u.GitlabID, &u.JiraID, &u.IsActive, &u.Role)
	if err != nil {
		return
	}
	return
}

func (c *Client) GetUsersByMrID(id int) (us []models.UserBrief, err error) {
	q := `SELECT id, telegram_id, telegram_username, role 
		  FROM users 
		  WHERE is_active = TRUE 
		    AND id IN (SELECT user_id 
		    		  FROM reviews 
		    		  WHERE mr_id = ?)`
	rows, err := c.db.Query(q, id)
	if err != nil {
		return
	}
	defer rows.Close()

	var u models.UserBrief
	for rows.Next() {
		if err = rows.Scan(&u.ID, &u.TelegramID, &u.TelegramUsername, &u.Role); err != nil {
			return
		}
		us = append(us, u)
	}
	return
}

func (c *Client) GetUserForReallocateMR(uID, mID int, role models.Role) (ups models.UserPayload, err error) {

	q := `SELECT id, 
       			 telegram_id,
       			 telegram_username,
       			 role, 
                 (SELECT count(*) 
                  FROM reviews r 
                  WHERE r.user_id = id
                    AND r.is_approved = FALSE) AS payload
          FROM users
          WHERE is_active = TRUE 
            AND role = ?
            AND id != ?
            AND id NOT IN (SELECT user_id 
            			   FROM reviews 
            			   WHERE mr_id = ? 
            			     AND user_id != ?)
		  ORDER BY payload
		  LIMIT 1;`

	rows, err := c.db.Query(q, role, uID, mID, uID)
	if err != nil {
		return
	}
	defer rows.Close()

	var up models.UserPayload
	for rows.Next() {
		if err = rows.Scan(&up.ID, &up.TelegramID, &up.TelegramUsername, &up.Role, &up.Payload); err != nil {
			return
		}
		ups = append(ups, up)
	}

	return
}