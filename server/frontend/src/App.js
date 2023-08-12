import React, { useState } from "react";
import Login from "./components/Login";
import Home from "./components/Home";

const App = () => {
  const [loggedIn, setLoggedIn] = useState(
    {
      status: false,
      username: ""
    }
  );
  const handleLogin = (user) =>{
    setLoggedIn({status:true, username:user})
  }

  const handleLogout = () =>{
    setLoggedIn({status: false, username: ""})
  }
  if(loggedIn.status){
    return(<Home handleLogoutFunction={handleLogout} username={loggedIn.username}/>)
  }
  else{
    return(<Login handleLoginFunction={handleLogin}/>)
  }
};

export default App;

