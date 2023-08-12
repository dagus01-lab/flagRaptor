# flagWarehouse
Flag submission system for Attack/Defense CTFs.

* [Server](#server)
    * [Installation](#installation)
    * [Configuration](#configuration)
    * [Usage](#usage)
* [Client](#client)

## Server
The server is a go web application that uses SQLite as its database engine: It receives flags stolen by the team members and sends them
periodically to the checksystem provided by the game organizators. It also allows users to retrieve some useful data from the database.

### Installation
First of all, you should clone the repository
```
git clone https://github.com/dagus01-lab/flagSubmitter.git
```
Then, you should install all the required dependencies for the go web application
```
go mod download
```
Finally, you should install all the required dependencies to run the frontend 
```
cd flagSubmitter/server/frontend
npm install
npx webpack --config webpack.config.js
```
### Configuration
Edit the parameters in [utils.go](server/backend/utils.go)

- `WEB_PASSWORD`: the password to access the web interface
- `API_TOKEN`: token used for communication between clients and server
- `TEAM_TOKEN`: token used to submit flags
- `FLAG_FORMAT`: string containing the regex format of the flags
- `YOUR_TEAM`: the ip address of your team
- `TEAMS`: the ip addresses of the teams in the competition
- `ROUND_DURATION`: the duration of a round (or *tick*) in seconds
- `FLAG_ALIVE`: the number of seconds a flag can be considered valid
- `SUB_PROTOCOL`: gameserver submission protocol. Valid values are `dummy` (will only print flags on stdout) and `ccit`
- `SUB_LIMIT`: number of flags that can be sent to the organizers' server each `SUB_INTERVAL`
- `SUB_INTERVAL`: interval in seconds for the submission; if the submission round takes more than the number of seconds
                  specified, the background submission loop will not sleep
- `SUB_URL`: the url used for the verification of the flags

### Usage
The backend needs to be compiled before execution
```
cd backend
go build
./backend
```
The web interface can be accessed on port 5000. To log in, use any username and the password you set.

In case of wrong credentials users will be notified and asked to enter the correct username and password


## Client
```
cd client
go build
./client -s SERVER_URL -u USERNAME -t TOKEN -d ./exploits_dir/ -n THREADS
```
The client is a go program that runs all the programs in the directory specified by the user.
Firstly the clien fetches the configuration from the server. Then it reads the standard output of every
script, extracts the flags using the regex provided in the server configurations, and sends them to the 
server. 

### REMEMBER

- Attack scripts must have coherent names
- Attack scripts must have shebang at the beginning
- Attack scripts must be executable
- Set
    - `TEAM_TOKEN`
    - `YOUR_TEAM`
    - `FLAG_FORMAT`
    - `SUB_URL`
    - Strings of verification messages if necessary
