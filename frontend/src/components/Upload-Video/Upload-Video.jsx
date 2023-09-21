import React, { useState } from "react";
import "./Upload-Video.css";

function UploadVideo() {
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [video, setVideo] = useState(null);

  const handleSubmit = (e) => {
    e.preventDefault();
    // Logic to upload the video
    console.log("Title:", title);
    console.log("Description:", description);
    console.log("Video File:", video);
  };

  return (
    <div className="upload-video-wrapper">
      <h2>Upload Video</h2>
      <form onSubmit={handleSubmit}>
        <div>
          <label htmlFor="title">Title:</label>
          <input
            type="video-text"
            id="title"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
          />
        </div>
        <div>
          <label htmlFor="description">Description:</label>
          <textarea
            id="description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
          />
        </div>
        <div>
          <label htmlFor="video">Video (.mp4 only):</label>
          <input
            type="file"
            id="video"
            accept=".mp4"
            onChange={(e) => setVideo(e.target.files[0])}
          />
        </div>
        <button type="submit">Upload</button>
      </form>
    </div>
  );
}

export default UploadVideo;
