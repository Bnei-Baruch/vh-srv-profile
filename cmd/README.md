# Import Relations Tool

A command-line tool for importing spouse relationships from CSV files into the VH profiles database.

## Overview

The `import-relations` command reads a CSV file containing relationship data, validates it against the database, and creates spouse relationships between users. It includes extensive validation, error handling, and reporting to ensure data quality.

## Usage

```bash
# Dry-run (default - validates without applying changes)
go run . import-relations --file relationships.csv

# Apply changes to database
go run . import-relations --file relationships.csv --dry-run=false
```

### Flags

- `--file, -f`: Path to CSV file containing relations (required)
- `--dry-run`: Perform validation without applying changes (default: `true`)

## CSV Format

The CSV file must contain the following columns:

| Column | Description |
|--------|-------------|
| `relationship_id` | Unique identifier for the relationship |
| `match_status` | Status of the relationship match |
| `contact_id_a` | Contact ID for Person A |
| `vh_user_id_a` | VH User ID (UUID) for Person A |
| `vh_first_name_a` | First name of Person A |
| `vh_last_name_a` | Last name of Person A |
| `vh_primary_email_a` | Email address of Person A |
| `contact_id_b` | Contact ID for Person B |
| `vh_user_id_b` | VH User ID (UUID) for Person B |
| `vh_first_name_b` | First name of Person B |
| `vh_last_name_b` | Last name of Person B |
| `vh_primary_email_b` | Email address of Person B |

**Note:** Unknown columns in the CSV will be ignored with a warning.

## Processing Pipeline

The import process follows these steps:

### 1. Read CSV
- Parses the CSV file
- Maps columns to internal data structures
- Tracks line numbers for error reporting
- Validates that PersonA and PersonB don't have the same UserID

### 2. Resolve IDs (Placeholder)
- **Current:** No-op placeholder
- **Future:** Will resolve missing UserIDs by:
  - Looking up users by email in the database
  - Matching by name and other attributes
  - Other resolution strategies

### 3. Collect Unique People
- Builds a map of unique people across all relations
- Detects data inconsistencies:
  - **Same ID Errors:** PersonA and PersonB in the same relation have identical UserIDs
  - **Data Mismatches:** Same UserID appears with different ContactID, Name, or Email across multiple relations
- Tracks CSV line numbers where each person appears
- Counts empty UserID records

### 4. Validate Against Database
- Checks if each user exists in the database
- Validates UUID format
- Retrieves keycloak IDs for spouse relationship creation

### 5. Apply Spouse Relationships
- Creates bidirectional spouse relationships in the database
- Handles idempotency (skips if relationship already exists)
- Only processes valid relations (both users found, no errors)

## Validation & Error Handling

### Error Categories

**1. Same ID Errors**
- PersonA and PersonB have identical UserIDs in the same relation
- These are **not** data mismatches - they represent invalid relationship data
- Counted separately in statistics

**2. Data Mismatch Errors**
- Same UserID appears with different personal data across multiple relations
- Example: UserID `abc-123` has different ContactIDs or names in different rows
- Indicates data quality issues in the source system

**3. Invalid UUIDs**
- UserID is not a valid UUID format
- Cannot be processed

**4. Users Not Found**
- UserID is valid but doesn't exist in the database
- Could indicate users that haven't been imported yet

**5. Empty UserIDs**
- PersonA or PersonB has no UserID
- Currently skipped; future versions may resolve via email lookup

### Skipping Strategy

Relations are skipped when:
- Either person has a "same ID" error
- Either person has inconsistent data across the CSV
- Either person has an invalid UUID
- Either person is not found in the database
- Keycloak ID cannot be retrieved for either person
- Relationship already exists (in non-dry-run mode)

## Output & Statistics

### CSV Statistics

```
INFO CSV Statistics:
INFO   total_relations count=395
INFO   total_person_records count=790
INFO   empty_user_id_person_a count=120
INFO   empty_user_id_person_b count=97
INFO   total_empty_user_ids count=217
INFO   unique_people count=573
INFO   people_with_errors count=6
INFO     - data_mismatch_errors count=3
INFO     - same_id_errors count=3
```

**Key Metrics:**
- `total_relations`: Number of relations in CSV
- `total_person_records`: Relations × 2 (theoretical maximum people)
- `empty_user_id_*`: Count of records without UserIDs
- `unique_people`: Actual unique UserIDs found (after deduplication)
- `people_with_errors`: Total people with validation errors
  - `data_mismatch_errors`: People with inconsistent data
  - `same_id_errors`: People flagged in "same ID" relations

### Detailed Error Logs

**Data Mismatch Example:**
```
WARN Inconsistent person details
  user_id=abc-123-def
  name="John Doe"
  email=john@example.com
  csv_lines=[15, 42, 87]
  error="Data mismatch - Existing: {ContactID: 1001, Name: John Doe, Email: john@example.com} vs New: {ContactID: 2002, Name: John Doe, Email: john@example.com}"
```

**Same ID Example:**
```
WARN Same ID person details
  user_id=xyz-789-abc
  name="Jane Smith"
  email=jane@example.com
  csv_lines=[157]
  error="Same user ID for both PersonA and PersonB: xyz-789-abc"
```

### Database Validation Statistics

```
INFO Database validation statistics
  people:
    found_in_db=545
    not_found_in_db=22
    inconsistent_data=3
    same_id_errors=3
    invalid_uuid=0
  relations:
    both_found=380
    one_found=12
    neither_found=3
    skipped_inconsistent=6
    skipped_invalid_uuid=0
  total_relations=395
  valid_for_import=380
```

### Application Statistics (Non-Dry-Run)

