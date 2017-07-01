package ghfilter

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/google/go-github/github"
)

func TestFilter_matches(t *testing.T) {
	filter := Filter{
		Conditions: []Condition{
			{ComparePublic: true, Public: false},
			{Type: "IssuesEvent"},
		},
	}

	tests := []struct {
		event *github.Event
		want  bool
	}{
		{
			event: &github.Event{
				Type:   github.String("IssuesEvent"),
				Public: github.Bool(false),
			},
			want: true,
		},
		{
			event: &github.Event{
				Type:   github.String("IssuesEvent"),
				Public: github.Bool(true), // we want false
			},
			want: false,
		},
		{
			event: &github.Event{
				Type:   github.String("PushEvent"), // We want IssuesEvent
				Public: github.Bool(false),
			},
			want: false,
		},
	}

	for _, test := range tests {
		have := filter.Matches(test.event)
		if have != test.want {
			t.Errorf("have: %v, want %v", have, test.want)
		}
	}
}

func TestCondition_type(t *testing.T) {
	events := []*github.Event{
		{
			Type: github.String("IssuesEvent"),
		},
		{
			Type: github.String("IssuesCommentEvent"),
		},
	}

	tests := []struct {
		Condition Condition
		Want      *github.Event
	}{
		{
			Condition: Condition{Type: "NonExistentEvent"},
			Want:      nil,
		},
		{
			Condition: Condition{Type: "IssuesEvent"},
			Want:      events[0],
		},
		{
			Condition: Condition{Type: "IssuesCommentEvent"},
			Want:      events[1],
		},
	}

	for _, test := range tests {
		for _, event := range events {
			if test.Condition.Matches(event) {
				if !reflect.DeepEqual(event, test.Want) {
					// Incorrectly matched
					t.Errorf("condition incorrectly matched\nevent: %+v\ncondition: %+v", event, test.Condition)
				}
			} else if reflect.DeepEqual(event, test.Want) {
				// Incorrectly missed
				t.Errorf("condition incorrectly missed\nevent: %+v\ncondition: %+v", event, test.Condition)
			}
		}
	}
}

func TestCondition_payloadAction(t *testing.T) {
	var (
		opened  = json.RawMessage(`{"action":"opened"}`)
		created = json.RawMessage(`{"action":"created"}`)
	)

	events := []*github.Event{
		{RawPayload: &opened},
		{RawPayload: &created},
	}

	tests := []struct {
		Condition Condition
		Want      *github.Event
	}{
		{
			Condition: Condition{PayloadAction: "opened"},
			Want:      events[0],
		},
		{
			Condition: Condition{PayloadAction: "created"},
			Want:      events[1],
		},
	}

	for _, test := range tests {
		for _, event := range events {
			if test.Condition.Matches(event) {
				if !reflect.DeepEqual(event, test.Want) {
					// Incorrectly matched
					t.Errorf("condition incorrectly matched\nevent: %+v\ncondition: %+v", event, test.Condition)
				}
			} else if reflect.DeepEqual(event, test.Want) {
				// Incorrectly missed
				t.Errorf("condition incorrectly missed\nevent: %+v\ncondition: %+v", event, test.Condition)
			}
		}
	}
}

func TestCondition_public(t *testing.T) {
	events := []*github.Event{
		{Public: github.Bool(true)},
		{Public: github.Bool(false)},
	}

	tests := []struct {
		Condition Condition
		Want      *github.Event
	}{
		{
			Condition: Condition{ComparePublic: true, Public: true},
			Want:      events[0],
		},
		{
			Condition: Condition{ComparePublic: true, Public: false},
			Want:      events[1],
		},
	}

	for _, test := range tests {
		for _, event := range events {
			if test.Condition.Matches(event) {
				if !reflect.DeepEqual(event, test.Want) {
					// Incorrectly matched
					t.Errorf("condition incorrectly matched\nevent: %+v\ncondition: %+v", event, test.Condition)
				}
			} else if reflect.DeepEqual(event, test.Want) {
				// Incorrectly missed
				t.Errorf("condition incorrectly missed\nevent: %+v\ncondition: %+v", event, test.Condition)
			}
		}
	}
}

func TestCondition_organizationID(t *testing.T) {
	events := []*github.Event{
		{Org: nil},
		{Org: &github.Organization{ID: github.Int(1)}},
		{Org: &github.Organization{ID: github.Int(2)}},
	}

	tests := []struct {
		Condition Condition
		Want      *github.Event
	}{
		{
			Condition: Condition{OrganizationID: 1},
			Want:      events[1],
		},
		{
			Condition: Condition{OrganizationID: 2},
			Want:      events[2],
		},
	}

	for _, test := range tests {
		for _, event := range events {
			if test.Condition.Matches(event) {
				if !reflect.DeepEqual(event, test.Want) {
					// Incorrectly matched
					t.Errorf("condition incorrectly matched\nevent: %+v\ncondition: %+v", event, test.Condition)
				}
			} else if reflect.DeepEqual(event, test.Want) {
				// Incorrectly missed
				t.Errorf("condition incorrectly missed\nevent: %+v\ncondition: %+v", event, test.Condition)
			}
		}
	}
}

func TestCondition_repositoryID(t *testing.T) {
	events := []*github.Event{
		{Repo: nil},
		{Repo: &github.Repository{ID: github.Int(1)}},
		{Repo: &github.Repository{ID: github.Int(2)}},
	}

	tests := []struct {
		Condition Condition
		Want      *github.Event
	}{
		{
			Condition: Condition{RepositoryID: 1},
			Want:      events[1],
		},
		{
			Condition: Condition{RepositoryID: 2},
			Want:      events[2],
		},
	}

	for _, test := range tests {
		for _, event := range events {
			if test.Condition.Matches(event) {
				if !reflect.DeepEqual(event, test.Want) {
					// Incorrectly matched
					t.Errorf("condition incorrectly matched\nevent: %+v\ncondition: %+v", event, test.Condition)
				}
			} else if reflect.DeepEqual(event, test.Want) {
				// Incorrectly missed
				t.Errorf("condition incorrectly missed\nevent: %+v\ncondition: %+v", event, test.Condition)
			}
		}
	}
}
