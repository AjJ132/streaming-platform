import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter as Router, Route, Routes } from "react-router-dom";

import "./index.css";
import Login_Register from "./pages/Login-Register";
import HomeScreen from "./HomeScreen";

ReactDOM.createRoot(document.getElementById("root")).render(
  <React.StrictMode>
    <Router>
      <Routes>
        <Route path="/" element={<Login_Register />} />
        <Route path="/Dashboard" element={<HomeScreen />} />
      </Routes>
    </Router>
  </React.StrictMode>
);
