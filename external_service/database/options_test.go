package database

import (
	"encoding/json"
	"testing"

	"tgj-bot/models"
	"tgj-bot/th"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_UpdateOptionByName(t *testing.T) {
	t.Run("should return error if name is invalid", func(t *testing.T) {
		f := newFixture(t)
		defer f.finish()

		_, err := f.LoadOptionByName(th.String())
		require.Error(t, err)
	})
	t.Run("should update", func(t *testing.T) {
		f := newFixture(t)
		defer f.finish()

		expValue := models.LastSendNotifyOption{Stamp: th.Int64()}

		err := f.UpdateOptionByName(models.OptionLastSendNotify, expValue)
		require.NoError(t, err)

		option, err := f.LoadOptionByName(models.OptionLastSendNotify)
		require.NoError(t, err)

		data, err := json.Marshal(expValue)
		require.NoError(t, err)

		assert.EqualValues(t, string(data), option.Item)
	})
}
