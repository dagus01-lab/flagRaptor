package common

import "time"

type SubmitterFormat struct {
	SUB_ACCEPTED, SUB_INVALID, SUB_OLD, SUB_YOUR_OWN, SUB_STOLEN, SUB_NOP, SUB_NOT_AVAILABLE string
}

type ResponseItem struct {
	Flag    string `json:"flag"`
	Message string `json:"message"`
}

// Flag represents the flag data structure
type Flag struct {
	Flag           string `json:"flag"`
	Username       string `json:"username"`
	ExploitName    string `json:"exploit_name"`
	TeamIp         string `json:"team_ip"`
	Time           string `json:"time"`
	Status         string `json:"status"`
	ServerResponse string `json:"server_response"`
}
type UploadFlagRequestBody struct {
	Username string `json:"username"`
	Flags    []Flag `json:"flags"`
}
type FlagSubmitterConfig struct {
	FlagFormat    string        `json:"flag_format"`
	RoundDuration time.Duration `json:"round_duration"`
	Teams         []string      `json:"teams"`
	NopTeam       string        `json:"nop_team"`
	FlagidUrl     string        `json:"flagid_url"`
	ClientPort    int           `json:"client_port"`
}
