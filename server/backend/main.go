package main

import (
	"database/sql"
	"flag"
	"log"
	"myflagsubmitter/common"
	"net/http"
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
	fileName := flag.String("f", "config.yaml", "configuration file name")
	flag.Parse()

	cfg, err := NewConfig(*fileName)
	if err != nil {
		log.Println("Error on reading configuration from file")
	}

	// Initialize the SQLite database
	db = initDatabase()
	err = createTables()
	if err != nil {
		log.Fatal("Error on creating tables:", err)
	}
	scriptRunners, err = get_stopped_exploits()
	if err != nil {
		log.Fatal("Error on getting stopped exploits:", err)
	}
	defer db.Close()

	//Initialize the session manager
	store = sessions.NewCookieStore([]byte(cfg.ServerConf.SecretKey))
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

	go submission_loop(&cfg)
	log.Println("Server listening on port 5000")
	http.ListenAndServe(":5000", appRouter)
}

func initDatabase() *sql.DB {
	db, err := sql.Open("sqlite3", cfg.ServerConf.DataBase)
	if err != nil {
		log.Fatal(err)
	}
	return db
}
