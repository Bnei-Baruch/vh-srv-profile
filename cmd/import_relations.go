package cmd

import (
	"context"
	"encoding/csv"
	"errors"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v4"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cobra"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/repo"
)

func init() {
	rootCmd.AddCommand(importRelationsCmd)
	importRelationsCmd.Flags().StringP("file", "f", "", "Path to CSV file containing relations")
	importRelationsCmd.MarkFlagRequired("file")
	importRelationsCmd.Flags().Bool("dry-run", true, "Perform validation without actually applying spouse relationships (default: true for safety)")
}

var importRelationsCmd = &cobra.Command{
	Use:   "import-relations",
	Short: "Import relations from CSV file",
	Long:  "Reads a CSV file containing user relations and validates them against the database",
	Run:   importRelationsFn,
}

// CSV column name mapping - update this when CSV structure changes
var csvFieldMapping = map[string]string{
	"relationship_id":    "RelationshipID",
	"match_status":       "MatchStatus",
	"contact_id_a":       "ContactIDA",
	"vh_user_id_a":       "VHUserIDA",
	"vh_first_name_a":    "VHFirstNameA",
	"vh_last_name_a":     "VHLastNameA",
	"vh_primary_email_a": "VHPrimaryEmailA",
	"contact_id_b":       "ContactIDB",
	"vh_user_id_b":       "VHUserIDB",
	"vh_first_name_b":    "VHFirstNameB",
	"vh_last_name_b":     "VHLastNameB",
	"vh_primary_email_b": "VHPrimaryEmailB",
}

type Person struct {
	ContactID string
	UserID    string
	FirstName string
	LastName  string
	Email     string
	FoundInDB bool
	Error     string // Set when person has inconsistent data in CSV
	Lines     []int  // CSV line numbers where this person appears
}

func (p Person) Equals(other Person) bool {
	return p.ContactID == other.ContactID &&
		p.UserID == other.UserID &&
		p.FirstName == other.FirstName &&
		p.LastName == other.LastName &&
		p.Email == other.Email
}

type Relation struct {
	RelationshipID string
	MatchStatus    string
	PersonA        Person
	PersonB        Person
	LineNumber     int // Line number in CSV (1-based, accounting for header)
}

func importRelationsFn(cmd *cobra.Command, args []string) {
	filePath, _ := cmd.Flags().GetString("file")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	slog.Info("starting import-relations", slog.String("file", filePath))
	if dryRun {
		slog.Info("DRY-RUN MODE: No changes will be applied to the database")
	}

	importer := NewRelationImporter()
	if err := importer.init(); err != nil {
		slog.Error("failed to initialize", slog.Any("err", err))
		os.Exit(1)
	}
	defer importer.close()

	if err := importer.run(filePath, dryRun); err != nil {
		slog.Error("failed to run import", slog.Any("err", err))
		os.Exit(1)
	}

	slog.Info("import-relations completed successfully")
}

type RelationImporter struct {
	repo *repo.ProfileDB
}

func NewRelationImporter() *RelationImporter {
	return &RelationImporter{}
}

func (ri *RelationImporter) init() error {
	dbUrl := repo.MakeDBURL()
	db, err := repo.NewProfileDB(context.TODO(), dbUrl, nil)
	if err != nil {
		return errors.Join(errors.New("repo.NewProfileDB "+dbUrl), err)
	}
	ri.repo = db
	return nil
}

func (ri *RelationImporter) close() {
	ri.repo.Close()
}

func (ri *RelationImporter) run(filePath string, dryRun bool) error {
	relations, sameIDUserIDs, err := ri.readCSV(filePath)
	if err != nil {
		return errors.Join(errors.New("readCSV"), err)
	}

	// Attempt to resolve missing user IDs (placeholder for future implementation)
	relations = ri.resolveIDs(relations)

	people := ri.collectUniquePeople(relations, sameIDUserIDs)

	ri.printCSVStatistics(relations, people)

	if err := ri.validateUsers(relations, people); err != nil {
		return errors.Join(errors.New("validateUsers"), err)
	}

	ri.printFinalStatistics(relations, people)

	if err := ri.applySpouseRelationships(relations, people, dryRun); err != nil {
		return errors.Join(errors.New("applySpouseRelationships"), err)
	}

	return nil
}

