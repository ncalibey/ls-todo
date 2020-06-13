package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"ls-todo/internal/config"
	"ls-todo/internal/models"
)

// PGManager is used for interacting with the PostgreSQL database.
type PGManager interface {
	// GetTodos retrieves all todos from the database.
	GetTodos() ([]*models.Todo, error)
}

// pgManager implements the PGManager interface for "production".
type pgManager struct {
	// db is the database connection.
	db *sqlx.DB
}

// New returns a new PGManager instance.
func New(db *sqlx.DB) PGManager {
 return &pgManager{db}
}

func (m *pgManager) GetTodos() ([]*models.Todo, error) {
	// We open a database transaction.
	tx, err := m.db.Beginx()
	if err != nil {
		return nil, err
	}
	// If the transaction fails, we wan't to rollback any changes. We use the `defer`
	// keyword to ensure this is called if we encounter an error at any point. If the
	// transaction is successfully committed, then this won't do anything (which is what
	// we want in that case).
	defer tx.Rollback()

	// Next, we query for all todos in the database.
	rows, err := tx.Queryx("SELECT * FROM todos")
	if err != nil {
		return nil, err
	}
	// We need to close the rows once we're done using them. We use `defer` so this happens
	// "automatically".
	defer rows.Close()

	// We create a slice of todos that we will store our results in.
	var todos []*models.Todo
	// We iterate over all the returned rows.
	for rows.Next() {
		// We create a todo struct that we'll scan the results into.
		var todo models.Todo
		// sqlx provides a StructScan method that will scan the contents of a row into a struct
		// that models the returned data. This is _very_ useful and is much easier than scanning
		// each of the individual data points and them assigning them to the fields of a struct.
		//
		// Notice that we need to pass a pointer along since we want to scan to the specific point
		// in memory.
		if err := rows.StructScan(&todo); err != nil {
			return nil, err
		}
		// This is essentially the same as `todos.push(todo)` in JS or Ruby. Again, we need to pass
		// a pointer since the type of the slice is `[]*models.Todo`.
		todos = append(todos, &todo)
	}

	// Next, we check to see if there were any errors in processing the rows.
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// Lastly, we commit the transaction.
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	// We return the slice of todos and a `nil` for the error (since no errors were found).
	return todos, nil
}


//////////////////////////////////////////////////////////////////////////////////////////
//// Helpers /////////////////////////////////////////////////////////////////////////////

// GetConnString returns the connection string for connecting to a PostgreSQL database.
func GetConnString(cfg *config.Config) string {
	return fmt.Sprintf(
		"host=%s user=%s dbname=%s port=%d sslmode=%v password=%s",
		cfg.PGHost,
		cfg.PGUser,
		cfg.PGDatabase,
		cfg.PGPort,
		cfg.PGSSLMode,
		cfg.PGPassword,
	)
}
