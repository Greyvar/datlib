// Package lint loads and validates Greyvar dat/ trees.
//
// Callers such as datlint and the game server share the same Check/Issue
// results so load-time diagnostics stay consistent.
package lint

// Severity classifies a lint Issue.
type Severity int

const (
	SeverityError Severity = iota
	SeverityWarning
)

// Issue is one problem found while loading or validating a dat file.
type Issue struct {
	Issue    string
	Severity Severity
}

// Check is the result of loading/validating one dat file (or scan target).
type Check struct {
	Type     string
	Filename string
	Issues   []Issue
}

// AddError appends an error-severity issue.
func (c *Check) AddError(issue string) {
	c.Issues = append(c.Issues, Issue{Issue: issue, Severity: SeverityError})
}

// AddWarning appends a warning-severity issue.
func (c *Check) AddWarning(issue string) {
	c.Issues = append(c.Issues, Issue{Issue: issue, Severity: SeverityWarning})
}

// HasErrors reports whether any issue is an error.
func (c *Check) HasErrors() bool {
	for _, issue := range c.Issues {
		if issue.Severity == SeverityError {
			return true
		}
	}
	return false
}

// Count tallies errors and warnings across checks.
func Count(checks []Check) (errors, warnings int) {
	for _, check := range checks {
		for _, issue := range check.Issues {
			switch issue.Severity {
			case SeverityWarning:
				warnings++
			default:
				errors++
			}
		}
	}
	return errors, warnings
}