```
INFO Spouse relationship statistics
  successfully_applied=375
  person_a_inconsistent=3
  person_b_inconsistent=3
  invalid_uuid=0
  person_a_not_found=10
  person_b_not_found=12
  get_keycloak_id_a_failed=0
  get_keycloak_id_b_failed=0
  set_spouse_failed=2
  total_errors=30
  total_relations=395
```

## Safety Features

### Dry-Run Mode (Default)
- Validates all data without making database changes
- Shows exactly what would be applied
- Logs: `"DRY-RUN - Would set spouse"` for each valid relation
- Default behavior to prevent accidental data modification

### Idempotency
- Checks if spouse relationship already exists before creating
- Skips gracefully if relationship is already set to the correct spouse
- Prevents duplicate relationships

### Error Recovery
- Continues processing after individual errors
- Tracks all errors with detailed context
- Returns non-zero exit code if critical errors occurred

### Detailed Logging
- Line numbers in CSV for all errors
- Full context for data mismatches
- User IDs, names, and emails for debugging
- Progress indicators

## Future Enhancements

### Planned Features

1. **ID Resolution by Email**
   - Automatically resolve empty UserIDs by looking up users via email
   - Query database to find matching users
   - Update relations with resolved UserIDs before validation

2. **ID Resolution by Name Matching**
   - Fuzzy matching on first name + last name
   - Handle common name variations
   - Confidence scoring for matches

3. **Enhanced Validation**
   - Cross-reference with Keycloak to validate user accounts
   - Check for circular relationships
   - Validate email format and uniqueness

4. **Conflict Resolution**
   - Handle cases where user already has a different spouse
   - Provide options: skip, override, or prompt
   - Audit trail for relationship changes

5. **Batch Processing**
   - Process multiple CSV files in sequence
   - Rollback support for failed batches
   - Progress tracking for large imports

6. **Export & Reporting**
   - Export validation results to CSV
   - Generate HTML report with statistics
   - Email notifications for import completion

### Implementation Hooks

The codebase includes placeholder functions ready for implementation:

```go
// cmd/import_relations.go:235
func (ri *RelationImporter) resolveIDs(relations []Relation) []Relation {
    // Placeholder for future implementation
    // This function can be used to resolve missing user IDs by:
    // - Looking up users by email in the database
    // - Matching by name and other attributes
    // - Other resolution strategies

    slog.Info("ID resolution (not yet implemented)")
    return relations
}
```

To implement email-based resolution:
1. Query database: `SELECT user_id FROM users WHERE email = $1`
2. Update PersonA/PersonB UserIDs in relations
3. Track resolution statistics
4. Log resolved vs unresolved records

## Database Schema

The tool interacts with the following tables:

### users table
```sql
- user_id (UUID, primary key)
- keycloak_id (UUID)
- email (text)
- spouse_keycloak_id (UUID, nullable)
- deleted (boolean)
```

### Key Operations

**Check user exists:**
```sql
SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1 AND deleted = false)
```

**Get keycloak ID:**
```sql
SELECT keycloak_id FROM users WHERE user_id = $1 AND deleted = false
```

**Set spouse relationship:**
```go
repo.SetSpouse(ctx, keycloakIDA, keycloakIDB, false)
```
- Creates bidirectional relationship
- Updates `spouse_keycloak_id` for both users
- Returns `ErrSpouseConflict` if user already has a different spouse

## Error Handling

### Exit Codes
- `0`: Success (dry-run or all relationships applied)
- `1`: Critical error (file not found, database connection failed, etc.)

### Common Issues

**"PersonA and PersonB cannot be the same user"**
- Cause: Same UserID used for both people in a relation
- Solution: Fix source data to use different UserIDs

**"Data mismatch - Existing: {...} vs New: {...}"**
- Cause: Same UserID has different ContactID/Name/Email in different rows
- Solution: Reconcile data in source system before import

**"User NOT found in DB"**
- Cause: UserID doesn't exist in the database
- Solution: Import users first, or implement email-based resolution

**"keycloak_id is NULL or empty"**
- Cause: User exists but has no Keycloak ID
- Solution: Run user sync with Keycloak, or fix data in users table

## Examples

### Dry-Run Validation
```bash
go run . import-relations -f relationships.csv

# Output:
# INFO CSV file read successfully total_relations=395
# INFO ID resolution (not yet implemented)
# INFO CSV Statistics: ...
# INFO Database validation completed
# INFO DRY-RUN - Would set spouse ... (for each valid relation)
# INFO Spouse relationship statistics (dry-run) would_apply=375 ...
```

### Actual Import
```bash
go run . import-relations -f relationships.csv --dry-run=false

# Output:
# INFO CSV file read successfully total_relations=395
# INFO ID resolution (not yet implemented)
# INFO CSV Statistics: ...
# INFO Database validation completed
# INFO SUCCESS - Set spouse ... (for each applied relation)
# INFO Spouse relationship statistics successfully_applied=375 ...
```

## Development

### Running Tests
```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific test
go test -run TestImportRelations ./...
```

### Code Structure
```
cmd/
├── import_relations.go       # Main implementation
├── README.md                 # This file
└── root.go                   # CLI root command setup

Key types:
- Person: Represents a person with contact info and validation state
- Relation: Represents a relationship between two people
- RelationImporter: Handles the import process
```

## Contributing

When modifying the import tool:
1. Preserve backward compatibility with CSV format
2. Add tests for new validation rules
3. Update statistics output to reflect new metrics
4. Document new error conditions in this README
5. Ensure dry-run mode always remains the default

## Support

For issues or questions:
- Check logs for detailed error messages with line numbers
- Run in dry-run mode first to validate data
- Ensure database connection is working
- Verify CSV format matches expected columns
