package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"myflagsubmitter/common"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	scriptLock         sync.Mutex
	executableScripts  = make(map[string]bool, 0)
	executingProcesses = make(map[string]*exec.Cmd, 0)
)

func isExecutable(file os.FileInfo) bool {
	return file.Mode()&0111 != 0
}
func hasShebang(file string) (bool, error) {
	f, err := os.Open(file)
	if err != nil {
		return false, err
	}
	defer f.Close()

	r := bufio.NewReader(f)
	firstTwoBytes := make([]byte, 2)
	_, err = r.Read(firstTwoBytes)
	if err != nil {
		return false, nil
	}
	return string(firstTwoBytes) == "#!", nil
}

func run_exploit(wg *sync.WaitGroup, script string, team string, round_duration time.Duration, server_url string, token string, flag_format string, user string) {
	defer wg.Done()

	//execute the exploit
	timeout := time.After(round_duration)
	done := make(chan error)

	cmd := exec.Command(script, team)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println("Error creating stdout pipe:", err)
		return
	}

	go func() {
		err := cmd.Start()
		done <- err
	}()
	go func() {
		select {
		case err := <-done:
			if err != nil {
				log.Printf("Script execution failed: %v\n", err)
			} else {
				log.Println("Script ", script, "executed successfully.")
			}
		case <-timeout:
			// Kill the process if it's still running after the timeout.
			err := cmd.Process.Kill()
			if err != nil {
				log.Printf("Failed to kill the process: %v\n", err)
			}
			log.Println("Script execution timed out")
		}
	}()

	if err != nil {
		log.Println("Error launching the exploit:", err)
		return
	}

	scriptLock.Lock()
	executingProcesses[script] = cmd
	scriptLock.Unlock()

	scanner := bufio.NewScanner(stdout)
	request_body := common.UploadFlagRequestBody{
		Username: user,
		Flags:    make([]common.Flag, 0),
	}
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			log.Println("Empty line received from", script)
			break
		} else {

			flags := regexp.MustCompile(flag_format).FindAllString(line, -1)

			if len(flags) == 0 {
				continue
			}
			log.Println(script, "@", team, ": Got", len(flags), "flags with", script, "from", team, ":", flags)
			timestamp := time.Now().Format("2006-01-02 15:04:05")
			for _, flag := range flags {
				request_body.Flags = append(request_body.Flags, common.Flag{
					Flag:        flag,
					ExploitName: script,
					TeamIp:      team,
					Time:        timestamp,
				})
			}
		}
	}

	json_data, err := json.Marshal(request_body)
	if err != nil {
		log.Println("Error in parsing json request:", err)
		return
	}

	req, err := http.NewRequest("POST", server_url+"/upload_flags", bytes.NewBufferString(string(json_data)))
	if err != nil {
		log.Println("Error creating GET request:", err)
		return
	}
	req.Header.Set("X-Auth-Token", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error sending GET request:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Received non-ok status code:", resp.StatusCode)
		return
	}
}

func handleStartProcess(w http.ResponseWriter, r *http.Request) {
	scriptLock.Lock()
	defer scriptLock.Unlock()

	scriptName := r.FormValue("name")
	executableScripts[scriptName] = true

	log.Println("Script" + scriptName + "ready to be started again")
	w.WriteHeader(http.StatusAccepted)
}

func handleStopProcess(w http.ResponseWriter, r *http.Request) {
	scriptLock.Lock()
	defer scriptLock.Unlock()

	scriptName := r.FormValue("name")

	cmd, found := executingProcesses[scriptName]
	if found {
		err := cmd.Process.Kill()
		if err != nil {
			log.Println("Error on killing process associated to", scriptName)
		}
		delete(executingProcesses, scriptName)
	}
	if executableScripts[scriptName] {
		executableScripts[scriptName] = false
		log.Println("Script", scriptName, "stopped")
	}
	w.WriteHeader(http.StatusAccepted)

}

func exploitsControl(serverPort int) {
	http.HandleFunc("/start", handleStartProcess)
	http.HandleFunc("/stop", handleStopProcess)
	http.ListenAndServe(":"+strconv.Itoa(serverPort), nil)
}

func main() {
	//define command line flags
	server_url := flag.String("s", "http://localhost:5000", "server url")
	user := flag.String("u", "", "user")
	token := flag.String("t", "", "token")
	exploit_dir := flag.String("d", "exploits", "exploit directory")
	num_threads := flag.Int("n", 128, "maximum number of threads")
	//parse the command line flags
	flag.Parse()
	//check if the required flags are provided
	if *user == "" || *token == "" {
		flag.PrintDefaults()
		return
	}

	//retrieve configuration from server
	req, err := http.NewRequest("GET", *server_url+"/get_config", nil)
	if err != nil {
		log.Println("Error creating GET request:", err)
		return
	}
	req.Header.Set("X-Auth-Token", *token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error sending GET request:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Received non-ok status code:", resp.StatusCode)
		return
	}

	//parse server configuration
	var server_conf common.FlagSubmitterConfig
	err = json.NewDecoder(resp.Body).Decode(&server_conf)
	if err != nil {
		log.Println("Error decoding response body:", err)
		return
	}

	log.Println("Got this configuration from server: round duration=", server_conf.RoundDuration, ", teams:", server_conf.Teams, ", nop team:", server_conf.NopTeam, ", flag submitter server url:", server_conf.FlagidUrl, ", flag format:", server_conf.FlagFormat, ", client port:", server_conf.ClientPort)

	//start the exploitsControl routine
	go exploitsControl(server_conf.ClientPort)

	//load exploits from exploit directory
	var scripts []string

	err = filepath.Walk(*exploit_dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && isExecutable(info) {
			if has, err := hasShebang(path); err == nil && has {
				executableScripts[path] = true
				scripts = append(scripts, path)
			}
		} else {
			log.Println("Script", path, "is not executable")
		}
		return nil
	})

	for {

		//run every script for each team
		i := 0
		terminate := false
		var wg sync.WaitGroup
		for _, script := range scripts {
			for _, team := range server_conf.Teams {
				log.Println("Started running exploit ", script, "on team", team)
				wg.Add(1)
				go run_exploit(&wg, script, team, server_conf.RoundDuration, *server_url, *token, server_conf.FlagFormat, *user)

				i++
				if i > *num_threads {
					terminate = true
					break
				}
			}
			if terminate {
				break
			}
		}
		wg.Wait()
		time.Sleep(1 * time.Second)

		//reload exploits from the exploit directory
		scripts = make([]string, 0)
		err = filepath.Walk(*exploit_dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && isExecutable(info) {
				if has, err := hasShebang(path); err == nil && has && executableScripts[path] {
					scripts = append(scripts, path)
				}
			} else {
				log.Println("Script", path, "is not executable")
			}
			return nil
		})
	}
}
