// import { v4 as uuidv4 } from "uuid";

// uploadLogic.js
// export const uploadVideo = async (video, title, description) => {
//   // Initialize variables and configurations
//   // e.g., chunkSize, uploadURL, etc.
//   const maxChunkSize = 1000000; // 1MB

//   // Step 1: Grab Info from video
//   // e.g., video format, duration, size, etc.

//   //Total byte size of video file
//   const videoSize = video.size;
//   console.log(videoSize);

//   // Step 2: Determine max bits for each chunk
//   //TEMP max bits for each chunk is maxChunkSize
//   //Determine num of chunks
//   const numOfChunks = Math.ceil(videoSize / maxChunkSize);

//   // Step 3: Prepare Metadata file
//   // e.g., JSON with title, description, total chunks, etc.
//   const metadata = {
//     title: title,
//     description: description,
//     numOfChunks: numOfChunks,
//     videoSize: videoSize,
//     uuid: "",
//   };
//   // Step 4: Generate a unique identifier (UUID) for the video
//   // This UUID will be used to create a folder for storing chunks.
//   const uuid = uuidv4();
//   metadata.uuid = uuid;

//   console.log(uuid);
//   console.log(metadata);

//   // const end = Math.min(videoBlob.size, start + chunkSize);
//   // const chunk = videoBlob.slice(start, end);

//   // await fetch(url, {
//   //   method: 'POST',
//   //   headers: {
//   //     'Authorization': 'Bearer YOUR_PASSKEY',
//   //     'Content-Range': `bytes ${start}-${end}/${videoBlob.size}`
//   //   },
//   //   body: chunk
//   // });

//   // Step 5: Start the upload session by sending metadata and UUID to the server
//   // The server initializes storage based on the UUID and waits for chunks.
//   const uploadURL = `/upload/metadata/${metadata.uuid}`;
//   //POST metadata to server
//   const response = await fetch(uploadURL, {
//     method: "POST",
//     headers: {
//       "Content-Type": "application/json",
//     },
//     body: JSON.stringify(metadata),
//   });

//   //confirm metadata was received by server via 200 status code
//   if (response.status !== 200) {
//     console.log("Error: Metadata not received by server. Please try again.");
//     return;
//   }
//   // Step 6: Split video into chunks
//   // Read the video blob and slice it into smaller blobs.
//   //Set array to hold video chunks
//   const videoChunks = [];

//   //loop through video and slice into chunks
//   for (let i = 0; i < numOfChunks; i++) {
//     //slice video into chunks
//     const start = i * maxChunkSize;
//     const end = Math.min(videoSize, start + maxChunkSize);
//     const chunk = video.slice(start, end);

//     //add chunk to array
//     videoChunks.push(chunk);
//   }
//   // Step 7: Upload chunks
//   // Each chunk is uploaded with its sequence number and UUID.
//   //loop through videoChunks array and upload each chunk
//   for (let i = 0; i < videoChunks.length; i++) {
//     //create form data
//     const formData = new FormData();
//     formData.append("chunk", videoChunks[i]);
//     formData.append("chunkNumber", i);
//     formData.append("uuid", uuid);

//     //upload chunk to server
//     const uploadChunkURL = `/upload/chunk/${uuid}`;
//     const response = await fetch(uploadChunkURL, {
//       method: "POST",
//       body: formData,
//     });

//     //confirm chunk was received by server via 200 status code
//     if (response.status !== 200) {
//       console.log("Error: Chunk not received by server. Please try again.");
//       return;
//     }
//   }
//   // Optionally, listen for acknowledgment from the server for each chunk.
//   // Step 8: Notify the server that all chunks are uploaded
//   // This can trigger the server-side assembly and transcoding process.
//   const uploadCompleteURL = `/upload/complete/${uuid}`;
//   // Step 9: Handle errors and retries
//   // e.g., if a chunk fails to upload, retry a certain number of times before failing.
//   // Step 10: Update the UI or redirect the user
//   // e.g., show a "Video uploaded successfully" message or navigate to another page.
//   console.log("Upload complete!");
// };

