# flagSubmitter
Flag submission system for Attack/Defense CTFs.

* [Server](#server)
    * [Installation](#installation)
    * [Configuration](#configuration)
    * [Usage](#usage)
* [Client](#client)
* [Screenshots](#screenshots)

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
It is possible to give the flagSubmitter a .yaml file for configuration, and if it is not provided, 
the file [config.yaml](server/backend/config.yaml) is chosen. These are the main parameters that can be specified:

- `webPassword`: the password to access the web interface
- `apiToken`: token used for communication between clients and server
- `flagFormat`: string containing the regex format of the flags
- `teamIP`: the ip address of your team
- `teamFormat`: the ip addresses of the teams in the competition, expressed with a wildcard
- `roundDuration`: the duration of a round (or *tick*) in seconds
- `subProtocol`: gameserver submission protocol. Valid values are `dummy` (will only print flags on stdout) and `ccit`
- `subLimit`: number of flags that can be sent to the organizers' server each `subInterval`
- `subInterval`: interval in seconds for the submission; if the submission round takes more than the number of seconds
                  specified, the background submission loop will not sleep
- `subUrl`: the url used for the verification of the flags
- `clientPort`: the port of the clients, through which the server can order them to stop/restart some exploits

### Usage
The backend needs to be compiled before execution
```
cd backend -f CONF_FILE
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
Firstly the client fetches the configuration from the server. Then it reads the standard output of every
script, extracts the flags using the regex provided in the server configurations, and sends them to the 
server. 
Moreover the client listens on a specific port, in order to allow the server to stop/restart some of the
exploits that are in the specified directory

## Screenshots
![Connection Content](https://github.com/dagus01-lab/flagSubmitter/tree/main/server/frontend/screenshots/home.png)
![Connection Content](https://github.com/dagus01-lab/flagSubmitter/tree/main/server/frontend/screenshots/explore.png)
![Connection Content](https://github.com/dagus01-lab/flagSubmitter/tree/main/server/frontend/screenshots/submit.png)

### REMEMBER

- Attack scripts must have coherent names
- Attack scripts must have shebang at the beginning
- Attack scripts must be executable
- Set
    - `team`
    - `flagFormat`
    - `subUrl`
    - Strings of verification messages if necessary