func (ri *RelationImporter) readCSV(filePath string) ([]Relation, map[string]bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, errors.Join(errors.New("os.Open"), err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	header, err := reader.Read()
	if err != nil {
		return nil, nil, errors.Join(errors.New("read header"), err)
	}

	// Check for unknown columns
	unknownColumns := []string{}
	for _, col := range header {
		if _, exists := csvFieldMapping[col]; !exists {
			unknownColumns = append(unknownColumns, col)
		}
	}

	if len(unknownColumns) > 0 {
		slog.Warn("Found unknown columns in CSV (will be ignored)", slog.Int("count", len(unknownColumns)), slog.Any("columns", unknownColumns))
	}

	// Build column index map
	colIndex := make(map[string]int)
	for i, col := range header {
		colIndex[col] = i
	}

	var relations []Relation
	sameIDUserIDs := make(map[string]bool) // Track UserIDs where PersonA == PersonB in same relation
	lineNum := 1

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, errors.Join(errors.New("read line "+strconv.Itoa(lineNum)), err)
		}

		getField := func(fieldName string) string {
			if idx, ok := colIndex[fieldName]; ok && idx < len(record) {
				return record[idx]
			}
			return ""
		}

		relation := Relation{
			RelationshipID: getField("relationship_id"),
			MatchStatus:    getField("match_status"),
			LineNumber:     lineNum + 1, // +1 to account for header
			PersonA: Person{
				ContactID: getField("contact_id_a"),
				UserID:    getField("vh_user_id_a"),
				FirstName: getField("vh_first_name_a"),
				LastName:  getField("vh_last_name_a"),
				Email:     getField("vh_primary_email_a"),
			},
			PersonB: Person{
				ContactID: getField("contact_id_b"),
				UserID:    getField("vh_user_id_b"),
				FirstName: getField("vh_first_name_b"),
				LastName:  getField("vh_last_name_b"),
				Email:     getField("vh_primary_email_b"),
			},
		}

		// Validate that PersonA and PersonB are not the same user
		if relation.PersonA.UserID != "" && relation.PersonA.UserID == relation.PersonB.UserID {
			slog.Warn("PersonA and PersonB have the same UserID - will skip this relation",
				slog.Int("line", relation.LineNumber),
				slog.String("user_id", relation.PersonA.UserID))
			errorMsg := "Same user ID for both PersonA and PersonB: " + relation.PersonA.UserID
			relation.PersonA.Error = errorMsg
			relation.PersonB.Error = errorMsg
			relation.PersonA.Lines = []int{relation.LineNumber}
			relation.PersonB.Lines = []int{relation.LineNumber}
			sameIDUserIDs[relation.PersonA.UserID] = true
		}

		relations = append(relations, relation)
		lineNum++
	}

	slog.Info("CSV file read successfully", slog.Int("total_relations", len(relations)))
	return relations, sameIDUserIDs, nil
}

