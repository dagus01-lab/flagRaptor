package main

import (
	"flagRaptor/common"
	"log"
	"sync"
	"time"
)

func getExpiredFlags(expiration_time string) ([]common.Flag, error) {
	expired_flags := make([]common.Flag, 0)
	expired_flags_query := "SELECT flag, username, exploit_name, team_ip, time, server_response FROM flags WHERE status = ? AND time <= ?"
	dbLock.Lock()
	rows, err := db.Query(expired_flags_query, cfg.SubmissionConf.DBEXP, expiration_time)
	dbLock.Unlock()
	if err != nil {
		return expired_flags, err
	}
	defer rows.Close()

	// save all the received flags into a list
	for rows.Next() {
		var flag common.Flag
		err := rows.Scan(&flag.Flag, &flag.Username, &flag.ExploitName, &flag.TeamIp, &flag.Time, &flag.ServerResponse)
		if err != nil {
			return expired_flags, err
		}
		flag.Status = cfg.SubmissionConf.DBEXP
		expired_flags = append(expired_flags, flag)
		//log.Println("Received flag:", flag)
	}
	return expired_flags, nil
}

var cfg Config

func submission_loop(config *Config) {
	cfg = *config
	//logica per scegliere il submitter protocol giusto
	submitter := GetAppSubmitter()
	if submitter == nil {
		log.Println("Invalid format for flag submission")
		return
	}

	time.Sleep(5 * time.Second)
	log.Println("Starting submission loop...")
	for {
		s_time := time.Now()
		expiration_time := time.Now().Add(-cfg.GameConf.FlagAlive).Format("2006-01-02 15:04:05")

		flags, err := getFlagsToCheck(expiration_time)
		if err != nil {
			log.Println("Error executing query: ", err)
			time.Sleep(1 * time.Second)
			continue
		}
		if len(flags) == 0 {
			time.Sleep(1 * time.Second)
			continue
		}
		// flags submission
		i := 0
		max_sub := 0
		if cfg.SubmissionConf.SubLimit > len(flags) {
			max_sub = len(flags)
		} else {
			max_sub = cfg.SubmissionConf.SubLimit
		}

		for i < max_sub {
			log.Println("Submitting flags to checker server...")
			result := submitter.Submit(flags)
			if result == nil {
				log.Println("Error in flag checker server response!")
				time.Sleep(cfg.SubmissionConf.SubInterval)
			}

			//log.Println("Flag check response:", result)
			accepted := 0
			old := 0
			nop := 0
			yours := 0
			invalid := 0
			not_available := 0
			var resultLock sync.Mutex

			var wg sync.WaitGroup
			for _, item := range result {
				wg.Add(1)
				go updateUploadedFlagsToDB(&wg, &accepted, &old, &nop, &yours, &invalid, &not_available, item, submitter.Format, &resultLock)
				i += 1
			}
			wg.Wait()
			log.Println("Submitted ", len(flags), " flags: ", accepted, " accepted,", old, " old,", nop, " nop,", yours, "yours, ", not_available, "not available")

			//update old flags on database
			err = setOldFlagsAsExpired(expiration_time)
			if err != nil {
				log.Println("Error in updating old flags: ", err)
			}

			//retrieve all expired flags from the database
			expired_flags, err := getExpiredFlags(expiration_time)
			if err != nil {
				log.Println("Error on getting expired flags: ", err)
			}

			//write the updates to all the clients connected to the webapp
			updated_flags, err := findFlagsByNames(flags)
			if err != nil {
				log.Println("Error in updating flags to clients: ", err)
			} else {
				updated_flags = append(updated_flags, expired_flags...)
				go updateNewFlags(updated_flags)
			}

			duration := time.Since(s_time)

			if duration < cfg.SubmissionConf.SubInterval {
				time.Sleep(cfg.SubmissionConf.SubInterval - duration)
			}
		}
	}
}
