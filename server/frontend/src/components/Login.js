import React, { useState } from "react";
import { Navigate } from "react-router-dom";

const Login = ({ handleLoginFunction }) => {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");

  const handleSubmit = async (e) => {
    e.preventDefault();
    fetch('/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      //credentials: 'include',
      body: "username="+username+"&password="+password,
    })
      .then((response) => {
        if (response.ok) {
          // If login is successful, call the onLogin callback to set the loggedIn state in the parent component
          handleLoginFunction(username);
        } else {
          // Handle login failure here, such as displaying an error message
          alert('Invalid credentials');
        }
      })
      .catch((error) => {
        // Handle any network or server-related errors here
        alert("Error occurred during log in")
        console.error('Error occurred while logging in:', error);
      });
  };

  return (
    <div>
      <h2>Login</h2>
      <form onSubmit={handleSubmit}>
        <input
          type="text"
          placeholder="Username"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
        />
        <input
          type="password"
          placeholder="Password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
        <button type="submit">Login</button>
      </form>
    </div>
  );
};

export default Login;

