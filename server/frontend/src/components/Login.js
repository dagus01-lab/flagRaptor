import { Input, Button } from "@mui/material";
import React, { useState } from "react";
import { Navigate } from "react-router-dom";

const Login = ({ handleLoginFunction }) => {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");

  const handleSubmit = async (e) => {
    e.preventDefault();
    handleLoginFunction(username, password);
  };

  return (
    <div className="login_page">
      <div className="login_form">
        <h2>Login</h2>
        <form onSubmit={handleSubmit}>
          <Input
            type="text"
            placeholder="Username"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
          /><br></br>
          <Input
            type="password"
            placeholder="Password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          /><br></br>
          <Button type="submit">Login</Button>
        </form>
      </div>
    </div>
  );
};

export default Login;

