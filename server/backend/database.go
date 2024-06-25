package main

import (
	"database/sql"
	"flagRaptor/common"
	"log"
	"strings"
	"sync"
	"time"
)

func initDatabase(databaseFile string) *sql.DB {
	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func createTables() error {
	err := createFlagsTable()
	if err != nil {
		return err
	}
	err = createUsersTable()
	if err != nil {
		return err
	}
	err = createStoppedExploitsTable()
	if err != nil {
		return err
	}
	return nil
}

func createFlagsTable() error {
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

func createUsersTable() error {
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS users
		(
			username        TEXT PRIMARY KEY,
			password        TEXT NOT NULL
		);
	`
	_, err := db.Exec(createTableSQL)
	for _, user := range cfg.AuthConf.Users {
		db.Exec("INSERT INTO users VALUES (", user.Username, ",", user.Password, ")")
	}
	return err
}

func createStoppedExploitsTable() error {
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS stopped_exploits
		(
			username TEXT NOT NULL,
			exploit_name    TEXT NOT NULL,
			PRIMARY KEY (exploit_name, username)
		);
	`
	_, err := db.Exec(createTableSQL)
	return err
}

func findFlagsByNames(flags []string) ([]common.Flag, error) {
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
		log.Println("Error executing query ", query, ": ", err)
		return nil, err
	}
	defer rows.Close()

	// save all the received flags into a list
	for rows.Next() {
		var flag common.Flag
		err := rows.Scan(&flag.Flag, &flag.Username, &flag.ExploitName, &flag.TeamIp, &flag.Time, &flag.Status, &flag.ServerResponse)
		if err != nil {
			log.Println("Error scanning row: ", err)
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
		log.Println("Error executing query: ", err)
		return nil, err
	}
	defer rows.Close()

	// save all the received flags into a list
	for rows.Next() {
		var flag common.Flag
		err := rows.Scan(&flag.Flag, &flag.Username, &flag.ExploitName, &flag.TeamIp, &flag.Time, &flag.Status, &flag.ServerResponse)
		if err != nil {
			log.Println("Error scanning row: ", err)
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
		log.Println("Error executing query: ", err)
		return nil, err
	}
	defer rows.Close()

	// save all the received flags into a list
	for rows.Next() {
		var flag common.Flag
		err := rows.Scan(&flag.Flag, &flag.Username, &flag.ExploitName, &flag.TeamIp, &flag.Time, &flag.Status, &flag.ServerResponse)
		if err != nil {
			log.Println("Error scanning row: ", err)
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
	rows, err := db.Query(query, expiration_time, cfg.SubmissionConf.DBNSUB)
	dbLock.Unlock()
	if err != nil {
		return flags, err
	}
	defer rows.Close()

	// save all the received flags into a list
	for rows.Next() && len(flags) < cfg.SubmissionConf.SubPayloadSize {
		var flag string
		err := rows.Scan(&flag)
		if err != nil {
			log.Println("Error scanning row: ", err)
			break
		}
		flags = append(flags, flag)
	}
	return flags, nil
}

func setOldFlagsAsExpired(expiration_time string) error {
	update_old_flags_query := "UPDATE flags SET status = ? WHERE time <= ?"
	dbLock.Lock()
	_, err := db.Exec(update_old_flags_query, cfg.SubmissionConf.DBEXP, expiration_time)
	dbLock.Unlock()
	return err
}

func updateUploadedFlagsToDB(wg *sync.WaitGroup, accepted *int, old *int, nop *int, yours *int, invalid *int, not_available *int, item ResponseItem, submitterFormat SubmitterFormat, resultLock *sync.Mutex) {
	defer wg.Done()
	query := "UPDATE flags SET status = ?, server_response = ? WHERE flag = ?"
	if strings.Contains(strings.ToLower(submitterFormat.SUB_INVALID), strings.ToLower(item.Message)) {
		dbLock.Lock()
		_, err := db.Exec(query, cfg.SubmissionConf.DBSUB, cfg.SubmissionConf.DBERR, item.Flag)
		dbLock.Unlock()
		if err != nil {
			log.Println("Error in updating flags: ", err)
		} else {
			//log.Println("Flag", item.Flag, "invalid")
			resultLock.Lock()
			*invalid += 1
			resultLock.Unlock()
		}

	} else if strings.Contains(strings.ToLower(submitterFormat.SUB_YOUR_OWN), strings.ToLower(item.Message)) {
		dbLock.Lock()
		_, err := db.Exec(query, cfg.SubmissionConf.DBSUB, cfg.SubmissionConf.DBERR, item.Flag)
		dbLock.Unlock()
		if err != nil {
			log.Println("Error in updating flags: ", err)
		} else {
			//log.Println("Flag", item.Flag, "yours")
			resultLock.Lock()
			*yours += 1
			resultLock.Unlock()
		}
	} else if strings.Contains(strings.ToLower(submitterFormat.SUB_NOP), strings.ToLower(item.Message)) {
		dbLock.Lock()
		_, err := db.Exec(query, cfg.SubmissionConf.DBSUB, cfg.SubmissionConf.DBERR, item.Flag)
		dbLock.Unlock()
		if err != nil {
			log.Println("Error in updating flags: ", err)
		} else {
			//log.Println("Flag", item.Flag, "of nop team")
			resultLock.Lock()
			*nop += 1
			resultLock.Unlock()
		}
	} else if strings.Contains(strings.ToLower(submitterFormat.SUB_OLD), strings.ToLower(item.Message)) {
		dbLock.Lock()
		_, err := db.Exec(query, cfg.SubmissionConf.DBSUB, cfg.SubmissionConf.DBEXP, item.Flag)
		dbLock.Unlock()
		if err != nil {
			log.Println("Error in updating flags: ", err)
		} else {
			//log.Println("Flag", item.Flag, "old")
			resultLock.Lock()
			*old += 1
			resultLock.Unlock()
		}
	} else if strings.Contains(strings.ToLower(submitterFormat.SUB_STOLEN), strings.ToLower(item.Message)) ||
		strings.Contains(strings.ToLower(submitterFormat.SUB_ACCEPTED), strings.ToLower(item.Message)) {
		dbLock.Lock()
		_, err := db.Exec(query, cfg.SubmissionConf.DBSUB, cfg.SubmissionConf.DBSUCC, item.Flag)
		dbLock.Unlock()
		if err != nil {
			log.Println("Error in updating flags: ", err)
		} else {
			//log.Println("Flag", item.Flag, "accepted")
			resultLock.Lock()
			*accepted += 1
			resultLock.Unlock()
		}
	} else if strings.Contains(strings.ToLower(submitterFormat.SUB_NOT_AVAILABLE), strings.ToLower(item.Message)) {
		dbLock.Lock()
		_, err := db.Exec(query, cfg.SubmissionConf.DBSUB, cfg.SubmissionConf.DBSUCC, item.Flag)
		dbLock.Unlock()
		if err != nil {
			log.Println("Error in updating flags: ", err)
		} else {
			//log.Println("Flag", item.Flag, "is not available")
			resultLock.Lock()
			*not_available += 1
			resultLock.Unlock()
		}
	} else {
		log.Println("Unknown message received for flag ", item.Flag, ": ", item.Message)
	}
}

func get_stopped_exploits() ([]ScriptRunner, error) {
	scriptRunners := make([]ScriptRunner, 0)
	query := "SELECT * FROM stopped_exploits"
	dbLock.Lock()
	rows, err := db.Query(query)
	dbLock.Unlock()
	if err != nil {
		return scriptRunners, err
	}
	defer rows.Close()

	// save all the received flags into a list
	for rows.Next() {
		var exploit struct {
			Username     string
			Exploit_name string
		}
		err := rows.Scan(&exploit)
		if err != nil {
			log.Println("Error scanning row: ", err)
			break
		}
		found := false
		for _, scriptRunner := range scriptRunners {
			if scriptRunner.user == exploit.Username {
				scriptRunner.exploits[exploit.Exploit_name] = false
				found = true
				break
			}
		}
		if !found {
			scriptRunner := ScriptRunner{
				user:     exploit.Username,
				exploits: make(map[string]bool, 0),
			}
			scriptRunner.exploits[exploit.Exploit_name] = false
			scriptRunners = append(scriptRunners, scriptRunner)
		}
	}

	return scriptRunners, nil
}

func addStoppedExploit(username string, exploit_name string) error {
	dbLock.Lock()
	_, err := db.Exec("INSERT INTO stopped_exploits VALUES (", username, ",", exploit_name, ")")
	dbLock.Unlock()
	return err
}
func removeStoppedExploit(username string, exploit_name string) error {
	dbLock.Lock()
	_, err := db.Exec("DELETE FROM stopped_exploits WHERE username=", username, " AND exploit_name=", exploit_name)
	dbLock.Unlock()
	return err
}
