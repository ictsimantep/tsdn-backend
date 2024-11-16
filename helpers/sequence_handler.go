package helpers

import (
	"database/sql"
	"fmt"
	"log"
)

// ResetSequenceToMax sets the specified sequence to start from the max value in the table + 1,
// but only increments the sequence if it matches the current max ID.
// ResetSequenceToMax adjusts the specified sequence to start from the maximum value in the table + 1.
func ResetSequenceToMax(db *sql.DB, tableName, columnName, sequenceName string) error {
	// Step 1: Get the current max value from the column
	var maxID int
	queryMax := fmt.Sprintf("SELECT COALESCE(MAX(%s), 0) FROM %s", columnName, tableName)
	err := db.QueryRow(queryMax).Scan(&maxID)
	if err != nil {
		log.Printf("Error getting max ID from %s: %v", tableName, err)
		return fmt.Errorf("failed to get max ID from %s: %w", tableName, err)
	}

	// Step 2: Calculate the next value for the sequence
	nextValue := maxID + 1

	// Step 3: Alter the sequence to start from the next value (no placeholders)
	queryAlter := fmt.Sprintf("ALTER SEQUENCE %s RESTART WITH %d", sequenceName, nextValue)
	_, err = db.Exec(queryAlter)
	if err != nil {
		log.Printf("Error altering sequence %s: %v", sequenceName, err)
		return fmt.Errorf("failed to alter sequence %s: %w", sequenceName, err)
	}

	log.Printf("Sequence %s successfully set to start from %d", sequenceName, nextValue)
	return nil
}
