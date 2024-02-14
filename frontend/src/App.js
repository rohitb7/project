import { useEffect, useState } from "react";
import "./App.css";
import { Search } from "./components/search";
import { List } from "./components/list";
import { fetchPatientImages } from "./apiClient/patients";
import { Upload } from "./components/upload";
import { BasicTable } from "./components/table";

function App() {
  const [imageUrl, setImageUrl] = useState("");
  const [patientImages, setPatientImages] = useState([]);

  const handleImageClick = (url) => {
    setImageUrl(url);
  };

  const getPatientImages = async () => {
    const result = await fetchPatientImages();
    console.log("in use effect", result);
    if (result && result.result && result.result.requestResult === "ACCEPTED") {
      setPatientImages(result.images);
    } else {
      throw new Error("Error");
    }
  };

  useEffect(() => {
    void getPatientImages();
  }, []);

  return (
    <>
      <header className="welcome-header">
        <h1>{`Welcome to Intuitive`}</h1>
      </header>
      <div className="main-container">
        <div className="image-detail-container">
          <BasicTable
            tableData={patientImages}
            handleImageClick={handleImageClick}
          />
          <Upload />
        </div>
        <div className="display-container">
          {imageUrl ? (
            <img src={imageUrl} className="image" />
          ) : (
            <div>{`No data`}</div>
          )}
        </div>
      </div>
    </>
  );
}

export default App;
