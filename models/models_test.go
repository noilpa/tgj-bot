package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetGitlabID(t *testing.T) {
	mr := "https://git.itv.restr.im/itv-backend/itv-api-ng/merge_requests/2392/pipelines#db7a71f25fb8f66d309e5c4726d920ecdd492523"
	_, err := GetGitlabID(mr)
	assert.Error(t, err)
}