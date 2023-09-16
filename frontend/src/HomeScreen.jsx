import { useState } from "react";
import "./HomeScreen.css";
import reactLogo from "./assets/react.svg";
import Sidebar from "./components/sidebar/sidebar.jsx";

<Route path="/Dashboard" component={HomeScreen} />;

function HomeScreen() {
  return (
    <div className="home-page-wrapper">
      <div className="navbar">
        <div className="navbar-image">
          <img src={reactLogo} alt="React Logo" />
        </div>
        <div className="navbar-search-wrapper">
          <input
            type="text"
            placeholder="Search"
            className="navbar-search-input"
          />
          <div className="navbar-search-button-wrapper">
            <button className="navbar-search-button">
              <div
                style={{ width: "100%", height: "100%", fill: "currentcolor" }}
              >
                <svg
                  enableBackground="new 0 0 24 24"
                  height="24"
                  viewBox="0 0 24 24"
                  width="24"
                  focusable="false"
                  style={{
                    pointerEvents: "none",
                    display: "block",
                    width: "100%",
                    height: "100%",
                  }}
                >
                  <path d="m20.87 20.17-5.59-5.59C16.35 13.35 17 11.75 17 10c0-3.87-3.13-7-7-7s-7 3.13-7 7 3.13 7 7 7c1.75 0 3.35-.65 4.58-1.71l5.59 5.59.7-.71zM10 16c-3.31 0-6-2.69-6-6s2.69-6 6-6 6 2.69 6 6-2.69 6-6 6z"></path>
                </svg>
              </div>
            </button>
          </div>
        </div>

        <div className="navbar-profile-wrapper">
          <div className="navbar-profile-icon">
            {/* Draw circle for now */}
            <svg
              width="40"
              height="40"
              viewBox="0 0 40 40"
              fill="none"
              xmlns="http://www.w3.org/2000/svg"
            >
              <circle cx="20" cy="20" r="18" fill="red" />
            </svg>
          </div>
        </div>
      </div>
      <div className="sidebar-container">
        <Sidebar />
      </div>
      <div className="main-content"></div>
    </div>
  );
}

export default HomeScreen;
