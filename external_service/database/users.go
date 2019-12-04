package database

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	ce "tgj-bot/custom_errors"
	"tgj-bot/models"
)

func (c *Client) SaveUser(u models.User) (int, error) {
	q := `INSERT INTO  users (telegram_id, telegram_username, gitlab_id, jira_id, is_active, role, gitlab_name)
		  VALUES ($1, $2, $3, $4, $5, $6, $7)
		  ON CONFLICT ON CONSTRAINT users_telegram_username_key
		  DO UPDATE SET telegram_id = $1, role = $6, gitlab_id = $3, gitlab_name = $7
		  RETURNING id`
	err := c.db.QueryRow(q, u.TelegramID, u.TelegramUsername, u.GitlabID, u.JiraID, u.IsActive, u.Role, u.GitlabName).Scan(&u.ID)
	if err != nil {
		err = ce.WrapWithLog(err, ce.ErrCreateUser.Error())
		return 0, err
	}
	return u.ID, nil
}

func (c *Client) ChangeIsActiveUser(telegramUsername string, isActive bool) (err error) {
	q := `UPDATE users SET is_active = $1 WHERE telegram_username = $2`
	_, err = c.db.Exec(q, isActive, telegramUsername)
	if err != nil {
		err = ce.WrapWithLog(err, ce.ErrChangeUserActivity.Error())
	}
	return
}

func (c *Client) GetUsersWithPayload(exceptTelegramID string) (ups models.UsersPayload, err error) {
	q := `SELECT id, 
       			 telegram_id,
       			 telegram_username,
       			 role,
       			 gitlab_id,
       			 gitlab_name,
                 (SELECT count(*) 
                  FROM reviews r 
                  WHERE r.user_id = id
                    AND r.is_approved = FALSE) AS payload
          FROM users
		  WHERE telegram_id != $1
			AND is_active = TRUE
		  ORDER BY payload;`

	rows, err := c.db.Query(q, exceptTelegramID)
	if err != nil {
		err = ce.WrapWithLog(err, "get users with payload")
		return
	}
	defer rows.Close()

	var up models.UserPayload
	for rows.Next() {
		if err = rows.Scan(&up.ID, &up.TelegramID, &up.TelegramUsername, &up.Role, &up.GitlabID, &up.GitlabName, &up.Payload); err != nil {
			err = ce.WrapWithLog(err, "get users with payload scan")
			return
		}
		ups = append(ups, up)
	}
	return
}

func (c *Client) GetUserByTgUsername(tgUname string) (u models.User, err error) {
	q := `SELECT id, telegram_id, telegram_username, gitlab_id, jira_id, is_active, role, gitlab_name 
		  FROM users 
          WHERE telegram_username = $1`
	err = c.db.QueryRow(q, tgUname).Scan(&u.ID, &u.TelegramID, &u.TelegramUsername, &u.GitlabID, &u.JiraID, &u.IsActive, &u.Role, &u.GitlabName)
	if err != nil {
		err = ce.WrapWithLog(err, "get user by telegram username")
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
		err = ce.WrapWithLog(ce.ErrInvalidVariableType, fmt.Sprintf("%T", id))
		return
	}

	q := `SELECT id, telegram_id, telegram_username, gitlab_id, jira_id, is_active, role, gitlab_name 
		  FROM users 
          WHERE gitlab_id = $1`
	err = c.db.QueryRow(q, id).Scan(&u.ID, &u.TelegramID, &u.TelegramUsername, &u.GitlabID, &u.JiraID, &u.IsActive, &u.Role, &u.GitlabName)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("get users by gitlab id: %v:", err)
			return
		}
		err = ce.WrapWithLog(err, "get users by gitlab id")
		return
	}
	return
}

func (c *Client) GetUsersByMrID(id int) (us []models.UserBrief, err error) {
	q := `SELECT id, telegram_id, telegram_username, role, gitlab_id, gitlab_name 
		  FROM users 
		  WHERE is_active = TRUE 
		    AND id IN (SELECT user_id 
		    		  FROM reviews 
		    		  WHERE mr_id = $1)`
	rows, err := c.db.Query(q, id)
	if err != nil {
		err = ce.WrapWithLog(err, "get users by mr id")
		return
	}
	defer rows.Close()

	var u models.UserBrief
	for rows.Next() {
		if err = rows.Scan(&u.ID, &u.TelegramID, &u.TelegramUsername, &u.Role, &u.GitlabID, &u.GitlabName); err != nil {
			err = ce.WrapWithLog(err, "get users by mr id scan")
			return
		}
		us = append(us, u)
	}
	return
}

func (c *Client) GetUserForReallocateMR(u models.UserBrief, mID int) (up models.UserPayload, err error) {

	q := `SELECT id, 
       			 telegram_id,
       			 telegram_username,
       			 role, 
                 (SELECT count(*) 
                  FROM reviews r 
                  WHERE r.user_id = id
                    AND r.is_approved = FALSE) AS payload,
       			 gitlab_id, 
       			 gitlab_name
          FROM users
          WHERE is_active = TRUE 
            AND role = $1
            AND id != $2
            AND id NOT IN (SELECT user_id 
            			   FROM reviews 
            			   WHERE mr_id = $3 
            			     AND user_id != $2)
		  ORDER BY payload
		  LIMIT 1;`

	err = c.db.QueryRow(q, u.Role, u.ID, mID).Scan(&up.ID, &up.TelegramID, &up.TelegramUsername, &up.Role, &up.Payload, &u.GitlabID, &u.GitlabName)
	if err != nil {
		err = ce.WrapWithLog(err, "get user for reallocate mr")
		return
	}
	return
}

func (c *Client) GetActiveUsers() (us models.UserList, err error) {
	q := `SELECT id, telegram_id, telegram_username, gitlab_id, jira_id, is_active, role, gitlab_name FROM users WHERE is_active = TRUE`

	rows, err := c.db.Query(q)
	if err != nil {
		return
	}
	defer rows.Close()

	var u models.User
	for rows.Next() {
		if err = rows.Scan(&u.ID, &u.TelegramID, &u.TelegramUsername, &u.GitlabID, &u.JiraID, &u.IsActive, &u.Role, &u.GitlabName); err != nil {
			return
		}
		us = append(us, u)
	}
	return
}
