package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"gopkg.in/yaml.v3"
)

type User struct {
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}
type ServerRawConf struct {
	SessionLifetime int    `yaml:"sessionLifetime,omitempty"`
	DataBase        string `yaml:"database,omitempty"`
	ClientPort      int    `yaml:"clientPort,omitempty"`
	APIToken        string `yaml:"apiToken,omitempty"`
}
type ServerConf struct {
	SessionName     string
	SessionLifetime int
	DataBase        string
	ClientPort      int
	APIToken        string
	SecretKey       string
}
type GameRawConf struct {
	NumberTeams   int    `yaml:"numberTeams,omitempty"`
	TeamFormat    string `yaml:"teamFormat,omitempty"`
	TeamToken     string `yaml:"teamToken,omitempty"`
	TeamIP        string `yaml:"teamIP,omitempty"`
	Wildcard      string `yaml:"wildcard,omitempty"`
	NopTeam       string `yaml:"nopTeam,omitempty"`
	RoundDuration int    `yaml:"roundDuration,omitempty"`
	FlagFormat    string `yaml:"flagFormat,omitempty"`
	FlagIDUrl     string `yaml:"flagIDurl,omitempty"`
	FlagAlive     int    `yaml:"flagAlive,omitempty"`
}
type GameConf struct {
	NumberTeams   int
	Teams         []string
	TeamIP        string
	TeamToken     string
	NopTeam       string
	RoundDuration time.Duration
	FlagFormat    string
	FlagIDUrl     string
	FlagAlive     time.Duration
}
type SubmissionRawConf struct {
	SubProtocol    string `yaml:"subProtocol,omitempty"`
	SubLimit       int    `yaml:"subLimit,omitempty"`
	SubInterval    int    `yaml:"subInterval,omitempty"`
	SubPayloadSize int    `yaml:"subPayloadSize,omitempty"`
	SubUrl         string `yaml:"subUrl,omitempty"`
	DBNSUB         string `yaml:"dbnsub,omitempty"`
	DBSUB          string `yaml:"dbsub,omitempty"`
	DBSUCC         string `yaml:"dbsucc,omitempty"`
	DBERR          string `yaml:"dberr,omitempty"`
	DBEXP          string `yaml:"dbexp,omitempty"`
}
type SubmissionConf struct {
	SubProtocol    string
	SubLimit       int
	SubInterval    time.Duration
	SubPayloadSize int
	SubUrl         string
	DBNSUB         string
	DBSUB          string
	DBSUCC         string
	DBERR          string
	DBEXP          string
}
type AuthConf struct {
	WebPassword string `yaml:"webPassword,omitempty"`
	AuthEnable  bool   `yaml:"authEnable,omitempty"`
	Users       []User `yaml:"users,omitempty"`
}

type RawConfig struct {
	ServerConf     *ServerRawConf     `yaml:"serverConf,omitempty"`
	GameConf       *GameRawConf       `yaml:"gameConf,omitempty"`
	SubmissionConf *SubmissionRawConf `yaml:"submissionConf,omitempty"`
	AuthConf       *AuthConf          `yaml:"authConf,omitempty"`
}

