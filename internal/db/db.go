package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"ls-todo/internal/config"
	"ls-todo/internal/models"
)

// PGManager is used for interacting with the PostgreSQL database.
type PGManager interface {
	// GetTodos retrieves all todos.
	GetTodos() ([]*models.Todo, error)
	// GetTodo retrieves a single todo.
	GetTodo(id int64) (*models.Todo, error)
	// CreateTodo creates a new todo.
	CreateTodo(todo *models.Todo) (*models.Todo, error)
	// UpdateTodo update a given todo.
	UpdateTodo(diff *models.Todo, id int64) (*models.Todo, error)
	// DeleteTodo deletes a given todo.
	DeleteTodo(id int64) (*models.Todo, error)
	// ToggleTodo toggles the completed state of a given todo.
	ToggleTodo(id int64) (*models.Todo, error)
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
	rows, err := tx.Queryx("SELECT * FROM todos ORDER BY id")
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

func (m *pgManager) GetTodo(id int64) (*models.Todo, error) {
	tx, err := m.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var todo models.Todo
	// Here we use `QueryRowx` which can be used when we know there will only be one result.
	// We then chain the StructScan call.
	if err := tx.QueryRowx("SELECT * FROM todos WHERE id = $1", id).StructScan(&todo); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &todo, err
}

func (m *pgManager) CreateTodo(todo *models.Todo) (*models.Todo, error) {
	tx, err := m.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var newTodo models.Todo
	// Just like JS, we use "``" for templating strings.
	if err := tx.QueryRowx(`
        INSERT INTO todos (title, day, month, year, completed, description) VALUES
			($1, $2, $3, $4, $5, $6) RETURNING *`,
		todo.Title, todo.Day, todo.Month, todo.Year, todo.Completed, todo.Description,
	).StructScan(&newTodo); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &newTodo, err
}

func (m *pgManager) UpdateTodo(diff *models.Todo, id int64) (*models.Todo, error) {
	tx, err := m.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	todo := &models.Todo{}
	// The following query uses two functions that you probably didn't encounter in the core
	// curriculum: coalesce and nullif. The first takes any number of arguments and returns
	// the first non-null one it finds (if they are all null then it returns null).
	//
	// The second one takes two arguments and returns null if they match. In our setup, if
	// a user doesn't submit any value for any of the string fields, then the value on the
	// `diff` model will be an empty string since that is the zero-value for the string type.
	// Thus nullif will return false, and then the current value is what will be used in the
	// database.
	//
	// This poses a problem when updating the completed field -- the zero-value for a bool is
	// false, but we only want to update the field if the user explicitly includes it in the
	// request body. There's a few ways we could handle this, but for now we'll just require
	// users to use the ToggleTodo endpoint to change this value.
	if err := tx.QueryRowx(`
		UPDATE todos
		   SET
			   title       = coalesce(nullif($2, ''), title),
			   day 	       = coalesce(nullif($3, ''), day),
			   month       = coalesce(nullif($4, ''), month),
			   year        = coalesce(nullif($5, ''), year),
			   description = coalesce(nullif($6, ''), description)
		 WHERE id = $1
	 RETURNING *`,
		id, diff.Title, diff.Day, diff.Month, diff.Year, diff.Description).StructScan(todo); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return todo, nil
}

func (m *pgManager) DeleteTodo(id int64) (*models.Todo, error) {
	tx, err := m.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	todo := &models.Todo{}
	if err := tx.QueryRowx("DELETE FROM todos WHERE id = $1 RETURNING *", id).StructScan(todo); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return todo, nil
}

func (m *pgManager) ToggleTodo(id int64) (*models.Todo, error) {
	tx, err := m.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var completed bool
	if err := tx.QueryRowx("SELECT completed FROM todos WHERE id = $1", id).Scan(&completed); err != nil {
		return nil, err
	}

	todo := &models.Todo{}
	if err := tx.QueryRowx("UPDATE todos SET completed = $1 WHERE id = $2 RETURNING *",
		!completed, id).StructScan(todo); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return todo, nil
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