func (ri *RelationImporter) resolveIDs(relations []Relation) []Relation {
	ctx := context.TODO()

	resolvedPersonA := 0
	resolvedPersonB := 0
	duplicateWarningsA := 0
	duplicateWarningsB := 0
	notFoundA := 0
	notFoundB := 0
	failedPersonA := 0
	failedPersonB := 0

	// First pass: Log all records with empty user IDs
	emptyUserIDPersonA := 0
	emptyUserIDPersonB := 0
	emptyUserIDNoEmailA := 0
	emptyUserIDNoEmailB := 0

	slog.Info("Scanning CSV for records with empty user IDs")
	for _, rel := range relations {
		if rel.PersonA.UserID == "" {
			emptyUserIDPersonA++
			if rel.PersonA.Email == "" {
				emptyUserIDNoEmailA++
				slog.Warn("PersonA missing both UserID and Email - cannot resolve",
					slog.Int("line", rel.LineNumber),
					slog.String("name", rel.PersonA.FirstName+" "+rel.PersonA.LastName),
					slog.String("contact_id", rel.PersonA.ContactID))
			} else {
				slog.Debug("PersonA missing UserID - will attempt email resolution",
					slog.Int("line", rel.LineNumber),
					slog.String("email", rel.PersonA.Email),
					slog.String("name", rel.PersonA.FirstName+" "+rel.PersonA.LastName))
			}
		}

		if rel.PersonB.UserID == "" {
			emptyUserIDPersonB++
			if rel.PersonB.Email == "" {
				emptyUserIDNoEmailB++
				slog.Warn("PersonB missing both UserID and Email - cannot resolve",
					slog.Int("line", rel.LineNumber),
					slog.String("name", rel.PersonB.FirstName+" "+rel.PersonB.LastName),
					slog.String("contact_id", rel.PersonB.ContactID))
			} else {
				slog.Debug("PersonB missing UserID - will attempt email resolution",
					slog.Int("line", rel.LineNumber),
					slog.String("email", rel.PersonB.Email),
					slog.String("name", rel.PersonB.FirstName+" "+rel.PersonB.LastName))
			}
		}
	}

	slog.Info("Empty UserID scan completed",
		slog.Int("empty_userid_person_a", emptyUserIDPersonA),
		slog.Int("empty_userid_person_b", emptyUserIDPersonB),
		slog.Int("with_email_person_a", emptyUserIDPersonA-emptyUserIDNoEmailA),
		slog.Int("with_email_person_b", emptyUserIDPersonB-emptyUserIDNoEmailB),
		slog.Int("no_email_person_a", emptyUserIDNoEmailA),
		slog.Int("no_email_person_b", emptyUserIDNoEmailB),
		slog.Int("total_resolvable", (emptyUserIDPersonA-emptyUserIDNoEmailA)+(emptyUserIDPersonB-emptyUserIDNoEmailB)))

	slog.Info("Starting ID resolution by email lookup (profiles database only)")

	for i := range relations {
		// Resolve PersonA if UserID is empty but Email exists
		if relations[i].PersonA.UserID == "" && relations[i].PersonA.Email != "" {
			userID, hasDuplicates, err := ri.lookupUserIDByEmail(ctx, relations[i].PersonA.Email)
			if err != nil {
				slog.Warn("Failed to lookup PersonA by email",
					slog.Int("line", relations[i].LineNumber),
					slog.String("email", relations[i].PersonA.Email),
					slog.Any("err", err))
				failedPersonA++
			} else if userID != "" {
				relations[i].PersonA.UserID = userID
				resolvedPersonA++

				if hasDuplicates {
					duplicateWarningsA++
					slog.Warn("PersonA email matches multiple users - using oldest account",
						slog.Int("line", relations[i].LineNumber),
						slog.String("email", relations[i].PersonA.Email),
						slog.String("user_id", userID))
				} else {
					slog.Debug("Resolved PersonA UserID by email",
						slog.Int("line", relations[i].LineNumber),
						slog.String("email", relations[i].PersonA.Email),
						slog.String("user_id", userID))
				}
			} else {
				notFoundA++
				slog.Debug("PersonA email not found in database",
					slog.Int("line", relations[i].LineNumber),
					slog.String("email", relations[i].PersonA.Email))
			}
		}

		// Resolve PersonB if UserID is empty but Email exists
		if relations[i].PersonB.UserID == "" && relations[i].PersonB.Email != "" {
			userID, hasDuplicates, err := ri.lookupUserIDByEmail(ctx, relations[i].PersonB.Email)
			if err != nil {
				slog.Warn("Failed to lookup PersonB by email",
					slog.Int("line", relations[i].LineNumber),
					slog.String("email", relations[i].PersonB.Email),
					slog.Any("err", err))
				failedPersonB++
			} else if userID != "" {
				relations[i].PersonB.UserID = userID
				resolvedPersonB++

				if hasDuplicates {
					duplicateWarningsB++
					slog.Warn("PersonB email matches multiple users - using oldest account",
						slog.Int("line", relations[i].LineNumber),
						slog.String("email", relations[i].PersonB.Email),
						slog.String("user_id", userID))
				} else {
					slog.Debug("Resolved PersonB UserID by email",
						slog.Int("line", relations[i].LineNumber),
						slog.String("email", relations[i].PersonB.Email),
						slog.String("user_id", userID))
				}
			} else {
				notFoundB++
				slog.Debug("PersonB email not found in database",
					slog.Int("line", relations[i].LineNumber),
					slog.String("email", relations[i].PersonB.Email))
			}
		}
	}

	slog.Info("ID resolution completed",
		slog.Int("resolved_person_a", resolvedPersonA),
		slog.Int("resolved_person_b", resolvedPersonB),
		slog.Int("not_found_person_a", notFoundA),
		slog.Int("not_found_person_b", notFoundB),
		slog.Int("duplicate_warnings_a", duplicateWarningsA),
		slog.Int("duplicate_warnings_b", duplicateWarningsB),
		slog.Int("failed_person_a", failedPersonA),
		slog.Int("failed_person_b", failedPersonB),
		slog.Int("total_resolved", resolvedPersonA+resolvedPersonB))

	return relations
}

