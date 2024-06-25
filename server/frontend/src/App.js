import React, { useState, useEffect} from "react";
import Login from "./components/Login";
import Home from "./components/Home";

const App = () => {
  const token = 'token'
  const [loggedIn, setLoggedIn] = useState(
    {
      status: false,
      username: ""
    }
  );

  const handleLogin = (user, password) =>{
    fetch('/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: "username="+user+"&password="+password,
    })
      .then((response) => {
        if (response.ok) {
          setLoggedIn({status:true, username:user})
        } else {
          alert('Invalid credentials');
        }
      })
      .catch((error) => {
        alert("Error occurred during log in")
        console.error('Error occurred while logging in:', error);
      });
  }

  const handleLogout = () =>{
    fetch('/logout', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
        'token':token
      },
    })
      .then((response) => {
        if (response.ok) {
          setLoggedIn({status: false, username: ""})
        } else {
          alert('Failed to log out');
        }
      })
      .catch((error) => {
        alert("Error occurred during log out")
        console.error('Error occurred while logging out:', error);
      });
    
  }

  useEffect(() => {
    const checkAuth = async () => {
      try {
        const response = await fetch('/check_auth', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'X-Auth-Token': 'token', // Add the custom header "X-Auth-Token" with your token
          },
        });

        if (!response.ok) {
          setLoggedIn({ status: false, username: "" });
        } else {
          const data = await response.json();
          console.log(data);

          if (data && data.username) {
            const curUser = data.username;
            console.log(curUser);
            setLoggedIn({ status: true, username: curUser });
          } else {
            setLoggedIn({ status: false, username: "" });
          }
        }
      } catch (error) {
        console.error('Error:', error);
        setLoggedIn({ status: false, username: "" });
      }
    };

    checkAuth();
  }, []); // The empty array makes it run only once after the initial render


  return (loggedIn.status)?
    <Home handleLogoutFunction={handleLogout} username={loggedIn.username} token={token}/>:
    <Login handleLoginFunction={handleLogin}/>
  
};

export default App;

