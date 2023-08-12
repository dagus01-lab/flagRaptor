package main

import (
	"database/sql"
	"fmt"
	"myflagsubmitter/common"
	"strings"
	"time"
)

func createFlagsTable(db *sql.DB) error {
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS flags
		(
			flag            TEXT PRIMARY KEY,
			username        TEXT NOT NULL,
			exploit_name    TEXT NOT NULL,
			team_ip         TEXT NOT NULL,
			time            TEXT NOT NULL,
			status          TEXT DEFAULT 'NOT_SUBMITTED',
			server_response TEXT
		);
		
		CREATE INDEX IF NOT EXISTS idx_time ON flags (time);
	`
	_, err := db.Exec(createTableSQL)
	return err
}

func find_flags_by_names(flags []string) ([]common.Flag, error) {
	result := make([]common.Flag, 0)
	if len(flags) == 0 {
		return result, nil
	}
	query := "SELECT flag,username,exploit_name,team_ip,time,status,server_response FROM flags WHERE flag=\"" + flags[0] + "\""
	for _, flag := range flags {
		query += strings.Replace(" OR flag=? ", "?", "\""+flag+"\"", 1)
	}
	rows, err := db.Query(query)
	if err != nil {
		fmt.Println("Error executing query ", query, ": ", err)
		return nil, err
	}
	defer rows.Close()

	// save all the received flags into a list
	for rows.Next() {
		var flag common.Flag
		err := rows.Scan(&flag.Flag, &flag.Username, &flag.ExploitName, &flag.TeamIp, &flag.Time, &flag.Status, &flag.ServerResponse)
		if err != nil {
			fmt.Println("Error scanning row: ", err)
			return nil, err
		}
		result = append(result, flag)
	}

	return result, nil
}
func get_flags_before(lastMinutes int) ([]common.Flag, error) {
	result := make([]common.Flag, 0)
	time := time.Now().Add(-time.Duration(lastMinutes * int(time.Minute))).Format("2006-01-02 15:04:05")
	query := "SELECT flag,username,exploit_name,team_ip,time,status,server_response FROM flags WHERE time >= \"" + time + "\""
	rows, err := db.Query(query)
	if err != nil {
		fmt.Println("Error executing query: ", err)
		return nil, err
	}
	defer rows.Close()

	// save all the received flags into a list
	for rows.Next() {
		var flag common.Flag
		err := rows.Scan(&flag.Flag, &flag.Username, &flag.ExploitName, &flag.TeamIp, &flag.Time, &flag.Status, &flag.ServerResponse)
		if err != nil {
			fmt.Println("Error scanning row: ", err)
			return nil, err
		}
		result = append(result, flag)
	}

	return result, nil
}
func find_flags_by_username(username string) ([]common.Flag, error) {
	result := make([]common.Flag, 0)
	query := "SELECT flag,username,exploit_name,team_ip,time,status,server_response WHERE username=" + username
	rows, err := db.Query(query)
	if err != nil {
		fmt.Println("Error executing query: ", err)
		return nil, err
	}
	defer rows.Close()

	// save all the received flags into a list
	for rows.Next() {
		var flag common.Flag
		err := rows.Scan(&flag.Flag, &flag.Username, &flag.ExploitName, &flag.TeamIp, &flag.Time, &flag.Status, &flag.ServerResponse)
		if err != nil {
			fmt.Println("Error scanning row: ", err)
			return nil, err
		}
		result = append(result, flag)
	}

	return result, nil
}
