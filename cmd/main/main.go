package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"

	"ls-todo/internal/config"
	"ls-todo/internal/db"
	"ls-todo/internal/server"
)

func main() {
	// First we get the environment variables of the application. If there is an error processing
	// these we shut down the application and log the error so we can see what went wrong.
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("error processing environment config: %v", err)
	}
	// Here we create the router that we will be using in our application, and pass it to the
	// constructor function of our server.
	router := mux.NewRouter()

	// Next, we open a connection to our PostgreSQL database. We then create a new PGManager
	// instance which is used for executing our queries. We then pass this to the server as
	// a dependency.
	connString := db.GetConnString(cfg)
	dbConn, err := sqlx.Connect("postgres", connString)
	if err != nil {
		log.Fatalf("error connection to database: %v", err)
	}
	// In order to prevent dangling open connections after our app closes, we use the `defer`
	// keyword. This ensures that the `dbConn.Close()` method will be called before the `main`
	// function finishes executing. It will also close the connection if there is an error that
	// crashes our program.
	defer dbConn.Close()
	pgManager := db.New(dbConn)

	s := server.New(router, pgManager)

	// By using an `if err :=`, we scope this `err` variable to the `if` block, meaning it shadows
	// the `err` variable on L15. Though we don't need to here, it would allow us to use the
	// previous `err` variable again after the `if` block.
	//
	// Since our server instance implements the `http.Handler` interface (because of our router), we
	// cann use it as the second argument to `http.ListenAndServe`. This makes Go use our router for
	// routing instead of the default router of the net/http package.
	if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), s); err != nil {
		log.Fatalf("error starting HTTP server: %v", err)
	}
}
