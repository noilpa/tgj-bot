package jira

import (
	"context"
	"github.com/pkg/errors"
	"strconv"

	"github.com/andygrunwald/go-jira"
)

const (
	PriorityUndefined = 0
	PriorityNormal    = 10
	PriorityHighest   = 20
)

type Config struct {
	BaseURL  string `json:"base_url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Issue struct {
	ID       int
	Priority int
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

	priority := PriorityUndefined
	if issue.Fields.Priority.Name == "Highest" {
		priority = PriorityHighest
	} else {
		priority = PriorityNormal
	}

	item := &Issue{
		ID:       ID,
		Priority: priority,
	}
	return item, nil
}

func (jir *Jira) taskAlias(ID int) string {
	return "NC-" + strconv.Itoa(ID)
}
