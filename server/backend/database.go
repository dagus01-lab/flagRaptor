package main

import (
	"database/sql"
	"fmt"
	"myflagsubmitter/common"
	"strings"
	"sync"
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
	dbLock.Lock()
	rows, err := db.Query(query)
	dbLock.Unlock()
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
	dbLock.Lock()
	rows, err := db.Query(query)
	dbLock.Unlock()
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
	dbLock.Lock()
	rows, err := db.Query(query)
	dbLock.Unlock()
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

func getFlagsToCheck(expiration_time string) ([]string, error) {

	flags := make([]string, 0)
	query := "SELECT flag FROM flags WHERE time > ? AND status = ?"
	// query the flags which are not expired yet and were not submitted
	dbLock.Lock()
	rows, err := db.Query(query, expiration_time, DB_NSUB)
	dbLock.Unlock()
	if err != nil {
		return flags, err
	}
	defer rows.Close()

	// save all the received flags into a list
	for rows.Next() && len(flags) < SUB_PAYLOAD_SIZE {
		var flag string
		err := rows.Scan(&flag)
		if err != nil {
			fmt.Println("Error scanning row: ", err)
			break
		}
		flags = append(flags, flag)
	}
	return flags, nil
}

func setOldFlagsAsExpired(expiration_time string) error {
	update_old_flags_query := "UPDATE flags SET status = ? WHERE time <= ?"
	dbLock.Lock()
	_, err := db.Exec(update_old_flags_query, DB_EXP, expiration_time)
	dbLock.Unlock()
	return err
}

func updateUploadedFlagsToDB(wg *sync.WaitGroup, accepted *int, old *int, nop *int, yours *int, invalid *int, not_available *int, item ResponseItem, submitterFormat SubmitterFormat, resultLock *sync.Mutex) {
	defer wg.Done()
	query := "UPDATE flags SET status = ?, server_response = ? WHERE flag = ?"
	if strings.Contains(strings.ToLower(submitterFormat.SUB_INVALID), strings.ToLower(item.Message)) {
		dbLock.Lock()
		_, err := db.Exec(query, DB_SUB, DB_ERR, item.Flag)
		dbLock.Unlock()
		if err != nil {
			fmt.Println("Error in updating flags: ", err)
		} else {
			//fmt.Println("Flag", item.Flag, "invalid")
			resultLock.Lock()
			*invalid += 1
			resultLock.Unlock()
		}

	} else if strings.Contains(strings.ToLower(submitterFormat.SUB_YOUR_OWN), strings.ToLower(item.Message)) {
		dbLock.Lock()
		_, err := db.Exec(query, DB_SUB, DB_ERR, item.Flag)
		dbLock.Unlock()
		if err != nil {
			fmt.Println("Error in updating flags: ", err)
		} else {
			//fmt.Println("Flag", item.Flag, "yours")
			resultLock.Lock()
			*yours += 1
			resultLock.Unlock()
		}
	} else if strings.Contains(strings.ToLower(submitterFormat.SUB_NOP), strings.ToLower(item.Message)) {
		dbLock.Lock()
		_, err := db.Exec(query, DB_SUB, DB_ERR, item.Flag)
		dbLock.Unlock()
		if err != nil {
			fmt.Println("Error in updating flags: ", err)
		} else {
			//fmt.Println("Flag", item.Flag, "of nop team")
			resultLock.Lock()
			*nop += 1
			resultLock.Unlock()
		}
	} else if strings.Contains(strings.ToLower(submitterFormat.SUB_OLD), strings.ToLower(item.Message)) {
		dbLock.Lock()
		_, err := db.Exec(query, DB_SUB, DB_EXP, item.Flag)
		dbLock.Unlock()
		if err != nil {
			fmt.Println("Error in updating flags: ", err)
		} else {
			//fmt.Println("Flag", item.Flag, "old")
			resultLock.Lock()
			*old += 1
			resultLock.Unlock()
		}
	} else if strings.Contains(strings.ToLower(submitterFormat.SUB_STOLEN), strings.ToLower(item.Message)) ||
		strings.Contains(strings.ToLower(submitterFormat.SUB_ACCEPTED), strings.ToLower(item.Message)) {
		dbLock.Lock()
		_, err := db.Exec(query, DB_SUB, DB_SUCC, item.Flag)
		dbLock.Unlock()
		if err != nil {
			fmt.Println("Error in updating flags: ", err)
		} else {
			//fmt.Println("Flag", item.Flag, "accepted")
			resultLock.Lock()
			*accepted += 1
			resultLock.Unlock()
		}
	} else if strings.Contains(strings.ToLower(submitterFormat.SUB_NOT_AVAILABLE), strings.ToLower(item.Message)) {
		dbLock.Lock()
		_, err := db.Exec(query, DB_SUB, DB_SUCC, item.Flag)
		dbLock.Unlock()
		if err != nil {
			fmt.Println("Error in updating flags: ", err)
		} else {
			//fmt.Println("Flag", item.Flag, "is not available")
			resultLock.Lock()
			*not_available += 1
			resultLock.Unlock()
		}
	} else {
		fmt.Println("Unknown message received for flag ", item.Flag, ": ", item.Message)
	}
}
