package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

// ReportEntry holds a successfully applied or already-linked relation.
type ReportEntry struct {
	RelationshipID string
	MatchStatus    string
	PersonA        Person
	PersonB        Person
}

// SkippedEntry holds a relation skipped due to a data issue.
type SkippedEntry struct {
	RelationshipID string
	MatchStatus    string
	Reason         string
	Detail         string
	PersonA        Person
	PersonB        Person
}

// FailedEntry holds a relation that failed during DB operations.
type FailedEntry struct {
	RelationshipID string
	MatchStatus    string
	Reason         string
	Err            string
	Detail         string
	PersonA        Person
	PersonB        Person
}

// ImportReport collects results during an import run and renders them as markdown.
type ImportReport struct {
	Date          time.Time
	CSVPath       string
	DryRun        bool
	Applied       []ReportEntry
	AlreadyLinked []ReportEntry
	Failed        []FailedEntry
	Skipped       []SkippedEntry
}

func NewImportReport(csvPath string, dryRun bool) *ImportReport {
	return &ImportReport{
		Date:    time.Now(),
		CSVPath: csvPath,
		DryRun:  dryRun,
	}
}

func (r *ImportReport) addApplied(rel Relation) {
	r.Applied = append(r.Applied, ReportEntry{
		RelationshipID: rel.RelationshipID,
		MatchStatus:    rel.MatchStatus,
		PersonA:        rel.PersonA,
		PersonB:        rel.PersonB,
	})
}

func (r *ImportReport) addAlreadyLinked(rel Relation) {
	r.AlreadyLinked = append(r.AlreadyLinked, ReportEntry{
		RelationshipID: rel.RelationshipID,
		MatchStatus:    rel.MatchStatus,
		PersonA:        rel.PersonA,
		PersonB:        rel.PersonB,
	})
}

func (r *ImportReport) addFailed(rel Relation, reason string, err error, detail string) {
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	r.Failed = append(r.Failed, FailedEntry{
		RelationshipID: rel.RelationshipID,
		MatchStatus:    rel.MatchStatus,
		Reason:         reason,
		Err:            errStr,
		Detail:         detail,
		PersonA:        rel.PersonA,
		PersonB:        rel.PersonB,
	})
}

func (r *ImportReport) addSkipped(rel Relation, reason string, detail string) {
	r.Skipped = append(r.Skipped, SkippedEntry{
		RelationshipID: rel.RelationshipID,
		MatchStatus:    rel.MatchStatus,
		Reason:         reason,
		Detail:         detail,
		PersonA:        rel.PersonA,
		PersonB:        rel.PersonB,
	})
}

