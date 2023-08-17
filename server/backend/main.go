package main

import (
	"database/sql"
	"fmt"
	"log"
	"myflagsubmitter/common"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
)

var webSocketClientsLock sync.Mutex
var clients = make([]WebSocketClient, 0)
var broadcast = make(chan []common.Flag)
var store *sessions.CookieStore
var scriptRunnersLock sync.Mutex
var scriptRunners = make([]ScriptRunner, 0)
var db *sql.DB
var dbLock sync.Mutex

func main() {
	for i := 1; i < NUMBER_OF_TEAMS; i++ {
		TEAMS = append(TEAMS, re.ReplaceAllString(TEAM_FORMAT, strconv.Itoa(i)))
	}

	// Initialize the SQLite database
	db = initDatabase()
	go createFlagsTable(db)
	defer db.Close()

	//Initialize the session manager
	store = sessions.NewCookieStore([]byte("your-secret-key"))
	initSessionManager()

	//set up a router for API
	appRouter := mux.NewRouter()
	appRouter.HandleFunc("/login", loginHandler).Methods("GET", "POST")
	appRouter.HandleFunc("/logout", logoutHandler).Methods("GET", "POST")
	appRouter.HandleFunc("/get_config", getConfigHandler).Methods("GET", "POST")
	appRouter.HandleFunc("/upload_flags", uploadFlagsHandler).Methods("GET", "POST")
	appRouter.HandleFunc("/get_flags", getFlagsHandler).Methods("GET", "POST")
	appRouter.HandleFunc("/update_flags", updateFlagsHandler)
	appRouter.HandleFunc("/restart_exploit", restartExploitHandler).Methods("GET", "POST")
	appRouter.HandleFunc("/stop_exploit", stopExploitHandler).Methods("GET", "POST")
	appRouter.HandleFunc("/get_stopped_exploits", getStoppedExploitsHandler).Methods("GET", "POST")
	appRouter.PathPrefix("/").Handler(http.FileServer(http.Dir("../frontend/dist")))
	//http.Handle("/", apiRouter)

	go submission_loop()
	fmt.Println("Server listening on port 5000")
	http.ListenAndServe(":5000", appRouter)

}

func initDatabase() *sql.DB {
	db, err := sql.Open("sqlite3", "instances/database.db")
	if err != nil {
		log.Fatal(err)
	}
	return db
}
