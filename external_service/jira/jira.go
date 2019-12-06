package jira

import (
	"context"
	"github.com/pkg/errors"
	"strconv"

	"github.com/andygrunwald/go-jira"
)

const (
	PriorityUndefined = 0
	PriorityLowest    = 5
	PriorityLow       = 7
	PriorityMedium    = 10
	PriorityHigh      = 15
	PriorityHighest   = 20
)

const (
	StatusUndefined  = 0
	StatusTrash      = 1
	StatusAnalysis   = 5
	StatusBacklog    = 10
	StatusTODO       = 20
	StatusReopened   = 30
	StatusInProgress = 40
	StatusOnReview   = 50
	StatusReadyForQA = 60
	StatusTesting    = 70
	StatusApproved   = 80
	StatusMerged     = 90
	StatusReady      = 100
	StatusDone       = 110
)

type Config struct {
	BaseURL     string `json:"base_url"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	UpdateTasks bool   `json:"update_tasks"`
}

type Issue struct {
	ID       int
	Priority int
	Status   int
}

type IJira interface {
	LoadIssueByID(ctx context.Context, ID int) (*Issue, error)
}

type Jira struct {
	conf   Config
	client *jira.Client
}

func NewJira(conf Config) (*Jira, error) {
	transport := jira.BasicAuthTransport{
		Username: conf.Username,
		Password: conf.Password,
	}

	client, err := jira.NewClient(transport.Client(), conf.BaseURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed init jira client")
	}

	if conf.UpdateTasks {
		// ping Jira
		if _, _, err := client.User.GetSelf(); err != nil {
			return nil, errors.Wrapf(err, "jira not available")
		}
	}

	return &Jira{
		conf:   conf,
		client: client,
	}, nil
}

func (jir *Jira) LoadIssueByID(ctx context.Context, ID int) (*Issue, error) {
	issue, _, err := jir.client.Issue.Get(jir.taskAlias(ID), nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed load jira issue by ID:%d", ID)
	}

	item := &Issue{
		ID:       ID,
		Priority: jir.getPriorityValue(issue.Fields.Priority.Name),
		Status:   jir.getStatusValue(issue.Fields.Status.Name),
	}
	return item, nil
}

func (jir *Jira) taskAlias(ID int) string {
	return "NC-" + strconv.Itoa(ID)
}

func (jir *Jira) getPriorityValue(name string) int {
	switch name {
	case "Highest":
		return PriorityHighest
	case "High":
		return PriorityHigh
	case "Medium":
		return PriorityMedium
	case "Low":
		return PriorityLow
	case "Lowest":
		return PriorityLowest
	}

	return PriorityUndefined
}

func (jir *Jira) getStatusValue(name string) int {
	switch name {
	case "Backlog":
		return StatusBacklog
	case "Trash":
		return StatusTrash
	case "Analysis":
		return StatusAnalysis
	case "To Do":
		return StatusTODO
	case "Reopened":
		return StatusReopened
	case "In Progress":
		return StatusInProgress
	case "ON REVIEW":
		return StatusOnReview
	case "Ready for QA":
		return StatusReadyForQA
	case "Testing":
		return StatusTesting
	case "Approved":
		return StatusApproved
	case "Merged":
		return StatusMerged
	case "Ready":
		return StatusReady
	case "Done":
		return StatusDone
	}

	return StatusUndefined
}
