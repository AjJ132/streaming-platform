import React from "react";
import "./sidebar.css";

import "font-awesome/css/font-awesome.min.css";

function Sidebar() {
  return (
    <div className="sidebar-wrapper">
      <div className="sidebar-content">
        <a className="sidebar-item">
          <i className="fa fa-home"></i> Home
        </a>
        <a className="sidebar-item">
          <i className="fa fa-hourglass"></i> History
        </a>
        <a className="sidebar-item">
          <i className="fa fa-user"></i> My Videos
        </a>
        <a className="sidebar-item">
          <i className="fa fa-gear"></i> Settings
        </a>
      </div>
    </div>
  );
}

export default Sidebar;
