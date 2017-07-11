package ghfilter

import (
	"encoding/json"
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
// TODO rename to EventCondition?
type Condition struct {
	// Type compares the Event's Type field. An empty Type will skip the check.
	Type string
	// PayloadAction compares the event's Action field in its payload. If not empty
	// the event must have a non-nil payload, must have an string action field. An
	// empty PayloadAction will skip the check. Comparison is case insensitive.
	PayloadAction string
	// PayloadIssueLabel compares the event's label issue labels array. If not empty
	// the payload must have a non-nil payload, issue and labels field. If empty the
	// fields are not checked. Comparison is case insensitive.
	PayloadIssueLabel string
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
		return false
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
			return false
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
			return false
		}
	}
	if c.ComparePublic && event.GetPublic() != c.Public {
		return false
	}
	if c.OrganizationID != 0 && (event.Org == nil || event.Org.GetID() != c.OrganizationID) {
		return false
	}
	if c.RepositoryID != 0 && (event.Repo == nil || event.Repo.GetID() != c.RepositoryID) {
		return false
	}
	return true
}
