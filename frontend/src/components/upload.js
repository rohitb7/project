import React, { useState } from "react";
import { uploadFile } from "../apiClient/patients";
import Button from "@mui/material/Button";
import { Input } from "@mui/material";

export const Upload = () => {
  const [selectedFile, setSelectedFile] = useState(null);

  const handleFileChange = (event) => {
    setSelectedFile(event.target.files[0]);
  };

  const handleUploadClick = async () => {
    if (!selectedFile) {
      alert("Please select a file first!");
      return;
    }

    try {
      console.log("selectedFile", selectedFile);
      await uploadFile(selectedFile, selectedFile.name, selectedFile.type);
    } catch (error) {
      console.error("Error uploading file:", error);
    }
  };

  return (
    <div className="upload">
      <input type="file" onChange={handleFileChange} />
      <Button
        sx={{
          color: "black",
          borderColor: "black",
        }}
        variant="outlined"
        onClick={handleUploadClick}
        disabled={selectedFile === null}
      >
        {"Upload File"}
      </Button>
    </div>
  );
};
