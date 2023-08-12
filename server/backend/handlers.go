package main

import (
	"encoding/json"
	"fmt"
	"log"
	"myflagsubmitter/common"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

func verifyAuthentication(r *http.Request) (string, bool, bool) {
	authToken := r.Header.Get("X-Auth-Token")

	if authToken == API_TOKEN {
		//fmt.Println("Authorization successful!")
		user, ok := getAuthenticatedUsername(r)
		if !ok {
			return "", true, false
		} else {
			return user, true, true
		}
	} else {
		//fmt.Println("Authorization not successful!")
		user, ok := getAuthenticatedUsername(r)
		if !ok {
			return "", false, false
		} else {
			return user, false, true
		}
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")
		// Authenticate user by checking credentials from the database
		if WEB_PASSWORD != password {
			http.Error(w, "Invalid password", http.StatusUnauthorized)
			return
		}

		// Set the session cookie upon successful login
		setAuthenticatedUsername(r, w, username)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Login successful"})
		return
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	clearAuthenticatedUsername(r, w)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Login successful"})
}

func getConfigHandler(w http.ResponseWriter, r *http.Request) {
	_, okToken, _ := verifyAuthentication(r)
	if okToken {
		config := common.FlagSubmitterConfig{
			FlagFormat:    FLAG_FORMAT,
			RoundDuration: ROUND_DURATION,
			Teams:         TEAMS,
			NopTeam:       NOP_TEAM,
			FlagidUrl:     FLAGID_URL,
		}

		jsonData, err := json.Marshal(config)
		if err != nil {
			fmt.Println("Error converting to json: ", err)
			//come mi comporto se non riesco a mandare il json?
		}

		//return the data in json format
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	} else {
		http.Error(w, "You have to login to use this feature", http.StatusUnauthorized)
		return
	}
}

func uploadFlagsHandler(w http.ResponseWriter, r *http.Request) {
	_, okToken, _ := verifyAuthentication(r)
	if okToken {
		//fmt.Println("Incoming Request of flags upload")
		var data common.UploadFlagRequestBody
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			fmt.Println("Error converting from json: ", err)
			return
		}
		// current_app.logger.debug(f"{len(data.get('flags'))} flags received from user {username}")
		var rows []common.Flag
		for _, item := range data.Flags {
			flag := common.Flag{
				Flag:           item.Flag,
				Username:       data.Username,
				ExploitName:    item.ExploitName,
				TeamIp:         item.TeamIp,
				Time:           item.Time,
				Status:         DB_NSUB,
				ServerResponse: DB_NSUB,
			}
			rows = append(rows, flag)
		}

		attempts := 20
		// Try multiple times before failing
		for i := 0; i < attempts; i++ {
			query := "INSERT OR IGNORE INTO flags (flag, username, exploit_name, team_ip, time, status, server_response) VALUES "
			for i, row := range rows {
				query += " (" + "\"" + row.Flag + "\",\"" + row.Username + "\",\"" + row.ExploitName + "\",\"" + row.TeamIp + "\",\"" + row.Time + "\",\"" + DB_NSUB + "\", \"" + DB_NSUB + "\") "
				if i != len(rows)-1 {
					query += ","
				}
			}
			//fmt.Println("Performing query \"" + query + "\" to update flags into database")
			_, err := db.Exec(query)
			if err != nil {
				if i == attempts-1 {
					fmt.Println("Error in uploading flags on the database: ", err)
					time.Sleep(5 * time.Second)
				}
				time.Sleep(1 * time.Second)
			} else {
				break
			}

		}
		//write the updates on the broadcast channel of the clients connected to the webapp
		go updateNewFlags(rows)
		//fmt.Println("Flags", rows, "correclty inserted into database")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Data received"))
	} else {
		http.Error(w, "You have to login first to use this feature", http.StatusUnauthorized)
		return
	}
}

func getFlagsHandler(w http.ResponseWriter, r *http.Request) {
	user, okToken, okUser := verifyAuthentication(r)
	if okToken && okUser {
		var result []common.Flag
		query := "SELECT flag, username, exploit_name, team_ip, time, status, COALESCE(server_response, 'NOT_SUBMITTED') as server_response FROM flags where username = ?"

		rows, err := db.Query(query, user)
		if err != nil {
			fmt.Println("Error while getting flags from the database: ", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var flag common.Flag
			err := rows.Scan(&flag.Flag, &flag.Username, &flag.ExploitName, &flag.TeamIp, &flag.Time, &flag.Status, &flag.ServerResponse)
			if err != nil {
				fmt.Println("Error scanning row while reading flags on the database: ", err)
				return
			}
			result = append(result, flag)
		}
		//return the result in json format
		jsonData, err := json.Marshal(result)
		if err != nil {
			fmt.Println("Error converting to json: ", err)
			//come mi comporto se non riesco a mandare il json?
		}

		//return the data in json format
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	} else {
		http.Error(w, "You have to login first to use this feature", http.StatusUnauthorized)
		return
	}
}
func updateFlagsHandler(w http.ResponseWriter, r *http.Request) {

	user, _, okUser := verifyAuthentication(r)
	if !okUser {
		http.Error(w, "You have to login first to use this feature", http.StatusUnauthorized)
		return
	}

	lastMins, err := strconv.Atoi(r.FormValue("lastMinutes"))
	if err != nil {
		lastMins = 30
	}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading websocket connection:", err)
		return
	}

	clients = append(clients, WebSocketClient{connection: conn, username: user, lastMinutes: lastMins})
	fmt.Println("Client connected", conn.RemoteAddr().String())

	max_attempts := 20
	for attempts := max_attempts; attempts > 0; attempts-- {
		flags, err := get_flags_before(lastMins)
		if err != nil {
			if attempts == max_attempts-1 {
				fmt.Println("Error in updating flags to client:", err)
			}
			time.Sleep(1 * time.Second)
		} else {
			updateClient(nil, conn, flags)
			break
		}
	}
}