func (r *ImportReport) Write(path string) error {
	var sb strings.Builder

	mode := "LIVE"
	if r.DryRun {
		mode = "DRY-RUN"
	}

	total := len(r.Applied) + len(r.AlreadyLinked) + len(r.Failed) + len(r.Skipped)

	sb.WriteString("# Import Relations Report\n\n")
	sb.WriteString(fmt.Sprintf("- **Date:** %s\n", r.Date.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("- **CSV:** `%s`\n", r.CSVPath))
	sb.WriteString(fmt.Sprintf("- **Mode:** %s\n\n", mode))

	sb.WriteString("## Summary\n\n")
	sb.WriteString("| | Count |\n|---|---|\n")
	sb.WriteString(fmt.Sprintf("| Total relations | %d |\n", total))
	sb.WriteString(fmt.Sprintf("| Successfully applied | %d |\n", len(r.Applied)))
	sb.WriteString(fmt.Sprintf("| Already linked | %d |\n", len(r.AlreadyLinked)))
	sb.WriteString(fmt.Sprintf("| Failed | %d |\n", len(r.Failed)))
	sb.WriteString(fmt.Sprintf("| Skipped (data issues) | %d |\n\n", len(r.Skipped)))

	sb.WriteString("---\n\n")

	// Applied
	sb.WriteString(fmt.Sprintf("## ✅ Successfully Applied (%d)\n\n", len(r.Applied)))
	if len(r.Applied) == 0 {
		sb.WriteString("_None_\n\n")
	} else {
		sb.WriteString("| # | Rel ID | Match Status | Person A | Person B |\n|---|--------|--------------|----------|----------|\n")
		for i, e := range r.Applied {
			sb.WriteString(fmt.Sprintf("| %d | %s | %s | %s (%s) | %s (%s) |\n",
				i+1, e.RelationshipID, e.MatchStatus,
				reportPersonName(e.PersonA), e.PersonA.Email,
				reportPersonName(e.PersonB), e.PersonB.Email,
			))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("---\n\n")

	// Already linked
	sb.WriteString(fmt.Sprintf("## 🔗 Already Linked (%d)\n\n", len(r.AlreadyLinked)))
	if len(r.AlreadyLinked) == 0 {
		sb.WriteString("_None_\n\n")
	} else {
		sb.WriteString("| # | Rel ID | Match Status | Person A | Person B |\n|---|--------|--------------|----------|----------|\n")
		for i, e := range r.AlreadyLinked {
			sb.WriteString(fmt.Sprintf("| %d | %s | %s | %s (%s) | %s (%s) |\n",
				i+1, e.RelationshipID, e.MatchStatus,
				reportPersonName(e.PersonA), e.PersonA.Email,
				reportPersonName(e.PersonB), e.PersonB.Email,
			))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("---\n\n")

	// Failed — grouped by reason
	sb.WriteString(fmt.Sprintf("## ❌ Failed (%d)\n\n", len(r.Failed)))
	if len(r.Failed) == 0 {
		sb.WriteString("_None_\n\n")
	} else {
		failedByReason := make(map[string][]FailedEntry)
		for _, e := range r.Failed {
			failedByReason[e.Reason] = append(failedByReason[e.Reason], e)
		}
		reasons := sortedKeys(failedByReason)
		for _, reason := range reasons {
			entries := failedByReason[reason]
			sb.WriteString(fmt.Sprintf("### %s (%d)\n\n", reason, len(entries)))
			for _, e := range entries {
				sb.WriteString(fmt.Sprintf("#### Relation %s\n\n", e.RelationshipID))
				writeRelationDetails(&sb, e.MatchStatus, e.PersonA, e.PersonB)
				if e.Err != "" {
					sb.WriteString(fmt.Sprintf("- **Error:** `%s`\n", e.Err))
				}
				if e.Detail != "" {
					sb.WriteString(fmt.Sprintf("- **Detail:** %s\n", e.Detail))
				}
				sb.WriteString("\n")
			}
		}
	}

	sb.WriteString("---\n\n")

	// Skipped — grouped by reason
	sb.WriteString(fmt.Sprintf("## ⚠️ Skipped — Data Issues (%d)\n\n", len(r.Skipped)))
	if len(r.Skipped) == 0 {
		sb.WriteString("_None_\n\n")
	} else {
		skippedByReason := make(map[string][]SkippedEntry)
		for _, e := range r.Skipped {
			skippedByReason[e.Reason] = append(skippedByReason[e.Reason], e)
		}
		reasons := sortedKeys(skippedByReason)
		for _, reason := range reasons {
			entries := skippedByReason[reason]
			sb.WriteString(fmt.Sprintf("### %s (%d)\n\n", reason, len(entries)))
			for _, e := range entries {
				sb.WriteString(fmt.Sprintf("#### Relation %s\n\n", e.RelationshipID))
				writeRelationDetails(&sb, e.MatchStatus, e.PersonA, e.PersonB)
				if e.Detail != "" {
					sb.WriteString(fmt.Sprintf("- **Detail:** %s\n", e.Detail))
				}
				sb.WriteString("\n")
			}
		}
	}

	return os.WriteFile(path, []byte(sb.String()), 0644)
}

// writeRelationDetails writes a full input detail table for a relation entry.
func writeRelationDetails(sb *strings.Builder, matchStatus string, a Person, b Person) {
	sb.WriteString(fmt.Sprintf("- **Match Status:** %s\n", matchStatus))
	sb.WriteString("\n")
	sb.WriteString("  | Field | Person A | Person B |\n  |---|---|---|\n")
	sb.WriteString(fmt.Sprintf("  | Contact ID | %s | %s |\n", orEmpty(a.ContactID), orEmpty(b.ContactID)))
	sb.WriteString(fmt.Sprintf("  | VH User ID | `%s` | `%s` |\n", orEmpty(a.UserID), orEmpty(b.UserID)))
	sb.WriteString(fmt.Sprintf("  | Name | %s | %s |\n", reportPersonName(a), reportPersonName(b)))
	sb.WriteString(fmt.Sprintf("  | Email | %s | %s |\n", orEmpty(a.Email), orEmpty(b.Email)))
	sb.WriteString("\n")
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func reportPersonName(p Person) string {
	name := strings.TrimSpace(p.FirstName + " " + p.LastName)
	if name == "" {
		return "(empty)"
	}
	return name
}

func orEmpty(s string) string {
	if s == "" {
		return "_(empty)_"
	}
	return s
}
