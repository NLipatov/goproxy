package repositories

import (
	"database/sql"
	"fmt"
)

type SqlResult struct {
	result sql.Result
}

func NewSqlResult(result sql.Result) *SqlResult {
	return &SqlResult{
		result: result,
	}
}

func (s *SqlResult) checkRowsAffected() error {
	affected, err := s.result.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not check rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("no rows affected")
	}
	return nil
}
