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

func TestFilter_string(t *testing.T) {
	tests := []struct {
		Condition Condition
		Want      string
	}{
		{
			Condition: Condition{Type: "foo"},
			Want:      `If type is "foo"`,
		},
		{
			Condition: Condition{Type: "foo", Negate: true},
			Want:      `If type is not "foo"`,
		},
		{
			Condition: Condition{PayloadAction: "foo"},
			Want:      `If payload action is "foo"`,
		},
		{
			Condition: Condition{PayloadAction: "foo", Negate: true},
			Want:      `If payload action is not "foo"`,
		},
		{
			Condition: Condition{Type: "foo", PayloadAction: "bar"},
			Want:      `If type is "foo" AND payload action is "bar"`,
		},
		{
			Condition: Condition{Type: "foo", PayloadAction: "bar", Negate: true},
			Want:      `If type is not "foo" AND payload action is not "bar"`,
		},
		{
			Condition: Condition{PayloadIssueLabel: "foo"},
			Want:      `If payload issue label contains "foo"`,
		},
		{
			Condition: Condition{PayloadIssueLabel: "foo", Negate: true},
			Want:      `If payload issue label does not contain "foo"`,
		},
		{
			Condition: Condition{PayloadIssueMilestoneTitle: "foo"},
			Want:      `If payload issue milestone title is "foo"`,
		},
		{
			Condition: Condition{PayloadIssueMilestoneTitle: "foo", Negate: true},
			Want:      `If payload issue milestone title is not "foo"`,
		},
		{
			Condition: Condition{PayloadIssueTitleRegexp: `foo['"]`},
			Want:      `If payload issue title matches regexp "foo['\"]"`,
		},
		{
			Condition: Condition{PayloadIssueTitleRegexp: `foo['"]`, Negate: true},
			Want:      `If payload issue title does not match regexp "foo['\"]"`,
		},
		{
			Condition: Condition{PayloadIssueBodyRegexp: `foo['"]`},
			Want:      `If payload issue body matches regexp "foo['\"]"`,
		},
		{
			Condition: Condition{PayloadIssueBodyRegexp: `foo['"]`, Negate: true},
			Want:      `If payload issue body does not match regexp "foo['\"]"`,
		},
		{
			Condition: Condition{ComparePublic: true, Public: true},
			Want:      `If event is public`,
		},
		{
			Condition: Condition{ComparePublic: true, Public: false},
			Want:      `If event is not public`,
		},
		{
			Condition: Condition{ComparePublic: true, Public: true, Negate: true},
			Want:      `If event is not public`,
		},
		{
			Condition: Condition{ComparePublic: true, Public: false, Negate: true},
			Want:      `If event is not not public`, // :/
		},
		{
			Condition: Condition{OrganizationID: 1},
			Want:      `If organization ID is 1`,
		},
		{
			Condition: Condition{OrganizationID: 1, Negate: true},
			Want:      `If organization ID is not 1`,
		},
		{
			Condition: Condition{RepositoryID: 1},
			Want:      `If repository ID is 1`,
		},
		{
			Condition: Condition{RepositoryID: 1, Negate: true},
			Want:      `If repository ID is not 1`,
		},
	}

	for _, test := range tests {
		if have := test.Condition.String(); have != test.Want {
			t.Errorf("String does not match\nhave: %v\nwant: %v", have, test.Want)
		}
	}

}

func TestCondition_type(t *testing.T) {
	events := []*github.Event{
		{
			Type: github.String("IssuesEvent"),
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
			Condition: Condition{Type: "NonExistentEvent", Negate: true},
			Want:      events[0],
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
			Condition: Condition{PayloadAction: "CREATED"},
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

func TestCondition_payloadIssueLabel(t *testing.T) {
	var (
		empty    = json.RawMessage(`{"issue":{"labels":[]}}`)
		contains = json.RawMessage(`{"issue":{"labels":["LBL", "x"]}}`)
	)

	events := []*github.Event{
		{RawPayload: &empty},
		{RawPayload: &contains},
	}

	tests := []struct {
		Condition Condition
		Want      *github.Event
	}{
		{
			Condition: Condition{PayloadIssueLabel: "nomatch"},
			Want:      nil,
		},
		{
			Condition: Condition{PayloadIssueLabel: "lbl"},
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

func TestCondition_payloadIssueMilestoneTitle(t *testing.T) {
	var (
		empty    = json.RawMessage(`{"issue":{"milestone":null}}`)
		contains = json.RawMessage(`{"issue":{"milestone":{"Title":"title"}}}`)
	)

	events := []*github.Event{
		{RawPayload: &empty},
		{RawPayload: &contains},
	}

	tests := []struct {
		Condition Condition
		Want      *github.Event
	}{
		{
			Condition: Condition{PayloadIssueMilestoneTitle: "nomatch"},
			Want:      nil,
		},
		{
			Condition: Condition{PayloadIssueMilestoneTitle: "title"},
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

func TestCondition_payloadIssueTitleRegexp(t *testing.T) {
	var (
		match   = json.RawMessage(`{"issue":{"title":"This will Match"}}`)
		nomatch = json.RawMessage(`{"issue":{"title":"This will Not Match"}}`)
	)

	events := []*github.Event{
		{RawPayload: &match},
		{RawPayload: &nomatch},
	}

	tests := []struct {
		Condition Condition
		Want      *github.Event
	}{
		{
			Condition: Condition{PayloadIssueTitleRegexp: "not a match"},
			Want:      nil,
		},
		{
			Condition: Condition{PayloadIssueTitleRegexp: `(?i)will\s+match`},
			Want:      events[0],
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

func TestCondition_payloadIssueBodyRegexp(t *testing.T) {
	var (
		match   = json.RawMessage(`{"issue":{"body":"This will Match"}}`)
		nomatch = json.RawMessage(`{"issue":{"body":"This will Not Match"}}`)
	)

	events := []*github.Event{
		{RawPayload: &match},
		{RawPayload: &nomatch},
	}

	tests := []struct {
		Condition Condition
		Want      *github.Event
	}{
		{
			Condition: Condition{PayloadIssueBodyRegexp: "not a match"},
			Want:      nil,
		},
		{
			Condition: Condition{PayloadIssueBodyRegexp: `(?i)will\s+match`},
			Want:      events[0],
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