type Config struct {
	ServerConf     ServerConf
	GameConf       GameConf
	SubmissionConf SubmissionConf
	AuthConf       AuthConf
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

	//SERVER CONFIGURATION
	sessionName := "SESSIONID"
	sessionLifetime := 3600 // Session lifetime in seconds (1 hour)
	if rawConfig.ServerConf.SessionLifetime != 0 {
		sessionLifetime = rawConfig.ServerConf.SessionLifetime
	}
	DATABASE := "instance/flagWarehouse.sqlite"
	if rawConfig.ServerConf.DataBase != "" {
		DATABASE = rawConfig.ServerConf.DataBase
	}
	CLIENT_PORT := 5050
	if rawConfig.ServerConf.ClientPort > 0 && rawConfig.ServerConf.ClientPort < 65535 {
		CLIENT_PORT = rawConfig.ServerConf.ClientPort
	}
	API_TOKEN := "token"
	if rawConfig.ServerConf.APIToken != "" {
		API_TOKEN = rawConfig.ServerConf.APIToken
	}
	SECRET_KEY := "not_secret_key"

	serverConf := ServerConf{
		SessionName:     sessionName,
		SessionLifetime: sessionLifetime,
		DataBase:        DATABASE,
		ClientPort:      CLIENT_PORT,
		APIToken:        API_TOKEN,
		SecretKey:       SECRET_KEY,
	}

	//GAME CONFIGURATION
	NUMBER_OF_TEAMS := 2
	if rawConfig.GameConf.NumberTeams > 0 && rawConfig.GameConf.NumberTeams < 256 {
		NUMBER_OF_TEAMS = rawConfig.GameConf.NumberTeams
	}
	TEAM_FORMAT := "10.60.x.1"
	if rawConfig.GameConf.TeamFormat != "" {
		TEAM_FORMAT = rawConfig.GameConf.TeamFormat
	}
	re := regexp.MustCompile("x")
	if rawConfig.GameConf.Wildcard != "" {
		re = regexp.MustCompile(rawConfig.GameConf.Wildcard)
	}
	TEAM_IP := "10.60.1.1"
	if rawConfig.GameConf.TeamIP != "" {
		TEAM_IP = rawConfig.GameConf.TeamIP
	}
	TEAM_TOKEN := "" // team token for flag submission
	if rawConfig.GameConf.TeamToken != "" {
		TEAM_TOKEN = rawConfig.GameConf.TeamToken
	}
	TEAMS := make([]string, 0)
	for i := 1; i < NUMBER_OF_TEAMS; i++ {
		team := re.ReplaceAllString(TEAM_FORMAT, strconv.Itoa(i))
		if team != TEAM_IP {
			TEAMS = append(TEAMS, team)
		}
	}
	NOP_TEAM := "10.60.0.1" // this will be used to ask for flag ids' service list by CCIT exploits
	if rawConfig.GameConf.NopTeam != "" {
		NOP_TEAM = rawConfig.GameConf.NopTeam
	}
	ROUND_DURATION := time.Duration(120) * time.Second
	if rawConfig.GameConf.RoundDuration != 0 {
		ROUND_DURATION = time.Duration(rawConfig.GameConf.RoundDuration) * time.Second
	}
	FLAG_FORMAT := "[A-Z0-9]{31}="
	if rawConfig.GameConf.FlagFormat != "" {
		FLAG_FORMAT = rawConfig.GameConf.FlagFormat
	}
	FLAGID_URL := "" // flag_ids endpoint, leave blank if none
	if rawConfig.GameConf.FlagIDUrl != "" {
		FLAGID_URL = rawConfig.GameConf.FlagIDUrl
	}
	FLAG_ALIVE := 5 * ROUND_DURATION
	if rawConfig.GameConf.FlagAlive > 0 {
		FLAG_ALIVE = time.Duration(rawConfig.GameConf.FlagAlive) * ROUND_DURATION
	}

	gameConf := GameConf{
		NumberTeams:   NUMBER_OF_TEAMS,
		Teams:         TEAMS,
		TeamIP:        TEAM_IP,
		TeamToken:     TEAM_TOKEN,
		NopTeam:       NOP_TEAM,
		RoundDuration: ROUND_DURATION,
		FlagFormat:    FLAG_FORMAT,
		FlagIDUrl:     FLAGID_URL,
		FlagAlive:     FLAG_ALIVE,
	}

	//SUBMISSION CONFIGURATION
	SUB_PROTOCOL := "ccit" // submitter protocol. Valid values are 'dummy', 'ccit', 'faust'
	if rawConfig.SubmissionConf.SubProtocol != "" {
		SUB_PROTOCOL = rawConfig.SubmissionConf.SubProtocol
	}
	SUB_LIMIT := 1 // number of requests per interval
	if rawConfig.SubmissionConf.SubLimit != 0 {
		SUB_LIMIT = rawConfig.SubmissionConf.SubLimit
	}
	SUB_INTERVAL := 20 * time.Second // interval duration
	if rawConfig.SubmissionConf.SubInterval != 0 {
		SUB_INTERVAL = time.Duration(rawConfig.SubmissionConf.SubInterval) * time.Second
	}
	SUB_PAYLOAD_SIZE := 500 // max flag per request
	if rawConfig.SubmissionConf.SubPayloadSize != 0 {
		SUB_PAYLOAD_SIZE = rawConfig.SubmissionConf.SubPayloadSize
	}
	SUB_URL := "http://localhost:8000/flags"
	if rawConfig.SubmissionConf.SubUrl != "" {
		SUB_URL = rawConfig.SubmissionConf.SubUrl
	}
	DB_NSUB := "NOT_SUBMITTED"
	if rawConfig.SubmissionConf.DBNSUB != "" {
		DB_NSUB = rawConfig.SubmissionConf.DBNSUB
	}
	DB_SUB := "SUBMITTED"
	if rawConfig.SubmissionConf.DBSUB != "" {
		DB_SUB = rawConfig.SubmissionConf.DBSUB
	}
	DB_SUCC := "SUCCESS"
	if rawConfig.SubmissionConf.DBSUCC != "" {
		DB_SUCC = rawConfig.SubmissionConf.DBSUCC
	}
	DB_ERR := "ERROR"
	if rawConfig.SubmissionConf.DBERR != "" {
		DB_ERR = rawConfig.SubmissionConf.DBERR
	}
	DB_EXP := "EXPIRED"
	if rawConfig.SubmissionConf.DBEXP != "" {
		DB_EXP = rawConfig.SubmissionConf.DBEXP
	}

	submissionConf := SubmissionConf{
		SubProtocol:    SUB_PROTOCOL,
		SubLimit:       SUB_LIMIT,
		SubInterval:    SUB_INTERVAL,
		SubPayloadSize: SUB_PAYLOAD_SIZE,
		SubUrl:         SUB_URL,
		DBNSUB:         DB_NSUB,
		DBSUB:          DB_SUB,
		DBSUCC:         DB_SUCC,
		DBERR:          DB_ERR,
		DBEXP:          DB_EXP,
	}

	//AUTHENTICATION CONFIGURATION
	WEB_PASSWORD := "password"
	if rawConfig.AuthConf.WebPassword != "" {
		WEB_PASSWORD = rawConfig.AuthConf.WebPassword
	}

	AuthEnabled := rawConfig.AuthConf.AuthEnable
	users := make([]User, 0)
	if rawConfig.AuthConf.Users != nil {
		users = rawConfig.AuthConf.Users
	}

	authConf := AuthConf{
		WebPassword: WEB_PASSWORD,
		AuthEnable:  AuthEnabled,
		Users:       users,
	}

	cfg := Config{
		ServerConf:     serverConf,
		GameConf:       gameConf,
		SubmissionConf: submissionConf,
		AuthConf:       authConf,
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
	session, err := store.Get(r, cfg.ServerConf.SessionName)
	if err != nil {
		return false
	}

	userName, ok := session.Values["authenticatedUserName"].(string)
	return ok && userName != ""
}

func getAuthenticatedUsername(r *http.Request) (string, bool) {
	session, _ := store.Get(r, cfg.ServerConf.SessionName)
	if isAuthenticated(r) {
		userID, ok := session.Values["authenticatedUserName"].(string)
		return userID, ok
	} else {
		return "", false
	}
}

func setAuthenticatedUsername(r *http.Request, w http.ResponseWriter, username string) {
	session, err := store.Get(r, cfg.ServerConf.SessionName)
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}
	session.Options.MaxAge = cfg.ServerConf.SessionLifetime
	session.Values["authenticatedUserName"] = username
	session.Save(r, w)
}

func clearAuthenticatedUsername(r *http.Request, w http.ResponseWriter) {
	session, _ := store.Get(r, cfg.ServerConf.SessionName)
	session.Options.MaxAge = -1
	session.Save(r, w)
}

func initSessionManager() {
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   cfg.ServerConf.SessionLifetime,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
}
