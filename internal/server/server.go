package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"

	"ls-todo/internal/models"
)

// data is an in-memory "store" that holds all the todos on the main. It uses
// a mutex to ensure saftey accessing the data between threads.
type data struct {
	todos []*models.Todo
	mu    *sync.Mutex
	nextID int64
}

var appData = &data{
	nextID: 4,
	todos: []*models.Todo{
		{
			ID: 1,
			Title:     "Todo 1",
			Completed: false,
		},
		{
			ID: 2,
			Title: "Todo 2",
			Month: "01",
			Day:   "01",
			Year:  "2018",
		},
		{
			ID: 3,
			Title:     "Todo 3",
			Completed: false,
		},
	},
	mu: &sync.Mutex{},
}

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
}

// New returns a new Server instance. Notice how we return the interface and not the struct.
func New(router *mux.Router) Server {
	server := &server{Handler: router}
	server.routes(router)

	return server
}

func (s *server) routes(router *mux.Router) {
	router.HandleFunc("/api/todos", s.HandleGetTodos).Methods("GET")
	router.HandleFunc("/api/todos/{id}", s.HandleGetTodo).Methods("GET")
	router.HandleFunc("/api/todos", s.HandleCreateTodo).Methods("POST")
	router.HandleFunc("/api/todos/{id}", s.HandleUpdateTodo).Methods("PUT")
	router.HandleFunc("/api/todos/{id}", s.HandleDeleteTodo).Methods("DELETE")
	router.HandleFunc("/api/todos/{id}/toggle_completed", s.HandleToggleTodo).Methods("POST")
}

func (s *server) HandleGetTodos(w http.ResponseWriter, r *http.Request) {
	appData.mu.Lock()
	defer appData.mu.Unlock()

	todos := appData.todos
	if err := json.NewEncoder(w).Encode(todos); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *server) HandleGetTodo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	appData.mu.Lock()
	defer appData.mu.Unlock()

	var todo *models.Todo
	for _, t := range appData.todos {
		if t.ID == id {
			todo = t
		}
	}
	if todo == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(todo); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *server) HandleCreateTodo(w http.ResponseWriter, r *http.Request) {
	appData.mu.Lock()
	defer appData.mu.Unlock()

	var todo models.Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	todo.ID = appData.nextID
	appData.nextID += 1
	appData.todos = append(appData.todos, &todo)

	if err := json.NewEncoder(w).Encode(todo); err != nil {
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

	appData.mu.Lock()
	defer appData.mu.Unlock()

	var todo *models.Todo
	for _, t := range appData.todos {
		if t.ID == id {
			t.Completed = !t.Completed
			todo = t
		}
	}
	if todo == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	updateTodo(todo, diff)
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

	appData.mu.Lock()
	defer appData.mu.Unlock()

	var todo *models.Todo
	var idx int

	for i, t := range appData.todos {
		if t.ID == id {
			idx = i
			todo = t
		}
	}
	if todo == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	appData.todos = append(appData.todos[:idx], appData.todos[idx+1:]...)

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

	appData.mu.Lock()
	defer appData.mu.Unlock()

	var todo *models.Todo
	for _, t := range appData.todos {
		if t.ID == id {
			t.Completed = !t.Completed
			todo = t
		}
	}
	if todo == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(todo); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

//////////////////////////////////////////////////////////////////////////////////////////
//// Helpers /////////////////////////////////////////////////////////////////////////////

// updateTodo updates todo with the non-zero values in diff.
//
// NOTE: it assumes diff does not just replace with the same value. It also does not check
// the completed or ID fields.
func updateTodo(todo *models.Todo, diff models.Todo) {
	if diff.Title != "" {
		todo.Title = diff.Title
	}
	if diff.Day != "" {
		todo.Day = diff.Day
	}
	if diff.Month != "" {
		todo.Month = diff.Month
	}
	if diff.Year != "" {
		todo.Year = diff.Year
	}
	if diff.Description != "" {
		todo.Description = diff.Description
	}
}
