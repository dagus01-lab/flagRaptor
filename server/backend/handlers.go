package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"myflagsubmitter/common"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func verifyAuthentication(r *http.Request) (string, bool, bool) {
	authToken := r.Header.Get("X-Auth-Token")

	if authToken == cfg.ServerConf.APIToken {
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
		if cfg.AuthConf.AuthEnable {
			rows, err := db.Query("SELECT * FROM users WHERE username=", username, " AND PASSWORD=", password)
			if err != nil {
				http.Error(w, "Internal error", http.StatusInternalServerError)
				return
			}
			if !rows.Next() {
				http.Error(w, "Invalid password", http.StatusUnauthorized)
				return
			}

		} else {
			if cfg.AuthConf.WebPassword != password {
				http.Error(w, "Invalid password", http.StatusUnauthorized)
				return
			}
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
			FlagFormat:    cfg.GameConf.FlagFormat,
			RoundDuration: cfg.GameConf.RoundDuration,
			Teams:         cfg.GameConf.Teams,
			NopTeam:       cfg.GameConf.NopTeam,
			FlagidUrl:     cfg.GameConf.FlagIDUrl,
			ClientPort:    cfg.ServerConf.ClientPort,
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

		var rows []common.Flag
		for _, item := range data.Flags {
			flag := common.Flag{
				Flag:           item.Flag,
				Username:       data.Username,
				ExploitName:    item.ExploitName,
				TeamIp:         item.TeamIp,
				Time:           item.Time,
				Status:         cfg.SubmissionConf.DBNSUB,
				ServerResponse: cfg.SubmissionConf.DBNSUB,
			}

			rows = append(rows, flag)
		}
		if len(rows) == 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Data received"))
			return
		}
		var exploits *map[string]bool
		//update the scriptRunners status
		scriptRunnersLock.Lock()
		userFound := false
		remoteAddr := strings.Split(r.RemoteAddr, ":")
		clientAddress := strings.Join(remoteAddr[:len(remoteAddr)-1], ":") + ":" + strconv.Itoa(cfg.ServerConf.ClientPort)
		for _, scriptRunner := range scriptRunners {
			if scriptRunner.user == rows[0].Username {
				addressFound := false
				for _, address := range scriptRunner.addresses {
					if address == clientAddress {
						addressFound = true
						break
					}
				}
				if !addressFound {
					scriptRunner.addresses = append(scriptRunner.addresses, clientAddress)
					//for every new client process trying to update some flags the server orders it to stop the script that have been stopped from the web interface
					for exploit, isRunning := range *exploits {
						if !isRunning {
							data := []byte(`name=` + exploit)
							resp, err := http.Post("http://"+clientAddress+"/stop", "application/x-www-form-urlencoded", bytes.NewBuffer(data))
							if err != nil {
								fmt.Println("Error on stopping script", exploit, ", error:", err)
								return
							}
							defer resp.Body.Close()
						}
					}
				}
				userFound = true
				exploits = &scriptRunner.exploits
				break
			}
		}
		if !userFound {
			scriptRunners = append(scriptRunners, ScriptRunner{user: rows[0].Username, addresses: append(make([]string, 0), clientAddress), exploits: make(map[string]bool, 0)})
			exploits = &scriptRunners[len(scriptRunners)-1].exploits

			//for every new client process trying to update some flags the server orders it to stop the script that have been stopped from the web interface
			for exploit, isRunning := range *exploits {
				if !isRunning {
					data := []byte(`name=` + exploit)
					resp, err := http.Post("http://"+clientAddress+"/stop", "application/x-www-form-urlencoded", bytes.NewBuffer(data))
					if err != nil {
						fmt.Println("Error on stopping script", exploit, ", error:", err)
						return
					}
					defer resp.Body.Close()
				}
			}
		}
		scriptRunnersLock.Unlock()

		//flag upload
		attempts := 5
		// Try multiple times before failing
		for i := 0; i < attempts; i++ {
			query := "INSERT OR IGNORE INTO flags (flag, username, exploit_name, team_ip, time, status, server_response) VALUES "
			for i, row := range rows {

				//update the exploits status inside the scriptRunners status, so that new scripts are added as executing scripts
				isRunning, found := (*exploits)[row.ExploitName]
				if !found {
					(*exploits)[row.ExploitName] = true
				} else if !isRunning {
					//if a client tries to upload a flag with a script that has been blocked, the upload is blocked and the client is ordered to stop the script execution
					data := []byte(`name=` + row.ExploitName)
					resp, err := http.Post("http://"+clientAddress+"/stop", "application/x-www-form-urlencoded", bytes.NewBuffer(data))
					if err != nil {
						fmt.Println("Error on stopping script", row.ExploitName, ", error:", err)
						return
					}
					defer resp.Body.Close()
					continue
				}

				query += " (" + "\"" + row.Flag + "\",\"" + row.Username + "\",\"" + row.ExploitName + "\",\"" + row.TeamIp + "\",\"" + row.Time + "\",\"" + cfg.SubmissionConf.DBNSUB + "\", \"" + cfg.SubmissionConf.DBNSUB + "\") "
				if i != len(rows)-1 {
					query += ","
				}

			}
			//fmt.Println("Performing query \"" + query + "\" to update flags into database")
			dbLock.Lock()
			_, err := db.Exec(query)
			dbLock.Unlock()
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

		dbLock.Lock()
		rows, err := db.Query(query, user)
		dbLock.Unlock()
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
		fmt.Println("Error upgrading websocket connection:", err)
		return
	}

	webSocketClientsLock.Lock()
	clients = append(clients, WebSocketClient{connection: conn, username: user, lastMinutes: lastMins})
	webSocketClientsLock.Unlock()
	fmt.Println("Client connected", conn.RemoteAddr().String())

	max_attempts := 5
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
func restartExploitHandler(w http.ResponseWriter, r *http.Request) {
	user, okToken, okUser := verifyAuthentication(r)
	if okToken && okUser {
		//retrieve the name of the exploit to restart
		exploit := r.FormValue("exploit")
		//retrieve the addresses associated to the user to notify
		found := false
		var clients []string
		for _, scriptRunner := range scriptRunners {
			if scriptRunner.user == user {
				clients = scriptRunner.addresses
				isExecuting, found_exec_status := scriptRunner.exploits[exploit]
				if found_exec_status && !isExecuting {
					scriptRunner.exploits[exploit] = true
					removeStoppedExploit(user, exploit)
					found = true
				} else {
					found = false
				}
				break
			}
		}
		if !found {
			http.Error(w, "Nonexisting user", http.StatusBadRequest)
			return
		} else {
			for _, client := range clients {
				data := []byte(`name=` + exploit)
				resp, err := http.Post("http://"+client+"/start", "application/x-www-form-urlencoded", bytes.NewBuffer(data))
				if err != nil {
					fmt.Println("Error:", err)
					return
				}
				defer resp.Body.Close()
			}
		}
	} else {
		http.Error(w, "You have to login first to use this feature", http.StatusUnauthorized)
		return
	}
}
func stopExploitHandler(w http.ResponseWriter, r *http.Request) {
	user, okToken, okUser := verifyAuthentication(r)
	if okToken && okUser {
		//retrieve the name of the exploit to restart
		exploit := r.FormValue("exploit")
		//retrieve the addresses associated to the user to notify
		found := false
		var clients []string
		for _, scriptRunner := range scriptRunners {
			if scriptRunner.user == user {
				clients = scriptRunner.addresses
				isExecuting, found_exec_status := scriptRunner.exploits[exploit]
				if found_exec_status && isExecuting {
					scriptRunner.exploits[exploit] = false
					addStoppedExploit(user, exploit)
					found = true
				} else {
					found = false
				}
				break
			}
		}
		fmt.Println("User", user, "wants to stop", exploit, ".Found", clients)
		if !found {
			http.Error(w, "Nonexisting user", http.StatusBadRequest)
			return
		} else {
			for _, client := range clients {
				data := []byte(`name=` + exploit)
				resp, err := http.Post("http://"+client+"/stop", "application/x-www-form-urlencoded", bytes.NewBuffer(data))
				if err != nil {
					fmt.Println("Error on stopping script", exploit, ", error:", err)
					return
				}
				defer resp.Body.Close()
			}
		}
	} else {
		http.Error(w, "You have to login first to use this feature", http.StatusUnauthorized)
		return
	}
}

func getStoppedExploitsHandler(w http.ResponseWriter, r *http.Request) {
	user, okToken, okUser := verifyAuthentication(r)
	if okToken && okUser {
		result := make([]string, 0)
		for _, scriptRunner := range scriptRunners {
			if scriptRunner.user == user {
				for exploit, isRunning := range scriptRunner.exploits {
					if !isRunning {
						result = append(result, exploit)
					}
				}
				//return the result in json format
				w.Header().Set("Content-Type", "application/json")
				err := json.NewEncoder(w).Encode(result)
				if err != nil {
					fmt.Println("Error converting to json: ", err)
				}
				return
			}
		}
		http.Error(w, "You are not running any script", http.StatusBadRequest)
	} else {
		http.Error(w, "You have to login first to use this feature", http.StatusUnauthorized)
		return
	}
}
