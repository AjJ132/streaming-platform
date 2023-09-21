// uploadLogic.js
export const uploadVideo = async (video, title, description) => {
  // Initialize variables and configurations
  // e.g., chunkSize, uploadURL, etc.
  const maxChunkSize = 1000000; // 1MB

  // Step 1: Grab Info from video
  // e.g., video format, duration, size, etc.

  //Total byte size of video file
  const videoSize = video.size;
  console.log(videoSize);

  // Step 2: Determine max bits for each chunk
  //TEMP max bits for each chunk is maxChunkSize
  //Determine num of chunks
  const numOfChunks = Math.ceil(videoSize / maxChunkSize);

  // Step 3: Prepare Metadata file
  // e.g., JSON with title, description, total chunks, etc.
  const metadata = {
    title: title,
    description: description,
    numOfChunks: numOfChunks,
    videoSize: videoSize,
    uuid: "",
  };
  // Step 4: Generate a unique identifier (UUID) for the video
  // This UUID will be used to create a folder for storing chunks.
  const uuid = uuidv4();
  metadata.uuid = uuid;

  // Step 5: Start the upload session by sending metadata and UUID to the server
  // The server initializes storage based on the UUID and waits for chunks.
  const uploadURL = `/upload/metadata/${metadata.uuid}`;
  //POST metadata to server
  const response = await fetch(uploadURL, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(metadata),
  });

  //confirm metadata was received by server via 200 status code
  if (response.status !== 200) {
    console.log("Error: Metadata not received by server. Please try again.");
    return;
  }
  // Step 6: Split video into chunks
  // Read the video blob and slice it into smaller blobs.
  //Set array to hold video chunks
  const videoChunks = [];

  //loop through video and slice into chunks
  for (let i = 0; i < numOfChunks; i++) {
    //slice video into chunks
    const start = i * maxChunkSize;
    const end = Math.min(videoSize, start + maxChunkSize);
    const chunk = video.slice(start, end);

    //add chunk to array
    videoChunks.push(chunk);
  }
  // Step 7: Upload chunks
  // Each chunk is uploaded with its sequence number and UUID.
  //loop through videoChunks array and upload each chunk
  for (let i = 0; i < videoChunks.length; i++) {
    //create form data
    const formData = new FormData();
    formData.append("chunk", videoChunks[i]);
    formData.append("chunkNumber", i);
    formData.append("uuid", uuid);

    //upload chunk to server
    const uploadChunkURL = `/upload/chunk/${uuid}`;
    const response = await fetch(uploadChunkURL, {
      method: "POST",
      body: formData,
    });

    //confirm chunk was received by server via 200 status code
    if (response.status !== 200) {
      console.log("Error: Chunk not received by server. Please try again.");
      return;
    }
  }
  // Optionally, listen for acknowledgment from the server for each chunk.
  // Step 8: Notify the server that all chunks are uploaded
  // This can trigger the server-side assembly and transcoding process.
  const uploadCompleteURL = `/upload/complete/${uuid}`;
  // Step 9: Handle errors and retries
  // e.g., if a chunk fails to upload, retry a certain number of times before failing.
  // Step 10: Update the UI or redirect the user
  // e.g., show a "Video uploaded successfully" message or navigate to another page.
  console.log("Upload complete!");
};
