import React, {useEffect, useState} from "react";
import BarChartComponent, {optionsTop, optionsCenter} from "./BarChartComponent";
import FlagTableComponent from "./FlagTableComponent";
import FlagSubmissionComponent from "./FlagSubmissionComponent";
import MenuComponent from "./MenuComponent"
import ScriptRunnerControlButton from "./ScriptRunnerControlButton";

const Home = ({handleLogoutFunction, username}) => {
  const [status, setStatus] = useState("home");
  const [flagsData, setFlagsData] = useState([]);
  const [personalExploitsdata, setPersonalExploitsData] = useState({
    labels: [],
    datasets: [],
  });
  const [teamExploitsdata, setTeamExploitsData] = useState({
    labels: [],
    datasets: [],
  });
  const [teamStatusdata, setTeamStatusData] = useState({
    labels: [],
    datasets: [],   
  })
  const [stoppedExploits, setStoppedExploits] = useState([])

  useEffect(() => {

    fetch('/get_stopped_exploits', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Auth-Token': 'token', // Add the custom header "X-Auth-Token" with your token
      },
    })
      .then((response)=>{
        if(!response.ok){
          throw new Error("HTTP error! Status: ${response.status}")
        }
        else{
          return response.json()
        }
      })
      .then((data) => {
        if(typeof data != "undefined"){
          setStoppedExploits(data)
        }
      })
      .catch((error) => {
        console.error('Error getting stopped exploits:', error);
      });

    const socket = new WebSocket('ws://localhost:5000/update_flags');
    // Listen for real-time updates from the server
    socket.onmessage = (event) => {
     handleFlagUpdate(event.data); // Handle the real-time flag update
    };

    return () => {
      socket.close(); // Clean up the WebSocket connection on unmount
    };
  }, []); //the empty array ensures that the function is executed only when the component is loaded

  useEffect(() => {
    updateBarChartData();
  },[flagsData]); //automatically update barchart data when new flags are received

  const updateBarChartData = () => {
    //alert("Flags: "+ JSON.stringify(flagsData))
    // Process the flags data to generate the chart datasets
    const personalExploits = [...new Set(flagsData.filter((flag) => flag.username == username).map((flag) => flag.exploit_name))];
    const teamExploits = [...new Set(flagsData.map((flag) => flag.exploit_name))];
    const teams = [...new Set(flagsData.map((flag) => flag.team_ip))];
    const responses = [...new Set(flagsData.map((flag) => flag.server_response))];
  
    const personalExploitsDatasets = responses.map((server_response) => {
      const data = personalExploits.map((exploit_name) => {
        return flagsData.filter((flag) => flag.exploit_name === exploit_name && flag.server_response === server_response).length;
      });

      return {
        label: server_response,
        data,
        backgroundColor: getColor(server_response),
        borderWidth: 1,
      };
    });
    const teamExploitsDatasets = responses.map((server_response) => {
      const data = teamExploits.map((exploit_name) => {
        return flagsData.filter((flag) => flag.exploit_name === exploit_name && flag.server_response === server_response).length;
      });

      return {
        label: server_response,
        data,
        backgroundColor: getColor(server_response),
        borderWidth: 1,
      };
    });
    const teamstatusdatasets = responses.map((server_response) => {
      const data = teams.map((team_ip) => {
        return flagsData.filter((flag) => flag.team_ip === team_ip && flag.server_response === server_response).length;
      });

      return {
        label: server_response,
        data,
        backgroundColor: getColor(server_response),
        borderWidth: 1,
      };
    });

    setPersonalExploitsData({ labels:personalExploits, datasets : personalExploitsDatasets});
    setTeamExploitsData({ labels: teamExploits, datasets: teamExploitsDatasets});
    setTeamStatusData({labels: teams, datasets:teamstatusdatasets})
  };

  const handleFlagUpdate = (jsonString) => {
    try {
      const jsonObject = JSON.parse(jsonString);
      console.log(jsonObject);

      setFlagsData(prevFlagsData => {
        return [...prevFlagsData,...jsonObject.map(flag => {
          const existingFlag = prevFlagsData.find(f => f.flag === flag.flag);
      
          if (existingFlag) {
            // Update the existing flag
            return { ...existingFlag, server_response: flag.server_response, status: flag.status };
          } else {
            // Add a new flag to the array
            return flag;
          }
        })];
      });
    } catch (error) {
      console.error('Error parsing JSON:', error.message);
    }
    
  };

  const getColor = (response) => {
    if(response == "NOT_SUBMITTED"){
      return '#0000FF'
    }
    else if(response == "SUCCESS"){
      return '#00FF00'
    }
    else if(response == "ERROR"){
      return '#FF0000'
    }
    else if(response == "EXPIRED"){
      return '#FFFF00'
    }
    else
      return '#' + (Math.random().toString(16) + '000000').substring(2, 8);
  };
  if(status == "home"){
    return (
      <div>
        <MenuComponent setStatus={setStatus} setLoginFunction={handleLogoutFunction}/>
        <div>
          <div className="container">
            <div style={{
              flex: 1,
              overflow: 'hidden',
            }}>
            <h2>Home</h2>
            <p>Welcome to the home page! You can visualize your flags here:</p>
            <BarChartComponent label="Your personal exploits:" data={personalExploitsdata} options={optionsCenter}/>
            <div style={{
                display: 'flex',
                flexDirection: 'row', /* Display children horizontally */
                width: '100%', /* Occupy all horizontal space */
              }}>
            {personalExploitsdata.labels.map((exploit) => 
              (
                <ScriptRunnerControlButton script={exploit} initialState={stoppedExploits.indexOf(exploit)<0?"started":"stopped"} style={{flex:1}}></ScriptRunnerControlButton>
              )
              )}
            </div>
            </div>
            <div style={{
                flex: 1,
                overflow: 'hidden',
                height: '100%'
            }}>
            <BarChartComponent label="Your team exploits:" data={teamExploitsdata} options={optionsTop}/>
            <BarChartComponent label="Your team submissions:" data={teamStatusdata} options={optionsTop} />
            </div>
          </div>
        </div>
  
      </div>
      
    );
  }
  else if(status == "explore"){
    return (
      <div>
        <MenuComponent setStatus={setStatus} setLoginFunction={handleLogoutFunction}/>
        <div>
          <h2>Explore</h2>
          <p>Welcome to the explore page! You can visualize all the flags that have been submitted by your team.</p>
          <FlagTableComponent data={flagsData}></FlagTableComponent>
        </div>
  
      </div>
      
    );
  }
  else if(status == "submit"){
    return (
      <div>
        <MenuComponent setStatus={setStatus} setLoginFunction={handleLogoutFunction}/>
        <div>
          <h2>Submit</h2>
          <p>Welcome to the submit page! You can insert here some flags you have stolen by hand:</p>
          <FlagSubmissionComponent user={username}></FlagSubmissionComponent>
        </div>
  
      </div>
      
    );
  }
  else if(status == "logout"){
    handleLogoutFunction()
  }
  else{
    alert("Unknown status "+status)
    setStatus("home")
  }
};

export default Home;

