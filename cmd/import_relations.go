package cmd

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"

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
	importRelationsCmd.Flags().String("report", "", "Path to write markdown report file (optional)")
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
}

func importRelationsFn(cmd *cobra.Command, args []string) {
	filePath, _ := cmd.Flags().GetString("file")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	reportPath, _ := cmd.Flags().GetString("report")

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

	if err := importer.run(filePath, dryRun, reportPath); err != nil {
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

func (ri *RelationImporter) run(filePath string, dryRun bool, reportPath string) error {
	var report *ImportReport
	if reportPath != "" {
		report = NewImportReport(filePath, dryRun)
	}

	relations, err := ri.readCSV(filePath)
	if err != nil {
		return errors.Join(errors.New("readCSV"), err)
	}

	people := ri.collectUniquePeople(relations)

	ri.printCSVStatistics(relations, people)

	if err := ri.validateUsers(relations, people); err != nil {
		return errors.Join(errors.New("validateUsers"), err)
	}

	ri.printFinalStatistics(relations, people)

	if err := ri.applySpouseRelationships(relations, people, dryRun, report); err != nil {
		if report != nil {
			if werr := report.Write(reportPath); werr != nil {
				slog.Error("failed to write report", slog.Any("err", werr))
			} else {
				slog.Info("report written", slog.String("path", reportPath))
			}
		}
		return errors.Join(errors.New("applySpouseRelationships"), err)
	}

	if report != nil {
		if werr := report.Write(reportPath); werr != nil {
			slog.Error("failed to write report", slog.Any("err", werr))
		} else {
			slog.Info("report written", slog.String("path", reportPath))
		}
	}

	return nil
}

func (ri *RelationImporter) readCSV(filePath string) ([]Relation, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, errors.Join(errors.New("os.Open"), err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	header, err := reader.Read()
	if err != nil {
		return nil, errors.Join(errors.New("read header"), err)
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
	lineNum := 1

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, errors.Join(errors.New("read line "+strconv.Itoa(lineNum)), err)
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
			return nil, errors.New("line " + strconv.Itoa(lineNum+1) + ": PersonA and PersonB cannot be the same user (UserID: " + relation.PersonA.UserID + ")")
		}

		relations = append(relations, relation)
		lineNum++
	}

	slog.Info("CSV file read successfully", slog.Int("total_relations", len(relations)))
	return relations, nil
}

func (ri *RelationImporter) collectUniquePeople(relations []Relation) map[string]Person {
	people := make(map[string]Person)
	inconsistentCount := 0

	for _, rel := range relations {
		// Process PersonA
		if rel.PersonA.UserID != "" {
			if existing, exists := people[rel.PersonA.UserID]; exists {
				if !existing.Equals(rel.PersonA) {
					if existing.Error == "" {
						// First time we detected inconsistency for this user
						inconsistentCount++
					}
					errorMsg := "Data mismatch - Existing: {ContactID: " + existing.ContactID + ", Name: " + existing.FirstName + " " + existing.LastName + ", Email: " + existing.Email + "} vs New: {ContactID: " + rel.PersonA.ContactID + ", Name: " + rel.PersonA.FirstName + " " + rel.PersonA.LastName + ", Email: " + rel.PersonA.Email + "}"
					existing.Error = errorMsg
					people[rel.PersonA.UserID] = existing
					slog.Warn("Inconsistent data detected", slog.String("user_id", rel.PersonA.UserID))
				}
			} else {
				people[rel.PersonA.UserID] = rel.PersonA
			}
		}

		// Process PersonB
		if rel.PersonB.UserID != "" {
			if existing, exists := people[rel.PersonB.UserID]; exists {
				if !existing.Equals(rel.PersonB) {
					if existing.Error == "" {
						// First time we detected inconsistency for this user
						inconsistentCount++
					}
					errorMsg := "Data mismatch - Existing: {ContactID: " + existing.ContactID + ", Name: " + existing.FirstName + " " + existing.LastName + ", Email: " + existing.Email + "} vs New: {ContactID: " + rel.PersonB.ContactID + ", Name: " + rel.PersonB.FirstName + " " + rel.PersonB.LastName + ", Email: " + rel.PersonB.Email + "}"
					existing.Error = errorMsg
					people[rel.PersonB.UserID] = existing
					slog.Warn("Inconsistent data detected", slog.String("user_id", rel.PersonB.UserID))
				}
			} else {
				people[rel.PersonB.UserID] = rel.PersonB
			}
		}
	}

	if inconsistentCount > 0 {
		slog.Warn("Found people with inconsistent data across relations (will be skipped)", slog.Int("count", inconsistentCount))
		for userID, person := range people {
			if person.Error != "" {
				slog.Warn("Inconsistent person details",
					slog.String("user_id", userID),
					slog.String("name", person.FirstName+" "+person.LastName),
					slog.String("email", person.Email),
					slog.String("error", person.Error))
			}
		}
	}

	return people
}

func (ri *RelationImporter) printCSVStatistics(relations []Relation, people map[string]Person) {
	inconsistentPeople := 0
	for _, person := range people {
		if person.Error != "" {
			inconsistentPeople++
		}
	}

	slog.Info("CSV Statistics",
		slog.Int("total_relations", len(relations)),
		slog.Int("unique_people", len(people)),
		slog.Int("inconsistent_people", inconsistentPeople))
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

func (ri *RelationImporter) getSpouseKeycloakID(ctx context.Context, userID uuid.UUID) (*string, error) {
	var spouseKeycloakID *string
	err := ri.repo.QueryRow(ctx, `SELECT spouse_keycloak_id FROM users WHERE user_id = $1 AND deleted = false`, userID).Scan(&spouseKeycloakID)
	if err != nil {
		return nil, err
	}
	return spouseKeycloakID, nil
}

func (ri *RelationImporter) getExistingSpouseDetail(ctx context.Context, userID uuid.UUID, intendedSpouseKeycloakID string) string {
	spouseKeycloakID, err := ri.getSpouseKeycloakID(ctx, userID)
	if err != nil {
		return "(lookup error: " + err.Error() + ")"
	}
	if spouseKeycloakID == nil {
		return "(none)"
	}
	if *spouseKeycloakID == intendedSpouseKeycloakID {
		return "(same as intended)"
	}
	var firstName, lastName, email, spouseUserID string
	lookupErr := ri.repo.QueryRow(ctx,
		`SELECT COALESCE(first_name_latin, ''), COALESCE(last_name_latin, ''), COALESCE(primary_email, ''), user_id::text FROM users WHERE keycloak_id::text = $1 AND deleted = false`,
		*spouseKeycloakID).Scan(&firstName, &lastName, &email, &spouseUserID)
	if lookupErr != nil {
		return fmt.Sprintf("keycloak_id=%s (info lookup failed)", *spouseKeycloakID)
	}
	name := strings.TrimSpace(firstName + " " + lastName)
	if name == "" {
		name = "(empty)"
	}
	return fmt.Sprintf("%s <%s> (user_id=%s)", name, email, spouseUserID)
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

func (ri *RelationImporter) applySpouseRelationships(relations []Relation, people map[string]Person, dryRun bool, report *ImportReport) error {
	ctx := context.TODO()

	mode := "APPLYING"
	if dryRun {
		mode = "DRY-RUN"
	}

	slog.Info("Starting to apply spouse relationships", slog.String("mode", mode))

	applied := 0
	alreadyLinked := 0
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
			if report != nil {
				report.addSkipped(rel, "Person A has inconsistent data", people[rel.PersonA.UserID].Error)
			}
			continue
		}

		if personBHasError {
			slog.Warn("SKIP - Person B has inconsistent data",
				slog.Int("line", lineNum),
				slog.String("user_id", rel.PersonB.UserID),
				slog.String("name", rel.PersonB.FirstName+" "+rel.PersonB.LastName))
			errorPersonBInconsistent++
			if report != nil {
				report.addSkipped(rel, "Person B has inconsistent data", people[rel.PersonB.UserID].Error)
			}
			continue
		}

		// Check for empty UserIDs (no VH match in CSV — not a UUID error)
		if rel.PersonA.UserID == "" {
			slog.Warn("SKIP - Person A has no VH match",
				slog.Int("line", lineNum),
				slog.String("relationship_id", rel.RelationshipID))
			errorPersonANotFound++
			if report != nil {
				report.addSkipped(rel, "Person A not found in DB", "no VH match in CSV (empty user_id)")
			}
			continue
		}
		if rel.PersonB.UserID == "" {
			slog.Warn("SKIP - Person B has no VH match",
				slog.Int("line", lineNum),
				slog.String("relationship_id", rel.RelationshipID))
			errorPersonBNotFound++
			if report != nil {
				report.addSkipped(rel, "Person B not found in DB", "no VH match in CSV (empty user_id)")
			}
			continue
		}

		// Check for invalid UUIDs (non-empty but malformed)
		userIDA, errA := uuid.FromString(rel.PersonA.UserID)
		userIDB, errB := uuid.FromString(rel.PersonB.UserID)
		if errA != nil || errB != nil {
			slog.Warn("SKIP - Invalid UUID format",
				slog.Int("line", lineNum),
				slog.String("user_id_a", rel.PersonA.UserID),
				slog.String("user_id_b", rel.PersonB.UserID))
			errorInvalidUUID++
			if report != nil {
				report.addSkipped(rel, "Invalid UUID", "user_id_a: "+rel.PersonA.UserID+", user_id_b: "+rel.PersonB.UserID)
			}
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
			if report != nil {
				report.addSkipped(rel, "Person A not found in DB", "user_id: "+rel.PersonA.UserID)
			}
			continue
		}

		if !personBFound {
			slog.Warn("SKIP - Person B not found in DB",
				slog.Int("line", lineNum),
				slog.String("user_id", rel.PersonB.UserID),
				slog.String("name", rel.PersonB.FirstName+" "+rel.PersonB.LastName),
				slog.String("email", rel.PersonB.Email))
			errorPersonBNotFound++
			if report != nil {
				report.addSkipped(rel, "Person B not found in DB", "user_id: "+rel.PersonB.UserID)
			}
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
			if report != nil {
				report.addFailed(rel, "Failed to get keycloak_id for Person A", err, "")
			}
			continue
		}

		keycloakIDB, err := ri.getKeycloakID(ctx, userIDB)
		if err != nil {
			slog.Error("Failed to get keycloak_id for user B",
				slog.Int("line", lineNum),
				slog.String("user_id", rel.PersonB.UserID),
				slog.Any("err", err))
			errorGetKeycloakIDB++
			if report != nil {
				report.addFailed(rel, "Failed to get keycloak_id for Person B", err, "")
			}
			continue
		}

		// Apply spouse relationship (or skip if dry-run)
		if dryRun {
			slog.Info("DRY-RUN - Would set spouse",
				slog.Int("line", lineNum),
				slog.String("person_a", rel.PersonA.FirstName+" "+rel.PersonA.LastName+" ("+rel.PersonA.UserID+")"),
				slog.String("person_b", rel.PersonB.FirstName+" "+rel.PersonB.LastName+" ("+rel.PersonB.UserID+")"))
			applied++
			if report != nil {
				report.addApplied(rel)
			}
		} else {
			// Pre-check: if A already points to B, the link is correct — skip without re-writing
			existingSpouseA, checkErr := ri.getSpouseKeycloakID(ctx, userIDA)
			if checkErr == nil && existingSpouseA != nil && *existingSpouseA == keycloakIDB.String() {
				slog.Info("SKIP - Spouse relationship already exists",
					slog.Int("line", lineNum),
					slog.String("person_a", rel.PersonA.FirstName+" "+rel.PersonA.LastName+" ("+rel.PersonA.UserID+")"),
					slog.String("person_b", rel.PersonB.FirstName+" "+rel.PersonB.LastName+" ("+rel.PersonB.UserID+")"))
				alreadyLinked++
				if report != nil {
					report.addAlreadyLinked(rel)
				}
				continue
			}

			err = ri.repo.SetSpouse(ctx, keycloakIDA, keycloakIDB, false)
			if err != nil {
				if errors.Is(err, common.ErrSpouseConflict) {
					// Conflict with different spouse - look up who they are linked to
					slog.Error("Spouse conflict: already linked to different person",
						slog.Int("line", lineNum),
						slog.String("relationship_id", rel.RelationshipID),
						slog.String("user_a", rel.PersonA.UserID),
						slog.String("user_b", rel.PersonB.UserID))
					errorSetSpouse++
					if report != nil {
						spouseADetail := ri.getExistingSpouseDetail(ctx, userIDA, keycloakIDB.String())
						spouseBDetail := ri.getExistingSpouseDetail(ctx, userIDB, keycloakIDA.String())
						detail := "Existing spouse of Person A: " + spouseADetail + "; Existing spouse of Person B: " + spouseBDetail
						report.addFailed(rel, "Spouse conflict: already linked to different person", err, detail)
					}
					continue
				}

				slog.Error("Failed to set spouse relationship",
					slog.Int("line", lineNum),
					slog.String("relationship_id", rel.RelationshipID),
					slog.String("user_a", rel.PersonA.UserID),
					slog.String("user_b", rel.PersonB.UserID),
					slog.Any("err", err))
				errorSetSpouse++
				if report != nil {
					report.addFailed(rel, "Failed to set spouse relationship", err, "")
				}
				continue
			
			}

			slog.Info("SUCCESS - Set spouse",
				slog.Int("line", lineNum),
				slog.String("person_a", rel.PersonA.FirstName+" "+rel.PersonA.LastName+" ("+rel.PersonA.UserID+")"),
				slog.String("person_b", rel.PersonB.FirstName+" "+rel.PersonB.LastName+" ("+rel.PersonB.UserID+")"))
			applied++
			if report != nil {
				report.addApplied(rel)
			}
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
			slog.Int("already_linked", alreadyLinked),
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
		slog.Int("already_linked", alreadyLinked),
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
	peopleInvalidUUID := 0

	// Count people statistics
	for userID, person := range people {
		if person.Error != "" {
			peopleInconsistent++
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