func (ri *RelationImporter) collectUniquePeople(relations []Relation, sameIDUserIDs map[string]bool) map[string]Person {
	people := make(map[string]Person)
	inconsistentCount := 0
	emptyUserIDCountA := 0
	emptyUserIDCountB := 0

	for _, rel := range relations {
		// Process PersonA
		if rel.PersonA.UserID != "" {
			// Skip data mismatch checking if this UserID has "same ID for two people" error
			if sameIDUserIDs[rel.PersonA.UserID] {
				// Just add/update with the error already set from readCSV
				if existing, exists := people[rel.PersonA.UserID]; !exists {
					people[rel.PersonA.UserID] = rel.PersonA
				} else {
					// Accumulate line numbers
					existing.Lines = append(existing.Lines, rel.LineNumber)
					people[rel.PersonA.UserID] = existing
				}
			} else {
				if existing, exists := people[rel.PersonA.UserID]; exists {
					// Accumulate line numbers
					existing.Lines = append(existing.Lines, rel.LineNumber)
					if !existing.Equals(rel.PersonA) {
						if existing.Error == "" {
							// First time we detected inconsistency for this user
							inconsistentCount++
						}
						errorMsg := "Data mismatch - Existing: {ContactID: " + existing.ContactID + ", Name: " + existing.FirstName + " " + existing.LastName + ", Email: " + existing.Email + "} vs New: {ContactID: " + rel.PersonA.ContactID + ", Name: " + rel.PersonA.FirstName + " " + rel.PersonA.LastName + ", Email: " + rel.PersonA.Email + "}"
						existing.Error = errorMsg
						slog.Warn("Inconsistent data detected", slog.String("user_id", rel.PersonA.UserID), slog.Int("line", rel.LineNumber))
					}
					people[rel.PersonA.UserID] = existing
				} else {
					rel.PersonA.Lines = []int{rel.LineNumber}
					people[rel.PersonA.UserID] = rel.PersonA
				}
			}
		} else {
			emptyUserIDCountA++
		}

		// Process PersonB
		if rel.PersonB.UserID != "" {
			// Skip data mismatch checking if this UserID has "same ID for two people" error
			if sameIDUserIDs[rel.PersonB.UserID] {
				// Just add/update with the error already set from readCSV
				if existing, exists := people[rel.PersonB.UserID]; !exists {
					people[rel.PersonB.UserID] = rel.PersonB
				} else {
					// Accumulate line numbers
					existing.Lines = append(existing.Lines, rel.LineNumber)
					people[rel.PersonB.UserID] = existing
				}
			} else {
				if existing, exists := people[rel.PersonB.UserID]; exists {
					// Accumulate line numbers
					existing.Lines = append(existing.Lines, rel.LineNumber)
					if !existing.Equals(rel.PersonB) {
						if existing.Error == "" {
							// First time we detected inconsistency for this user
							inconsistentCount++
						}
						errorMsg := "Data mismatch - Existing: {ContactID: " + existing.ContactID + ", Name: " + existing.FirstName + " " + existing.LastName + ", Email: " + existing.Email + "} vs New: {ContactID: " + rel.PersonB.ContactID + ", Name: " + rel.PersonB.FirstName + " " + rel.PersonB.LastName + ", Email: " + rel.PersonB.Email + "}"
						existing.Error = errorMsg
						slog.Warn("Inconsistent data detected", slog.String("user_id", rel.PersonB.UserID), slog.Int("line", rel.LineNumber))
					}
					people[rel.PersonB.UserID] = existing
				} else {
					rel.PersonB.Lines = []int{rel.LineNumber}
					people[rel.PersonB.UserID] = rel.PersonB
				}
			}
		} else {
			emptyUserIDCountB++
		}
	}

	// Count same ID people separately
	sameIDCount := 0
	for _, person := range people {
		if person.Error != "" && strings.HasPrefix(person.Error, "Same user ID for both PersonA and PersonB: ") {
			sameIDCount++
		}
	}

	// Log inconsistent data people (data mismatch errors)
	slog.Info("Found people with inconsistent data across relations (will be skipped)", slog.Int("count", inconsistentCount))
	for userID, person := range people {
		if person.Error != "" && !strings.HasPrefix(person.Error, "Same user ID for both PersonA and PersonB: ") {
			slog.Warn("Inconsistent person details",
				slog.String("user_id", userID),
				slog.String("name", person.FirstName+" "+person.LastName),
				slog.String("email", person.Email),
				slog.Any("csv_lines", person.Lines),
				slog.String("error", person.Error))
		}
	}

	// Log same ID people separately
	slog.Info("Found people with same ID for both PersonA and PersonB (will be skipped)", slog.Int("count", sameIDCount))
	for userID, person := range people {
		if person.Error != "" && strings.HasPrefix(person.Error, "Same user ID for both PersonA and PersonB: ") {
			slog.Warn("Same ID person details",
				slog.String("user_id", userID),
				slog.String("name", person.FirstName+" "+person.LastName),
				slog.String("email", person.Email),
				slog.Any("csv_lines", person.Lines),
				slog.String("error", person.Error))
		}
	}

	return people
}

