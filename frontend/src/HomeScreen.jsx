import { useState } from "react";
import "./HomeScreen.css";
import reactLogo from "./assets/react.svg";

function HomeScreen() {
  const [count, setCount] = useState(0);

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
          <button className="navbar-search-button"></button>
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
    </div>
  );
}

export default HomeScreen;
