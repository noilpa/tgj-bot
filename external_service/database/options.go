package database

import (
	"encoding/json"
	"fmt"

	ce "tgj-bot/custom_errors"
	"tgj-bot/models"
)

func (c *Client) LoadOptionByName(name string) (option models.Option, err error) {
	q := `SELECT id, name, item, updated_at FROM options WHERE name = $1`
	err = c.db.QueryRow(q, name).Scan(&option.ID, &option.Name, &option.Item, &option.UpdatedAt)
	if err != nil {
		err = ce.WrapWithLog(err, fmt.Sprintf("get option by name: %s", name))
	}
	return
}

func (c *Client) UpdateOptionByName(name string, item interface{}) error {
	value, err := json.Marshal(item)
	if err != nil {
		return err
	}

	q := `UPDATE options SET item=$2 WHERE name=$1`
	_, err = c.db.Exec(q, name, string(value))
	if err != nil {
		return ce.WrapWithLog(err, fmt.Sprintf("save option by name: %s", name))
	}
	return nil
}
