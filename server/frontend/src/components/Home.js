import React, {useEffect, useState} from "react";
import BarChartComponent from "./BarChartComponent";
import FlagTableComponent from "./FlagTableComponent";
import FlagSubmissionComponent from "./FlagSubmissionComponent";
import MenuComponent from "./MenuComponent"
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

  useEffect(() => {
    const socket = new WebSocket('ws://localhost:5000/update_flags');
    // Listen for real-time updates from the server
    socket.onmessage = (event) => {
     handleFlagUpdate(event.data); // Handle the real-time flag update
    };

    return () => {
      socket.close(); // Clean up the WebSocket connection on unmount
    };
  }, []); //the empty array ensures that the function is executed only when the component is loaded

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

    setPersonalExploitsData({ labels: personalExploits, datasets : personalExploitsDatasets});
    setTeamExploitsData({ labels: teamExploits, datasets: teamExploitsDatasets});
    setTeamStatusData({labels: teams, datasets:teamstatusdatasets})
  };

  const handleFlagUpdate = (jsonString) => {
    try {
      const jsonObject = JSON.parse(jsonString);
      console.log(jsonObject);

      jsonObject.forEach((flag)=>{
        if(flagsData.map((f) => f.flag).includes(flag.flag)){
          flagsData.filter((f) => f.flag === flag.flag).forEach((f) => {f.server_response = flag.server_response; f.status = flag.status;})
        }
        flagsData.push(flag)
      })
      if(status === "home")
        updateBarChartData(); // Update the chart data with the updated flags data
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
          <h2>Home</h2>
          <p>Welcome to the home page! You can visualize your flags here.</p>
          <BarChartComponent label="Your team exploits:" data={teamExploitsdata}/>
          <BarChartComponent label="Your personal exploits:" data={personalExploitsdata}/>
          <BarChartComponent label="Your team submissions:" data={teamStatusdata}/>
        </div>
  
      </div>
      
    );
  }
  else if(status == "explore"){
    return (
      <div>
        <MenuComponent setStatus={setStatus} setLoginFunction={handleLogoutFunction}/>
        <div>
          <h2>Home</h2>
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
          <h2>Home</h2>
          <p>Welcome to the submit page! You can insert some flags you retrieved by hand here:</p>
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

