package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"gopkg.in/yaml.v3"
)

type User struct {
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

type RawConfig struct {
	SessionLifetime int    `yaml:"sessionLifetime,omitempty"`
	Team            int    `yaml:"team,omitempty"`
	NumberTeams     int    `yaml:"numberTeams,omitempty"`
	WebPassword     string `yaml:"webPassword,omitempty"`
	APIToken        string `yaml:"apiToken,omitempty"`
	TeamFormat      string `yaml:"teamFormat,omitempty"`
	TeamIP          string `yaml:"teamIP,omitempty"`
	Wildcard        string `yaml:"wildcard,omitempty"`
	ClientPort      int    `yaml:"clientPort,omitempty"`
	NopTeam         string `yaml:"nopTeam,omitempty"`
	RoundDuration   int    `yaml:"roundDuration,omitempty"`
	FlagFormat      string `yaml:"flagFormat,omitempty"`
	FlagIDUrl       string `yaml:"flagIDurl,omitempty"`
	SubProtocol     string `yaml:"subProtocol,omitempty"`
	SubLimit        int    `yaml:"subLimit,omitempty"`
	SubInterval     int    `yaml:"subInterval,omitempty"`
	SubPayloadSize  int    `yaml:"subPayloadSize,omitempty"`
	SubUrl          string `yaml:"subUrl,omitempty"`
	DBNSUB          string `yaml:"dbnsub,omitempty"`
	DBSUB           string `yaml:"dbsub,omitempty"`
	DBSUCC          string `yaml:"dbsucc,omitempty"`
	DBERR           string `yaml:"dberr,omitempty"`
	DBEXP           string `yaml:"dbexp,omitempty"`
	DataBase        string `yaml:"database,omitempty"`
	AuthEnable      bool   `yaml:"authEnable,omitempty"`
	Users           []User `yaml:"users,omitempty"`
}

type Config struct {
	SessionName     string
	SessionLifetime int
	Team            int
	NumberTeams     int
	TeamToken       string
	WebPassword     string
	APIToken        string
	SecretKey       string
	TeamFormat      string
	re              *regexp.Regexp
	TeamIP          string
	Teams           []string
	ClientPort      int
	NopTeam         string
	RoundDuration   int
	FlagAlive       int
	FlagFormat      string
	FlagIDUrl       string
	SubProtocol     string
	SubLimit        int
	SubInterval     int
	SubPayloadSize  int
	SubUrl          string
	DBNSUB          string
	DBSUB           string
	DBSUCC          string
	DBERR           string
	DBEXP           string
	DataBase        string
	AuthEnable      bool
	Users           []User
}

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

func NewConfig(fileName string) (Config, error) {
	configFile, err := ioutil.ReadFile(fileName)
	if err != nil {
		return Config{}, fmt.Errorf("could not read the config file %s: %w", fileName, err)
	}

	var rawConfig RawConfig
	if err := yaml.Unmarshal(configFile, &rawConfig); err != nil {
		return Config{}, fmt.Errorf("could not parse the config file %s: %w", fileName, err)
	}

	sessionName := "SESSIONID"
	sessionLifetime := 3600 // Session lifetime in seconds (1 hour)
	if rawConfig.SessionLifetime != 0 {
		sessionLifetime = rawConfig.SessionLifetime
	}

	//GAME CONFIGURATION
	// CHANGE THIS
	TEAM := 9
	if rawConfig.Team != 0 {
		TEAM = rawConfig.Team
	}
	NUMBER_OF_TEAMS := 2
	if rawConfig.NumberTeams != 0 {
		NUMBER_OF_TEAMS = rawConfig.NumberTeams
	}
	TEAM_TOKEN := "" // team token for flag submission

	WEB_PASSWORD := "password"
	if rawConfig.WebPassword != "" {
		WEB_PASSWORD = rawConfig.WebPassword
	}
	API_TOKEN := "token"
	if rawConfig.APIToken != "" {
		API_TOKEN = rawConfig.APIToken
	}
	SECRET_KEY := "not_secret_key"

	// Teams
	TEAM_FORMAT := "10.60.x.1"
	if rawConfig.TeamFormat != "" {
		TEAM_FORMAT = rawConfig.TeamFormat
	}
	re := regexp.MustCompile("x")
	if rawConfig.Wildcard != "" {
		re = regexp.MustCompile(rawConfig.Wildcard)
	}

	TEAM_IP := "10.60.1.1"
	if rawConfig.TeamIP != "" {
		TEAM_IP = rawConfig.TeamIP
	}

	TEAMS := make([]string, 0)
	for i := 1; i < NUMBER_OF_TEAMS; i++ {
		TEAMS = append(TEAMS, re.ReplaceAllString(TEAM_FORMAT, strconv.Itoa(i)))
	}

	CLIENT_PORT := 5050
	if rawConfig.ClientPort != 0 {
		CLIENT_PORT = rawConfig.ClientPort
	}

	//TEAMS.remove(TEAM_IP)
	NOP_TEAM := "10.60.0.1" //TEAM_FORMAT.format(0) // this will be used to ask for flag ids' service list by CCIT exploits
	if rawConfig.NopTeam != "" {
		NOP_TEAM = rawConfig.NopTeam
	}

	ROUND_DURATION := 120
	if rawConfig.RoundDuration != 0 {
		ROUND_DURATION = rawConfig.RoundDuration
	}
	FLAG_ALIVE := 5 * ROUND_DURATION
	FLAG_FORMAT := "[A-Z0-9]{31}="
	if rawConfig.FlagFormat != "" {
		FLAG_FORMAT = rawConfig.FlagFormat
	}

	FLAGID_URL := "" //"http://10.10.0.1:8081/flagIds" # flag_ids endpoint, leave blank if none
	if rawConfig.FlagIDUrl != "" {
		FLAGID_URL = rawConfig.FlagIDUrl
	}

	SUB_PROTOCOL := "ccit" // submitter protocol. Valid values are 'dummy', 'ccit', 'faust'
	if rawConfig.SubProtocol != "" {
		SUB_PROTOCOL = rawConfig.SubProtocol
	}
	SUB_LIMIT := 1 // number of requests per interval
	if rawConfig.SubLimit != 0 {
		SUB_LIMIT = rawConfig.SubLimit
	}
	SUB_INTERVAL := 20 // interval duration
	if rawConfig.SubInterval != 0 {
		SUB_INTERVAL = rawConfig.SubInterval
	}
	SUB_PAYLOAD_SIZE := 500 // max flag per request
	if rawConfig.SubPayloadSize != 0 {
		SUB_PAYLOAD_SIZE = rawConfig.SubPayloadSize
	}
	SUB_URL := "http://localhost:8000/flags"
	if rawConfig.SubUrl != "" {
		SUB_URL = rawConfig.SubUrl
	}

	//Don't worry about this
	DB_NSUB := "NOT_SUBMITTED"
	if rawConfig.DBNSUB != "" {
		DB_NSUB = rawConfig.DBNSUB
	}
	DB_SUB := "SUBMITTED"
	if rawConfig.DBSUB != "" {
		DB_SUB = rawConfig.DBSUB
	}
	DB_SUCC := "SUCCESS"
	if rawConfig.DBSUCC != "" {
		DB_SUCC = rawConfig.DBSUCC
	}
	DB_ERR := "ERROR"
	if rawConfig.DBERR != "" {
		DB_ERR = rawConfig.DBERR
	}
	DB_EXP := "EXPIRED"
	if rawConfig.DBEXP != "" {
		DB_EXP = rawConfig.DBEXP
	}

	DATABASE := "instance/flagWarehouse.sqlite"
	if rawConfig.DataBase != "" {
		DATABASE = rawConfig.DataBase
	}
	AuthEnabled := rawConfig.AuthEnable
	users := make([]User, 0)
	if rawConfig.Users != nil {
		users = rawConfig.Users
	}

	cfg := Config{
		SessionName:     sessionName,
		SessionLifetime: sessionLifetime,
		Team:            TEAM,
		NumberTeams:     NUMBER_OF_TEAMS,
		TeamToken:       TEAM_TOKEN,
		WebPassword:     WEB_PASSWORD,
		APIToken:        API_TOKEN,
		SecretKey:       SECRET_KEY,
		TeamFormat:      TEAM_FORMAT,
		re:              re,
		TeamIP:          TEAM_IP,
		Teams:           TEAMS,
		ClientPort:      CLIENT_PORT,
		NopTeam:         NOP_TEAM,
		RoundDuration:   ROUND_DURATION,
		FlagAlive:       FLAG_ALIVE,
		FlagFormat:      FLAG_FORMAT,
		FlagIDUrl:       FLAGID_URL,
		SubProtocol:     SUB_PROTOCOL,
		SubLimit:        SUB_LIMIT,
		SubInterval:     SUB_INTERVAL,
		SubPayloadSize:  SUB_PAYLOAD_SIZE,
		SubUrl:          SUB_URL,
		DBNSUB:          DB_NSUB,
		DBSUB:           DB_SUB,
		DBSUCC:          DB_SUCC,
		DBERR:           DB_ERR,
		DBEXP:           DB_EXP,
		DataBase:        DATABASE,
		AuthEnable:      AuthEnabled,
		Users:           users,
	}

	return cfg, err
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
	session, err := store.Get(r, cfg.SessionName)
	if err != nil {
		return false
	}

	userName, ok := session.Values["authenticatedUserName"].(string)
	return ok && userName != ""
}

func getAuthenticatedUsername(r *http.Request) (string, bool) {
	session, _ := store.Get(r, cfg.SessionName)
	if isAuthenticated(r) {
		userID, ok := session.Values["authenticatedUserName"].(string)
		return userID, ok
	} else {
		return "", false
	}
}

func setAuthenticatedUsername(r *http.Request, w http.ResponseWriter, username string) {
	session, err := store.Get(r, cfg.SessionName)
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}
	session.Options.MaxAge = cfg.SessionLifetime
	session.Values["authenticatedUserName"] = username
	session.Save(r, w)
}

func clearAuthenticatedUsername(r *http.Request, w http.ResponseWriter) {
	session, _ := store.Get(r, cfg.SessionName)
	session.Options.MaxAge = -1
	session.Save(r, w)
}

func initSessionManager() {
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   cfg.SessionLifetime,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
}
