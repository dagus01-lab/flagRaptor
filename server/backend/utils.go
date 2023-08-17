package main

import (
	"net/http"
	"regexp"

	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
)

var (
	sessionName     = "SESSIONID"
	sessionLifetime = 3600 // Session lifetime in seconds (1 hour)

	//GAME CONFIGURATION
	// CHANGE THIS
	TEAM            = 9
	NUMBER_OF_TEAMS = 2
	TEAM_TOKEN      = "" // team token for flag submission

	WEB_PASSWORD = "password"
	API_TOKEN    = "token"
	SECRET_KEY   = "not_secret_key"

	// Teams
	TEAM_FORMAT = "10.60.x.1"
	re          = regexp.MustCompile("x")
	TEAM_IP     = "10.60.1.1"
	TEAMS       = make([]string, 0)

	CLIENT_PORT = 5050

	//TEAMS.remove(TEAM_IP)
	NOP_TEAM = "10.60.0.1" //TEAM_FORMAT.format(0) // this will be used to ask for flag ids' service list by CCIT exploits

	ROUND_DURATION = 120
	FLAG_ALIVE     = 5 * ROUND_DURATION
	FLAG_FORMAT    = "[A-Z0-9]{31}=" // /^[A-Z0-9]{31}=$/

	FLAGID_URL = "" //"http://10.10.0.1:8081/flagIds" # flag_ids endpoint, leave blank if none

	SUB_PROTOCOL     = "ccit" // submitter protocol. Valid values are 'dummy', 'ccit', 'faust'
	SUB_LIMIT        = 1      // number of requests per interval
	SUB_INTERVAL     = 20     // interval duration
	SUB_PAYLOAD_SIZE = 500    // max flag per request
	SUB_URL          = "http://localhost:8000/flags"

	//Don't worry about this
	DB_NSUB = "NOT_SUBMITTED"
	DB_SUB  = "SUBMITTED"
	DB_SUCC = "SUCCESS"
	DB_ERR  = "ERROR"
	DB_EXP  = "EXPIRED"

	DATABASE = "instance/flagWarehouse.sqlite"
	///////////////
)

type WebSocketClient struct {
	connection  *websocket.Conn
	lastMinutes int
	username    string
}
type ScriptRunner struct {
	user      string
	addresses []string
	exploits  map[string]bool
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isAuthenticated(r) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func isAuthenticated(r *http.Request) bool {
	session, err := store.Get(r, sessionName)
	if err != nil {
		return false
	}

	userID, ok := session.Values["authenticatedUserID"].(int)
	return ok && userID != 0
}

func getAuthenticatedUsername(r *http.Request) (string, bool) {
	session, _ := store.Get(r, sessionName)
	userID, ok := session.Values["authenticatedUserName"].(string)
	return userID, ok
}

func setAuthenticatedUsername(r *http.Request, w http.ResponseWriter, username string) {
	session, err := store.Get(r, sessionName)
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}
	session.Options.MaxAge = sessionLifetime
	session.Values["authenticatedUserName"] = username
	session.Save(r, w)
}

func clearAuthenticatedUsername(r *http.Request, w http.ResponseWriter) {
	session, _ := store.Get(r, sessionName)
	session.Options.MaxAge = -1
	session.Save(r, w)
}

func initSessionManager() {
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   sessionLifetime,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
}
