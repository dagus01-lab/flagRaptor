package main

import (
	"database/sql"
	"fmt"
	"myflagsubmitter/common"
	"strings"
	"time"
)

func submission_loop(db *sql.DB) {
	//logica per scegliere il submitter protocol giusto
	submitterFormat, submitter := GetAppSubmitter()
	if submitter == nil {
		fmt.Println("Invalid format for flag submission")
		return
	}

	time.Sleep(5 * time.Second)
	fmt.Println("Starting submission loop...")
	query := "SELECT flag FROM flags WHERE time > ? AND status = ?" //ORDER BY time DESC"
	for {
		s_time := time.Now()
		expiration_time := time.Now().Add(-time.Duration(FLAG_ALIVE * int(time.Second))).Format("2006-01-02 15:04:05")
		// query the flags which are not expired yet and were not submitted
		rows, err := db.Query(query, expiration_time, DB_NSUB)
		if err != nil {
			fmt.Println("Error executing query: ", err)
			time.Sleep(5 * time.Second)
			continue
		}
		defer rows.Close()

		var flags []string

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

		if len(flags) == 0 {
			time.Sleep(1 * time.Second)
			continue
		}
		//fmt.Println("Flags on the DB: ", flags)

		// flags submission
		i := 0
		max_sub := 0
		if SUB_LIMIT > len(flags) {
			max_sub = len(flags)
		} else {
			max_sub = SUB_LIMIT
		}

		for i < max_sub {
			fmt.Println("Submitting flags to checker server...")
			result := submitter(flags)
			if result == nil {
				fmt.Println("Error in flag checker server response!")
				time.Sleep(time.Duration(SUB_INTERVAL) * time.Second)
			}
			//fmt.Println("Flag check response:", result)
			accepted := 0
			old := 0
			nop := 0
			yours := 0
			invalid := 0
			not_available := 0
			update_flag_query := "UPDATE flags SET status = ?, server_response = ? WHERE flag = ?"

			for _, item := range result {
				if strings.Contains(strings.ToLower(submitterFormat.SUB_INVALID), strings.ToLower(item.Message)) {
					_, err := db.Exec(update_flag_query, DB_SUB, DB_ERR, item.Flag)
					if err != nil {
						fmt.Println("Error in updating flags: ", err)
					} else {
						//fmt.Println("Flag", item.Flag, "invalid")
						invalid += 1
					}

				} else if strings.Contains(strings.ToLower(submitterFormat.SUB_YOUR_OWN), strings.ToLower(item.Message)) {
					_, err := db.Exec(update_flag_query, DB_SUB, DB_ERR, item.Flag)
					if err != nil {
						fmt.Println("Error in updating flags: ", err)
					} else {
						//fmt.Println("Flag", item.Flag, "yours")
						yours += 1
					}
				} else if strings.Contains(strings.ToLower(submitterFormat.SUB_NOP), strings.ToLower(item.Message)) {
					_, err := db.Exec(update_flag_query, DB_SUB, DB_ERR, item.Flag)
					if err != nil {
						fmt.Println("Error in updating flags: ", err)
					} else {
						//fmt.Println("Flag", item.Flag, "of nop team")
						nop += 1
					}
				} else if strings.Contains(strings.ToLower(submitterFormat.SUB_OLD), strings.ToLower(item.Message)) {
					_, err := db.Exec(update_flag_query, DB_SUB, DB_EXP, item.Flag)
					if err != nil {
						fmt.Println("Error in updating flags: ", err)
					} else {
						//fmt.Println("Flag", item.Flag, "old")
						old += 1
					}
				} else if strings.Contains(strings.ToLower(submitterFormat.SUB_STOLEN), strings.ToLower(item.Message)) ||
					strings.Contains(strings.ToLower(submitterFormat.SUB_ACCEPTED), strings.ToLower(item.Message)) {
					_, err := db.Exec(update_flag_query, DB_SUB, DB_SUCC, item.Flag)
					if err != nil {
						fmt.Println("Error in updating flags: ", err)
					} else {
						//fmt.Println("Flag", item.Flag, "accepted")
						accepted += 1
					}
				} else if strings.Contains(strings.ToLower(submitterFormat.SUB_NOT_AVAILABLE), strings.ToLower(item.Message)) {
					_, err := db.Exec(update_flag_query, DB_SUB, DB_SUCC, item.Flag)
					if err != nil {
						fmt.Println("Error in updating flags: ", err)
					} else {
						//fmt.Println("Flag", item.Flag, "is not available")
						not_available += 1
					}
				} else {
					fmt.Println("Unknown message received for flag ", item.Flag, ": ", item.Message)
				}
				i += 1

				//write result to the command line
			}
			fmt.Println("Submitted ", len(flags), " flags: ", accepted, " accepted,", old, " old,", nop, " nop,", yours, "yours, ", not_available, "not available")

			var expired_flags []common.Flag
			expired_flags_query := "SELECT flag, username, exploit_name, team_ip, time, server_response FROM flags WHERE server_response = ? AND time <= ?"
			rows, err := db.Query(expired_flags_query, DB_EXP, expiration_time)
			if err != nil {
				fmt.Println("Error executing query: ", err)
				//continue
			}
			defer rows.Close()

			// save all the received flags into a list
			for rows.Next() {
				var flag common.Flag
				err := rows.Scan(&flag.Flag, &flag.Username, &flag.ExploitName, &flag.TeamIp, &flag.Time, &flag.ServerResponse)
				if err != nil {
					fmt.Println("Error scanning row: ", err)
				}
				flag.Status = DB_EXP
				expired_flags = append(expired_flags, flag)
				//fmt.Println("Received flag:", flag)
			}

			//update old flags on database
			update_old_flags_query := "UPDATE flags SET status = ? WHERE time <= ?"
			_, err = db.Exec(update_old_flags_query, DB_EXP, expiration_time)
			if err != nil {
				fmt.Println("Error in updating old flags: ", err)
			}

			//write the updates to all the clients connected to the webapp
			updated_flags, err := find_flags_by_names(flags)
			if err != nil {
				fmt.Println("Error in updating flags to clients: ", err)
			} else {
				updated_flags = append(updated_flags, expired_flags...)
				updateNewFlags(updated_flags)
			}

			duration := time.Now().Sub(s_time)
			if duration < time.Duration(SUB_INTERVAL)*time.Second {
				//fmt.Println("Sleeping for ", time.Duration(SUB_INTERVAL)*time.Second-duration)
				time.Sleep(time.Duration(SUB_INTERVAL)*time.Second - duration)
			}
		}
	}
}
