import React, {useEffect, useState} from 'react';
import { Button } from '@mui/material';

const ScriptRunnerControlButton = ({script, initialState, style, token})=>{

    const [status, setStatus] = useState(initialState)

    const onclickFunction = (event) => {
        if(status === "started"){
                fetch('/stop_exploit', {
                    method: 'POST',
                    headers: {
                      'Content-Type': 'application/x-www-form-urlencoded',
                      'X-Auth-Token': token,
                    },
                    body: "exploit="+event.target.id,
                  })
                    .then((response)=>{
                      if(!response.ok){
                        if(response.status == 400 && typeof response.text != "undefined")
                            throw new Error(response.text)
                        else
                            throw new Error("HTTP error! Status: ${response.status}")
                      }
                    })
                    .then((data) => {
                      alert("Exploit correctly stopped")
                      console.log('Exploit stopping response:', data);
                    })
                    .catch((error) => {
                      alert("Error on stopping exploit")
                      console.error('Error stopping exploit:', error);
                    });
                  
                  setStatus("stopped")
        }
        else if(status === "stopped"){
            fetch('/restart_exploit', {
                method: 'POST',
                headers: {
                  'Content-Type': 'application/x-www-form-urlencoded',
                  'X-Auth-Token': token, // Add the custom header "X-Auth-Token" with your token
                },
                body: "exploit="+event.target.id,
              })
                .then((response)=>{
                  if(!response.ok){
                    throw new Error("HTTP error! Status: ${response.status}")
                  }
                })
                .then((data) => {
                  alert("exploit correctly restarted")
                  console.log('Exploit restart response:', data);
                })
                .catch((error) => {
                  alert("Error on restarting exploit")
                  console.error('Error restarting exploit:', error);
                });
          
            setStatus("started")
        }
        else {
            setStatus("started")
        }
    }
    return(
        <Button id={script} onClick={onclickFunction} style={style} color={status==="stopped"?"error":"success"} token={token}>{status}</Button>
    )
}
export default ScriptRunnerControlButton;