func (ri *RelationImporter) printCSVStatistics(relations []Relation, people map[string]Person) {
	inconsistentPeople := 0
	sameIDPeople := 0
	emptyUserIDCountA := 0
	emptyUserIDCountB := 0
	totalPersonRecords := len(relations) * 2

	// Count empty UserIDs
	for _, rel := range relations {
		if rel.PersonA.UserID == "" {
			emptyUserIDCountA++
		}
		if rel.PersonB.UserID == "" {
			emptyUserIDCountB++
		}
	}

	// Count error types
	for _, person := range people {
		if person.Error != "" {
			// Check if this is a "same ID" error
			if strings.HasPrefix(person.Error, "Same user ID for both PersonA and PersonB: ") {
				sameIDPeople++
			} else {
				inconsistentPeople++
			}
		}
	}

	slog.Info("CSV Statistics:")
	slog.Info("  total_relations", slog.Int("count", len(relations)))
	slog.Info("  total_person_records", slog.Int("count", totalPersonRecords))
	slog.Info("  empty_user_id_person_a", slog.Int("count", emptyUserIDCountA))
	slog.Info("  empty_user_id_person_b", slog.Int("count", emptyUserIDCountB))
	slog.Info("  total_empty_user_ids", slog.Int("count", emptyUserIDCountA+emptyUserIDCountB))
	slog.Info("  unique_people", slog.Int("count", len(people)))
	slog.Info("  people_with_errors", slog.Int("count", inconsistentPeople+sameIDPeople))
	slog.Info("    - data_mismatch_errors", slog.Int("count", inconsistentPeople))
	slog.Info("    - same_id_errors", slog.Int("count", sameIDPeople))
}

func (ri *RelationImporter) validateUsers(relations []Relation, people map[string]Person) error {
	ctx := context.TODO()

	slog.Info("Starting database validation")

	// Check each unique person in the database
	for userID, person := range people {
		// Skip people with inconsistent data
		if person.Error != "" {
			continue
		}

		// Try to parse UUID
		userUUID, err := uuid.FromString(userID)
		if err != nil {
			slog.Warn("Invalid UUID format", slog.String("user_id", userID))
			person.FoundInDB = false
			people[userID] = person
			continue
		}

		// Check if user exists in database by querying
		found, err := ri.userExistsInDB(ctx, userUUID)
		if err != nil {
			return errors.Join(errors.New("failed to check user "+userID+" in DB"), err)
		}

		person.FoundInDB = found
		people[userID] = person

		if found {
			slog.Debug("User found in DB", slog.String("user_id", userID))
		} else {
			slog.Debug("User NOT found in DB", slog.String("user_id", userID))
		}
	}

	slog.Info("Database validation completed")
	return nil
}

