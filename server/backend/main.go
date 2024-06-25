package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
)

var webSocketClientsLock sync.Mutex
var clients = make([]WebSocketClient, 0)
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
	fmt.Println(`
	________ ___       ________  ________                           
	|\  _____\\  \     |\   __  \|\   ____\                          
	\ \  \__/\ \  \    \ \  \|\  \ \  \___|                          
	 \ \   __\\ \  \    \ \   __  \ \  \  ___                        
	  \ \  \_| \ \  \____\ \  \ \  \ \  \|\  \                       
	   \ \__\   \ \_______\ \__\ \__\ \_______\                      
	 ________  __________________|____________________  ________     
	|\   __  \|\   __  \|\   __  \|\___   ___\\   __  \|\   __  \    
	\ \  \|\  \ \  \|\  \ \  \|\  \|___ \  \_\ \  \|\  \ \  \|\  \   
	 \ \   _  _\ \   __  \ \   ____\   \ \  \ \ \  \\\  \ \   _  _\  
	  \ \  \\  \\ \  \ \  \ \  \___|    \ \  \ \ \  \\\  \ \  \\  \| 
	   \ \__\\ _\\ \__\ \__\ \__\        \ \__\ \ \_______\ \__\\ _\ 
	    \|__|\|__|\|__|\|__|\|__|         \|__|  \|_______|\|__|\|__|

	`)

	// Initialize the SQLite database
	db = initDatabase(cfg.ServerConf.DataBase)
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
	appRouter.HandleFunc("/check_auth", checkAuthHandler).Methods("GET", "POST")
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
	log.Println("Server listening on port ", cfg.ServerConf.ClientPort)
	addr := ":" + strconv.Itoa(cfg.ServerConf.ClientPort)
	http.ListenAndServe(addr, appRouter)
}
