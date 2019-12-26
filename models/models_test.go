package models

import (
	"testing"

	"tgj-bot/external_service/jira"
)

func TestMR_CheckNoNeedUpdateFromJira(t *testing.T) {
	tests := []struct {
		Status int
		NoNeed bool
	}{
		{Status: jira.StatusUndefined, NoNeed: true},
		{Status: jira.StatusTrash, NoNeed: true},
		{Status: jira.StatusAnalysis, NoNeed: true},
		{Status: jira.StatusBacklog, NoNeed: true},
		{Status: jira.StatusTODO, NoNeed: true},
		{Status: jira.StatusReopened, NoNeed: true},
		{Status: jira.StatusInProgress, NoNeed: true},
		{Status: jira.StatusOnReview, NoNeed: true},
		{Status: jira.StatusReadyForQA, NoNeed: true},
		{Status: jira.StatusTesting, NoNeed: true},
		{Status: jira.StatusApproved, NoNeed: false},
		{Status: jira.StatusMerged, NoNeed: false},
		{Status: jira.StatusReady, NoNeed: false},
		{Status: jira.StatusDone, NoNeed: false},
	}

	mr := MR{}

	for index, item := range tests {
		mr.JiraStatus = item.Status
		mr.NeedJiraUpdate = true
		mr.CheckNoNeedUpdateFromJira()

		if item.NoNeed != mr.NeedJiraUpdate {
			t.Errorf("failed at index:%d %v vs %v", index, item.NoNeed, mr.NeedJiraUpdate)
		}
	}
}

func TestMR_IsWIP(t *testing.T) {
	tests := []struct {
		labels []string
		exp    bool
	}{
		{nil, false},
		{[]string{"foo", "bar"}, false},
		{[]string{"foo", "bar", "WIP"}, true},
	}

	mr := MR{}

	for index, item := range tests {
		mr.GitlabLabels = item.labels

		if item.exp != mr.IsWIP() {
			t.Errorf("failed at index:%d %v vs %v", index, item.exp, mr.IsWIP())
		}
	}
}

func TestMR_IsLabelsEqual(t *testing.T) {
	tests := []struct {
		gitlabLabels   []string
		comparedLabels []string
		exp            bool
	}{
		{exp: true},
		{exp: false, gitlabLabels: []string{"boo", "bar"}},
		{exp: true, gitlabLabels: []string{"boo", "bar"}, comparedLabels: []string{"boo", "bar"}},
		{exp: true, gitlabLabels: []string{"boo", "bar"}, comparedLabels: []string{"bar", "boo"}},
		{exp: false, gitlabLabels: []string{"boo", "bar"}, comparedLabels: []string{"tmp"}},
	}

	mr := MR{}

	for index, item := range tests {
		mr.GitlabLabels = item.gitlabLabels

		value := mr.IsLabelsEqual(item.comparedLabels)
		if item.exp != value {
			t.Errorf("failed at index:%d %v vs %v", index, item.gitlabLabels, item.comparedLabels)
		}
	}
}
