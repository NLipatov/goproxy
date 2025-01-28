package repositories

import (
	"database/sql"
	"fmt"
)

func checkRowsAffected(result sql.Result) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not check rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("no rows affected")
	}
	return nil
}
