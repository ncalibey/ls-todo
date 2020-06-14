package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"ls-todo/internal/db"
	"ls-todo/internal/models"
)

// Server is the HTTP main that handles requests.
type Server interface {
	http.Handler

	// HandleGetTodos retrieves all todos.
	HandleGetTodos(w http.ResponseWriter, r *http.Request)
	// HandleGetTodo retrieves a single todo.
	HandleGetTodo(w http.ResponseWriter, r *http.Request)
	// HandleCreateTodo creates a new todo.
	HandleCreateTodo(w http.ResponseWriter, r *http.Request)
	// HandleUpdateTodo updates a todo.
	HandleUpdateTodo(w http.ResponseWriter, r *http.Request)
	// HandleDeleteTodo deletes a todo.
	HandleDeleteTodo(w http.ResponseWriter, r *http.Request)
	// HandleToggleTodo toggles a todo's completed status.
	HandleToggleTodo(w http.ResponseWriter, r *http.Request)
}

// server implements Server for "production". In other words, this is the live server used
// in the application.
type server struct {
	http.Handler

	db db.PGManager
}

// New returns a new Server instance. Notice how we return the interface and not the struct.
// Likewise, we use the PGManager interface instead of a pgManager struct. This allows us to
// pass in a mock database that implements the PGManager interface for when we want to do
// unit tests.
func New(router *mux.Router, db db.PGManager) Server {
	// This creates a new *server struct instance. Notice the pointer (&): this means when
	// the server is returned it will be the same place in memory when used elsewhere (i.e.
	// the struct isn't copied).
	server := &server{
		Handler: router,
		db:      db,
	}
	// We set up our routes as part of the constructor function.
	server.routes(router)
	return server
}

// routes attaches all of the handler functions for the api paths that we need to handle.
func (s *server) routes(router *mux.Router) {
	router.HandleFunc("/api/todos", s.HandleGetTodos).Methods("GET")
	router.HandleFunc("/api/todos/{id}", s.HandleGetTodo).Methods("GET")
	router.HandleFunc("/api/todos", s.HandleCreateTodo).Methods("POST")
	router.HandleFunc("/api/todos/{id}", s.HandleUpdateTodo).Methods("PUT")
	router.HandleFunc("/api/todos/{id}", s.HandleDeleteTodo).Methods("DELETE")
	router.HandleFunc("/api/todos/{id}/toggle_completed", s.HandleToggleTodo).Methods("POST")
}

func (s *server) HandleGetTodos(w http.ResponseWriter, r *http.Request) {
	// First, we make our call to the database. If we get an error, we return and ISE
	// (Internal Server Error -- 500). This is because the only error we should get
	// is one where the database fails to perform the query. An empty result set is
	// fine.
	todos, err := s.db.GetTodos()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// the `json.NewEncoder` needs a data type that satisfies the `io.Writer` interface,
	// which the `http.ResponseWriter` hapens to do! Thus, to send JSON back in the response
	// body, we create a new encoder using our response writer, and then encode the todos.
	if err := json.NewEncoder(w).Encode(todos); err != nil {
		// We return an ISE here because it means something went wrong with the encoding
		// process, and is not a user error.
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *server) HandleGetTodo(w http.ResponseWriter, r *http.Request) {
	// `mux.Vars` extracts the identifiers found in the path (in this case the `id` in
	// `/api/todos/{id}`.
	vars := mux.Vars(r)
	// Since the id is a string in the URL, we need to convert it to an int64 (since the
	// todo model's ID field is an int64).
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		// We send a 400 here because realistically the only reason this would fail is
		// because the user sent a non-integer value in the slug.
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	todo, err := s.db.GetTodo(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// We have to check for the condition where no todo was found. In that case it should
	// be nil.
	if todo == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(todo); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *server) HandleCreateTodo(w http.ResponseWriter, r *http.Request) {
	// First, we decode the JSON into a Todo struct.
	var todo models.Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		// While it's arguable that we should return an ISE in case some went wrong
		// with the decoding, the likely reason why that would happen is because of
		// bad JSON sent in the request body.
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	todoWithID, err := s.db.CreateTodo(&todo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(todoWithID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *server) HandleUpdateTodo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var diff models.Todo
	if err := json.NewDecoder(r.Body).Decode(&diff); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	todo, err := s.db.UpdateTodo(&diff, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if todo == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(todo); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *server) HandleDeleteTodo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	todo, err := s.db.DeleteTodo(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if todo == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(todo); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *server) HandleToggleTodo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	todo, err := s.db.ToggleTodo(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if todo == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(todo); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
