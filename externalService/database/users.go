package database

import (
	"fmt"
	"strconv"

	ce "tgj-bot/customErrors"
	"tgj-bot/models"
)

func (c *Client) SaveUser(user models.User) (err error) {
	q := `REPLACE INTO  users (telegram_id, telegram_username, gitlab_id, jira_id, is_active, role)
		  VALUES (?, ?, ?, ?, ?, ?)`
	_, err = c.db.Exec(q, user.TelegramID, user.TelegramUsername, user.GitlabID, user.JiraID, user.IsActive, user.Role)
	if err != nil {
		err = ce.WrapWithLog(err, ce.ErrCreateUser.Error())
	}
	return
}

func (c *Client) ChangeIsActiveUser(telegramUsername string, isActive bool) (err error) {
	q := `UPDATE users SET is_active = ? WHERE telegram_username = ?`
	_, err = c.db.Exec(q, isActive, telegramUsername)
	if err != nil {
		err = ce.WrapWithLog(err, ce.ErrChangeUserActivity.Error())
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
		err = ce.WrapWithLog(err, "get users with payload")
		return
	}
	defer rows.Close()

	var up models.UserPayload
	for rows.Next() {
		if err = rows.Scan(&up.ID, &up.TelegramID, &up.TelegramUsername, &up.Role, &up.Payload); err != nil {
			err = ce.WrapWithLog(err, "get users with payload scan")
			return
		}
		ups = append(ups, up)
	}
	return
}

func (c *Client) GetUserByTgUsername(tgUname string) (u models.User, err error) {
	q := `SELECT id, telegram_id, telegram_username, gitlab_id, jira_id, is_active, role 
		  FROM users 
          WHERE telegram_username = ?`
	err = c.db.QueryRow(q, tgUname).Scan(&u.ID, &u.TelegramID, &u.TelegramUsername, &u.GitlabID, &u.JiraID, &u.IsActive, &u.Role)
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

	q := `SELECT id, telegram_id, telegram_username, gitlab_id, jira_id, is_active, role 
		  FROM users 
          WHERE gitlab_id = ?`
	err = c.db.QueryRow(q, id).Scan(&u.ID, &u.TelegramID, &u.TelegramUsername, &u.GitlabID, &u.JiraID, &u.IsActive, &u.Role)
	if err != nil {
		err = ce.WrapWithLog(err, "get users by gitlab id")
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
		err = ce.WrapWithLog(err, "get users by mr id")
		return
	}
	defer rows.Close()

	var u models.UserBrief
	for rows.Next() {
		if err = rows.Scan(&u.ID, &u.TelegramID, &u.TelegramUsername, &u.Role); err != nil {
			err = ce.WrapWithLog(err, "get users by mr id scan")
			return
		}
		us = append(us, u)
	}
	return
}

func (c *Client) GetUserForReallocateMR(uID, mID int, role models.Role) (up models.UserPayload, err error) {

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

	err = c.db.QueryRow(q, role, uID, mID, uID).Scan(&up.ID, &up.TelegramID, &up.TelegramUsername, &up.Role, &up.Payload)
	if err != nil {
		err = ce.WrapWithLog(err, "get user for reallocate mr")
		return
	}
	return
}