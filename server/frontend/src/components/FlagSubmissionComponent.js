import { Button } from '@mui/material';
import React, { useState } from 'react';

const FlagSubmissionComponent = ({username}) => {
  const [flagText, setFlagText] = useState('');

  const handleFlagChange = (event) => {
    setFlagText(event.target.value);
  };
  const getCurrentDateTime = () => {
    const now = new Date();
    const formattedDateTime = now.toLocaleString('en-US', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      hour12: false // Use 24-hour format
    }).replace(',', ' '); // Remove the comma between date and time
  
    return formattedDateTime;
  }

  const handleSubmit = () => {
    // Split the flagText into individual lines (flags)
    const flags = flagText.split('\n').map((flag) => flag.trim()).map((f)=> ({
                                                                        flag:f, 
                                                                        username:username, 
                                                                        exploit_name:"manual", 
                                                                        team_ip:"", 
                                                                        time:getCurrentDateTime(), 
                                                                        status:"NOT_SUBMITTED", 
                                                                        server_response:"NOT_SUBMITTED"}));
    const data = {username:username, flags:flags}
    alert(data)
    // Send the flags to the server using a POST request
    fetch('/upload_flags', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Auth-Token': 'token', // Add the custom header "X-Auth-Token" with your token
      },
      body: JSON.stringify(data), // Send the array of flags
    })
      .then((response)=>{
        if(!response.ok){
          throw new Error("HTTP error! Status: ${response.status}")
        }
      })
      .then((data) => {
        // Handle response data if needed
        alert("Flags correctly submitted")
        console.log('Flag submission response:', data);
      })
      .catch((error) => {
        alert("Error on submitting flags")
        console.error('Error submitting flags:', error);
      });

    // Clear the textarea after submission
    setFlagText('');
  };

  return (
    <div>
      <h1>Flag Submission</h1>
      <textarea
        rows={4}
        cols={50}
        value={flagText}
        onChange={handleFlagChange}
        placeholder="Enter your flags here"
      />
      <br />
      <Button onClick={handleSubmit}>Submit Flag</Button>
    </div>
  );
};

export default FlagSubmissionComponent;