func (ri *RelationImporter) userExistsInDB(ctx context.Context, userID uuid.UUID) (bool, error) {
	var exists bool
	err := ri.repo.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1 AND deleted = false)`, userID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// lookupUserIDByEmail attempts to find a user_id by searching for the email
// in primary_email, alternate_email_1, and alternate_email_2 fields.
// Returns: (user_id, hasDuplicates, error)
func (ri *RelationImporter) lookupUserIDByEmail(ctx context.Context, email string) (string, bool, error) {
	if email == "" {
		return "", false, errors.New("email is empty")
	}

	var userID string

	// Case-insensitive email lookup across all email fields
	// Note: Multiple users may have the same email (41 duplicates exist in production)
	// We take the first match (oldest account) and return a flag indicating duplicates
	query := `
		SELECT user_id::text
		FROM users
		WHERE deleted = false
		AND (
			LOWER(primary_email) = LOWER($1)
			OR LOWER(alternate_email_1) = LOWER($1)
			OR LOWER(alternate_email_2) = LOWER($1)
		)
		ORDER BY created_at ASC
		LIMIT 1
	`

	err := ri.repo.QueryRow(ctx, query, email).Scan(&userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", false, nil // Not found, but not an error
		}
		return "", false, err
	}

	// Check if there are duplicate emails (for warning purposes)
	var duplicateCount int
	countQuery := `
		SELECT COUNT(*)
		FROM users
		WHERE deleted = false
		AND (
			LOWER(primary_email) = LOWER($1)
			OR LOWER(alternate_email_1) = LOWER($1)
			OR LOWER(alternate_email_2) = LOWER($1)
		)
	`
	_ = ri.repo.QueryRow(ctx, countQuery, email).Scan(&duplicateCount)
	hasDuplicates := duplicateCount > 1

	return userID, hasDuplicates, nil
}

func (ri *RelationImporter) getSpouseKeycloakID(ctx context.Context, userID uuid.UUID) (*string, error) {
	var spouseKeycloakID *string
	err := ri.repo.QueryRow(ctx, `SELECT spouse_keycloak_id FROM users WHERE user_id = $1 AND deleted = false`, userID).Scan(&spouseKeycloakID)
	if err != nil {
		return nil, err
	}
	return spouseKeycloakID, nil
}

func (ri *RelationImporter) getKeycloakID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	var keycloakID *string
	err := ri.repo.QueryRow(ctx, `SELECT keycloak_id FROM users WHERE user_id = $1 AND deleted = false`, userID).Scan(&keycloakID)
	if err != nil {
		return uuid.Nil, errors.Join(errors.New("failed to query keycloak_id"), err)
	}

	// Validate keycloak_id is not NULL or empty
	if keycloakID == nil || *keycloakID == "" {
		return uuid.Nil, errors.New("keycloak_id is NULL or empty for user_id " + userID.String())
	}

	// Parse and validate UUID format
	parsedUUID, err := uuid.FromString(*keycloakID)
	if err != nil {
		return uuid.Nil, errors.Join(errors.New("invalid keycloak_id UUID format for user_id "+userID.String()), err)
	}

	return parsedUUID, nil
}

func (ri *RelationImporter) applySpouseRelationships(relations []Relation, people map[string]Person, dryRun bool) error {
	ctx := context.TODO()

	mode := "APPLYING"
	if dryRun {
		mode = "DRY-RUN"
	}

	slog.Info("Starting to apply spouse relationships", slog.String("mode", mode))

	applied := 0
	errorPersonAInconsistent := 0
	errorPersonBInconsistent := 0
	errorInvalidUUID := 0
	errorPersonANotFound := 0
	errorPersonBNotFound := 0
	errorGetKeycloakIDA := 0
	errorGetKeycloakIDB := 0
	errorSetSpouse := 0

	for i, rel := range relations {
		lineNum := i + 2 // +2 because of CSV header and 0-based index

		// Check for inconsistent data
		personAHasError := rel.PersonA.UserID != "" && people[rel.PersonA.UserID].Error != ""
		personBHasError := rel.PersonB.UserID != "" && people[rel.PersonB.UserID].Error != ""

		if personAHasError {
			slog.Warn("SKIP - Person A has inconsistent data",
				slog.Int("line", lineNum),
				slog.String("user_id", rel.PersonA.UserID),
				slog.String("name", rel.PersonA.FirstName+" "+rel.PersonA.LastName))
			errorPersonAInconsistent++
			continue
		}

		if personBHasError {
			slog.Warn("SKIP - Person B has inconsistent data",
				slog.Int("line", lineNum),
				slog.String("user_id", rel.PersonB.UserID),
				slog.String("name", rel.PersonB.FirstName+" "+rel.PersonB.LastName))
			errorPersonBInconsistent++
			continue
		}

		// Check for invalid UUIDs
		userIDA, errA := uuid.FromString(rel.PersonA.UserID)
		userIDB, errB := uuid.FromString(rel.PersonB.UserID)
		if errA != nil || errB != nil {
			slog.Warn("SKIP - Invalid UUID format",
				slog.Int("line", lineNum),
				slog.String("user_id_a", rel.PersonA.UserID),
				slog.String("user_id_b", rel.PersonB.UserID))
			errorInvalidUUID++
			continue
		}

		// Check if found in DB
		personAFound := people[rel.PersonA.UserID].FoundInDB
		personBFound := people[rel.PersonB.UserID].FoundInDB

		if !personAFound {
			slog.Warn("SKIP - Person A not found in DB",
				slog.Int("line", lineNum),
				slog.String("user_id", rel.PersonA.UserID),
				slog.String("name", rel.PersonA.FirstName+" "+rel.PersonA.LastName),
				slog.String("email", rel.PersonA.Email))
			errorPersonANotFound++
			continue
		}

		if !personBFound {
			slog.Warn("SKIP - Person B not found in DB",
				slog.Int("line", lineNum),
				slog.String("user_id", rel.PersonB.UserID),
				slog.String("name", rel.PersonB.FirstName+" "+rel.PersonB.LastName),
				slog.String("email", rel.PersonB.Email))
			errorPersonBNotFound++
			continue
		}

		// Get keycloak IDs for both users
		keycloakIDA, err := ri.getKeycloakID(ctx, userIDA)
		if err != nil {
			slog.Error("Failed to get keycloak_id for user A",
				slog.Int("line", lineNum),
				slog.String("user_id", rel.PersonA.UserID),
				slog.Any("err", err))
			errorGetKeycloakIDA++
			continue
		}

		keycloakIDB, err := ri.getKeycloakID(ctx, userIDB)
		if err != nil {
			slog.Error("Failed to get keycloak_id for user B",
				slog.Int("line", lineNum),
				slog.String("user_id", rel.PersonB.UserID),
				slog.Any("err", err))
			errorGetKeycloakIDB++
			continue
		}

		// Apply spouse relationship (or skip if dry-run)
		if dryRun {
			slog.Info("DRY-RUN - Would set spouse",
				slog.Int("line", lineNum),
				slog.String("person_a", rel.PersonA.FirstName+" "+rel.PersonA.LastName+" ("+rel.PersonA.UserID+")"),
				slog.String("person_b", rel.PersonB.FirstName+" "+rel.PersonB.LastName+" ("+rel.PersonB.UserID+")"))
			applied++
		} else {
			err = ri.repo.SetSpouse(ctx, keycloakIDA, keycloakIDB, false)
			if err != nil {
				// Handle idempotency: if relationship already exists with same spouse, skip gracefully
				if errors.Is(err, common.ErrSpouseConflict) {
					existingSpouseA, checkErr := ri.getSpouseKeycloakID(ctx, userIDA)
					if checkErr == nil && existingSpouseA != nil && *existingSpouseA == keycloakIDB.String() {
						// Relationship already exists - skip gracefully
						slog.Info("SKIP - Spouse relationship already exists",
							slog.Int("line", lineNum),
							slog.String("person_a", rel.PersonA.FirstName+" "+rel.PersonA.LastName+" ("+rel.PersonA.UserID+")"),
							slog.String("person_b", rel.PersonB.FirstName+" "+rel.PersonB.LastName+" ("+rel.PersonB.UserID+")"))
						applied++
						continue
					}
					// Conflict with different spouse - real error, fall through
				}

				slog.Error("Failed to set spouse relationship",
					slog.Int("line", lineNum),
					slog.String("relationship_id", rel.RelationshipID),
					slog.String("user_a", rel.PersonA.UserID),
					slog.String("user_b", rel.PersonB.UserID),
					slog.Any("err", err))
				errorSetSpouse++
				continue
			}

			slog.Info("SUCCESS - Set spouse",
				slog.Int("line", lineNum),
				slog.String("person_a", rel.PersonA.FirstName+" "+rel.PersonA.LastName+" ("+rel.PersonA.UserID+")"),
				slog.String("person_b", rel.PersonB.FirstName+" "+rel.PersonB.LastName+" ("+rel.PersonB.UserID+")"))
			applied++
		}

		slog.Debug("Spouse relationship processed",
			slog.String("user_a", rel.PersonA.UserID),
			slog.String("user_b", rel.PersonB.UserID),
			slog.Bool("dry_run", dryRun))
	}

	// Log final statistics
	totalErrors := errorPersonAInconsistent + errorPersonBInconsistent + errorInvalidUUID +
		errorPersonANotFound + errorPersonBNotFound + errorGetKeycloakIDA + errorGetKeycloakIDB + errorSetSpouse

	if dryRun {
		slog.Info("Spouse relationship statistics (dry-run)",
			slog.Int("would_apply", applied),
			slog.Int("person_a_inconsistent", errorPersonAInconsistent),
			slog.Int("person_b_inconsistent", errorPersonBInconsistent),
			slog.Int("invalid_uuid", errorInvalidUUID),
			slog.Int("person_a_not_found", errorPersonANotFound),
			slog.Int("person_b_not_found", errorPersonBNotFound),
			slog.Int("get_keycloak_id_a_failed", errorGetKeycloakIDA),
			slog.Int("get_keycloak_id_b_failed", errorGetKeycloakIDB),
			slog.Int("total_skipped", totalErrors),
			slog.Int("total_relations", len(relations)))
	} else {
		slog.Info("Spouse relationship statistics",
			slog.Int("successfully_applied", applied),
			slog.Int("person_a_inconsistent", errorPersonAInconsistent),
			slog.Int("person_b_inconsistent", errorPersonBInconsistent),
			slog.Int("invalid_uuid", errorInvalidUUID),
			slog.Int("person_a_not_found", errorPersonANotFound),
			slog.Int("person_b_not_found", errorPersonBNotFound),
			slog.Int("get_keycloak_id_a_failed", errorGetKeycloakIDA),
			slog.Int("get_keycloak_id_b_failed", errorGetKeycloakIDB),
			slog.Int("set_spouse_failed", errorSetSpouse),
			slog.Int("total_errors", totalErrors),
			slog.Int("total_relations", len(relations)))
	}

	slog.Info("Spouse relationship application completed",
		slog.Int("applied", applied),
		slog.Int("total_errors", totalErrors),
		slog.Bool("dry_run", dryRun))

	if !dryRun && (errorGetKeycloakIDA > 0 || errorGetKeycloakIDB > 0 || errorSetSpouse > 0) {
		return errors.New("encountered errors while applying spouse relationships")
	}

	return nil
}

func (ri *RelationImporter) printFinalStatistics(relations []Relation, people map[string]Person) {
	// Count different categories
	bothFound := 0
	oneFound := 0
	noneFound := 0
	skipped := 0
	invalidUUID := 0

	peopleFound := 0
	peopleNotFound := 0
	peopleInconsistent := 0
	peopleSameID := 0
	peopleInvalidUUID := 0

	// Count people statistics
	for userID, person := range people {
		if person.Error != "" {
			// Check if this is a "same ID" error
			if strings.HasPrefix(person.Error, "Same user ID for both PersonA and PersonB: ") {
				peopleSameID++
			} else {
				peopleInconsistent++
			}
			continue
		}

		_, err := uuid.FromString(userID)
		if err != nil {
			peopleInvalidUUID++
			continue
		}

		if person.FoundInDB {
			peopleFound++
		} else {
			peopleNotFound++
		}
	}

	// Count relation statistics
	for _, rel := range relations {
		personAHasError := rel.PersonA.UserID != "" && people[rel.PersonA.UserID].Error != ""
		personBHasError := rel.PersonB.UserID != "" && people[rel.PersonB.UserID].Error != ""

		// Skip if either person has inconsistent data
		if personAHasError || personBHasError {
			skipped++
			continue
		}

		// Check for invalid UUIDs
		_, errA := uuid.FromString(rel.PersonA.UserID)
		_, errB := uuid.FromString(rel.PersonB.UserID)
		if (rel.PersonA.UserID != "" && errA != nil) || (rel.PersonB.UserID != "" && errB != nil) {
			invalidUUID++
			continue
		}

		// Check if found in DB
		personAFound := rel.PersonA.UserID != "" && people[rel.PersonA.UserID].FoundInDB
		personBFound := rel.PersonB.UserID != "" && people[rel.PersonB.UserID].FoundInDB

		if personAFound && personBFound {
			bothFound++
		} else if personAFound || personBFound {
			oneFound++
		} else {
			noneFound++
		}
	}

	slog.Info("Database validation statistics",
		slog.Group("people",
			slog.Int("found_in_db", peopleFound),
			slog.Int("not_found_in_db", peopleNotFound),
			slog.Int("inconsistent_data", peopleInconsistent),
			slog.Int("same_id_errors", peopleSameID),
			slog.Int("invalid_uuid", peopleInvalidUUID)),
		slog.Group("relations",
			slog.Int("both_found", bothFound),
			slog.Int("one_found", oneFound),
			slog.Int("neither_found", noneFound),
			slog.Int("skipped_inconsistent", skipped),
			slog.Int("skipped_invalid_uuid", invalidUUID)),
		slog.Int("total_relations", len(relations)),
		slog.Int("valid_for_import", bothFound))
}
