package ghfilter

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/google/go-github/github"
)

// Filter is a collection of conditions.
type Filter struct {
	Conditions []Condition
}

// Matches returns true if event matches all conditions, else return false.
func (f *Filter) Matches(event *github.Event) bool {
	for _, condition := range f.Conditions {
		if !condition.Matches(event) {
			return false
		}
	}
	return true
}

// A Condition is a test which compares multiple fields with a GitHub event's.
type Condition struct {
	// Negate causes a positive match to return false, instead of true.
	//
	// If a condition requires fields to be set, they will continue to return false.
	// For example, if Nagate is true and PayloadAction a non zero value, if the
	// event does not have a payload with an action key, the match will continue
	// to be false.
	Negate bool
	// Type compares the Event's Type field. An empty Type will skip the check.
	Type string
	// PayloadAction compares the event's Action field in its payload. If not empty
	// the event must have a non-nil payload, must have an string action field. An
	// empty PayloadAction will skip the check. Comparison is case insensitive.
	PayloadAction string
	// PayloadIssueLabel compares the event's issue labels array. If not empty
	// the payload must have a non-nil payload, issue and labels field. If empty the
	// fields are not checked. Comparison is case insensitive.
	PayloadIssueLabel string
	// PayloadIssueMilestoneTitle compares the event's issue milestone's title. If not
	// empty the payload must have a non-nil payload, issue and milestone field. If
	// empty the fields are not checked. Comparison is case insensitive.
	PayloadIssueMilestoneTitle string
	// PayloadIssueTitleRegexp compares the event's issue title against regexp. If not
	// empty the payload must have a non-nil payload, issue and title field. If
	// empty the fields are not checked. See https://golang.org/pkg/regexp for syntax.
	PayloadIssueTitleRegexp string
	// PayloadIssueBodyRegexp compares the event's issue body against regexp. If not
	// empty the payload must have a non-nil payload, issue and body field. If
	// empty the fields are not checked. See https://golang.org/pkg/regexp for syntax.
	PayloadIssueBodyRegexp string
	// ComparePublic enables comparing of the event's public field with the condition's
	// Public value. Setting to false will skip checking the Public field.
	ComparePublic bool
	// Public compares the event's Public field. ComparePublic must be set to true to
	// compare the Public field.
	Public bool
	// OrganizationID compares the event's Organizaton's ID field. The event must have
	// a non-nil Organization. A zero value will skip the check.
	OrganizationID int
	// RepositoryID compares the event's Repository's ID field. The event must have
	// a non-nil Repository. A zero value will skip the check.
	RepositoryID int
}

// Matches returns false if any test fails. In other words, it returns true if all
// tests pass or no tests are set.
// TODO rename to Test?
func (c *Condition) Matches(event *github.Event) bool {
	if c.Type != "" && event.GetType() != c.Type {
		return c.Negate
	}
	if c.PayloadAction != "" {
		if event.RawPayload == nil {
			return false
		}
		var payload struct {
			Action string `json:"action"`
		}
		if err := json.Unmarshal(*event.RawPayload, &payload); err != nil {
			// TODO return, log, ignore? could just be the payload doesn't have an action?
			return false
		}
		if strings.ToLower(payload.Action) != strings.ToLower(c.PayloadAction) {
			return c.Negate
		}
	}
	if c.PayloadIssueLabel != "" {
		if event.RawPayload == nil {
			return false
		}
		var payload struct {
			Issue struct {
				Labels []string `json:"labels"`
			} `json:"issue"`
		}
		if err := json.Unmarshal(*event.RawPayload, &payload); err != nil {
			// May not have issue.labels
			return false
		}
		found := false
		for _, label := range payload.Issue.Labels {
			if strings.ToLower(label) == strings.ToLower(c.PayloadIssueLabel) {
				found = true
			}
		}
		if !found {
			return c.Negate
		}
	}
	if c.PayloadIssueMilestoneTitle != "" {
		if event.RawPayload == nil {
			return false
		}
		var payload struct {
			Issue struct {
				Milestone struct {
					Title string `json:"title"`
				} `json:"milestone"`
			} `json:"issue"`
		}
		if err := json.Unmarshal(*event.RawPayload, &payload); err != nil {
			// May not have issue.milestone.title
			return false
		}
		if strings.ToLower(payload.Issue.Milestone.Title) != strings.ToLower(c.PayloadIssueMilestoneTitle) {
			return c.Negate
		}
	}
	if c.PayloadIssueTitleRegexp != "" {
		if event.RawPayload == nil {
			return false
		}
		var payload struct {
			Issue struct {
				Title string `json:"title"`
			} `json:"issue"`
		}
		if err := json.Unmarshal(*event.RawPayload, &payload); err != nil {
			// May not have issue.title
			return false
		}
		re, err := regexp.Compile(c.PayloadIssueTitleRegexp)
		if err != nil {
			return false
		}
		if !re.MatchString(payload.Issue.Title) {
			return c.Negate
		}
	}
	if c.PayloadIssueBodyRegexp != "" {
		if event.RawPayload == nil {
			return false
		}
		var payload struct {
			Issue struct {
				Body string `json:"body"`
			} `json:"issue"`
		}
		if err := json.Unmarshal(*event.RawPayload, &payload); err != nil {
			// May not have issue.title
			return false
		}
		re, err := regexp.Compile(c.PayloadIssueBodyRegexp)
		if err != nil {
			return false
		}
		if !re.MatchString(payload.Issue.Body) {
			return c.Negate
		}
	}
	if c.ComparePublic && event.GetPublic() != c.Public {
		return c.Negate
	}
	if c.OrganizationID != 0 && (event.Org == nil || event.Org.GetID() != c.OrganizationID) {
		return c.Negate
	}
	if c.RepositoryID != 0 && (event.Repo == nil || event.Repo.GetID() != c.RepositoryID) {
		return c.Negate
	}
	return !c.Negate
}
