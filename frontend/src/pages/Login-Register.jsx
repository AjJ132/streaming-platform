import React, { useState } from "react";
import { Navigate } from "react-router-dom";
import jwt_decode from "jwt-decode";
import axios from "axios";
import "./Login-Register.css";

function Login_Register() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [isRegistered, setIsRegistered] = useState(true);

  let authToken = localStorage.getItem("token");

  if (authToken !== null) {
    let decodedToken = jwt_decode(authToken);
    let currentDate = new Date();
    if (decodedToken.exp * 1000 >= currentDate.getTime()) {
      return <Navigate to="/HomeScreen" />;
    }
  }

  const handleLogin = async () => {
    try {
      const response = await axios.post("/api/login", { email, password });
      localStorage.setItem("token", response.data.token);
      // TODO: Navigate to HomeScreen
    } catch (error) {
      console.error("Login failed", error);
    }
  };

  const handleRegister = async () => {
    try {
      const response = await axios.post("/api/register", { email, password });
      localStorage.setItem("token", response.data.token);
      // TODO: Navigate to HomeScreen
    } catch (error) {
      console.error("Registration failed", error);
    }
  };

  return (
    <div className="login-register-container">
      {isRegistered ? (
        <>
          <h2>Login</h2>
          <input
            type="email"
            placeholder="Email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
          <input
            type="password"
            placeholder="Password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
          <button onClick={handleLogin}>Login</button>
          <button onClick={() => setIsRegistered(false)}>
            Switch to Register
          </button>
        </>
      ) : (
        <>
          <h2>Register</h2>
          <input
            type="email"
            placeholder="Email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
          <input
            type="password"
            placeholder="Password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
          <button onClick={handleRegister}>Register</button>
          <button onClick={() => setIsRegistered(true)}>Switch to Login</button>
        </>
      )}
    </div>
  );
}

export default Login_Register;