export const uploadVideo = async (video, title, name) => {
  RequestUpload(name, title, video);
};

export const RequestUpload = async (name, title, video) => {
  const metadata = {
    name: name,
    videoName: title,
  };

  const authToken = localStorage.getItem("token");
  //const uploadURL = `/upload/request`;
  //TEMP URL SINCE NGINX ISNT RUNNING
  const uploadURL = `http://localhost:8086/request`;
  //POST metadata to server
  const response = await fetch(uploadURL, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: authToken,
    },

    body: JSON.stringify(metadata),
  });

  if (response.status !== 200 && response.status !== 500) {
    console.log(
      "Error: There was an error with the request. Please try again."
    );
    return;
  } else if (response.status === 500) {
    console.log("Error: Internal Server Error ");
    +console.log(response);
    return;
  }

  //decode body
  const body = await response.json();

  //get token from body
  const token = body.queueToken;

  //check if token was received
  if (token === undefined) {
    console.log("Error: Token not received by server. Please try again.");
    return;
  } else if (token === "") {
    console.log("Error: Invalid Token. Please try again.");
    return;
  }

  //if token is valid store in local storage
  localStorage.setItem("queueToken", token);

  //connect to queue
  ConnectToQueue(video);
};

export const ConnectToQueue = (video) => {
  const token = localStorage.getItem("queueToken");
  let clientID = "";
  clearInterval;
  const ws = new WebSocket(
    `ws://localhost:8010/ws/connect?queueToken=${token}`
  );

  ws.onopen = () => {
    // When the connection is open, send a message to the server to carry the token.
    ws.send(JSON.stringify({ token }));
  };

  ws.onmessage = (event) => {
    console.log("Received:", event.data);
    const message = event.data;

    if (message.includes("CLIENTID") && message.includes("PASSKEY")) {
      const parts = message.split(";");
      const clientIDPart = parts.find((part) => part.startsWith("CLIENTID:"));
      const passkeyPart = parts.find((part) => part.startsWith("PASSKEY:"));

      if (clientIDPart && passkeyPart) {
        const clientID = clientIDPart.split(":")[1];
        const passkey = passkeyPart.split(":")[1];

        console.log("Received ClientID:", clientID);
        console.log("Received passkey:", passkey);

        HandleUpload(video, passkey, clientID);
      }
    }
  };

  ws.onerror = (error) => {
    console.log(`WebSocket Error: ${error}`);
  };

  ws.onclose = (event) => {
    console.log("WebSocket closed:", event);
  };
};

export const HandleUpload = async (video, passKey, clientID) => {
  const maxChunkSize = 1000000; // 1MB

  //Total byte size of video file
  const videoSize = video.size;
  console.log(videoSize);

  //Determine max bits for each chunk
  const numOfChunks = Math.ceil(videoSize / maxChunkSize);

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

  // Each chunk is uploaded with its sequence number and UUID.
  //loop through videoChunks array and upload each chunk
  for (let i = 0; i < videoChunks.length; i++) {
    //create form data
    const formData = new FormData();
    formData.append("chunk", videoChunks[i]);

    //upload chunk to server
    //const uploadChunkURL = `/upload/handle-upload`;
    //TEMP URL SINCE NGINX ISNT RUNNING
    const uploadChunkURL = `http://localhost:8010/handle-upload`;
    const response = await fetch(uploadChunkURL, {
      method: "POST",
      body: formData,
      headers: {
        "Chunk-Number": i,
        "Video-Name": "TEMP-NAME",
        "Client-ID": clientID,
        Authorization: passKey,
      },
    });

    //confirm chunk was received by server via 200 status code
    if (response.status !== 200) {
      console.log("Error: Chunk not received by server. Please try again.");
      return;
    }
  }
};
