package main

import (
	"database/sql"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	_ "github.com/lib/pq" // Import the PostgreSQL driver
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
	"www.rvb.com/patient-service/protoenumutils"

	//"www.rvb.com/patient-service/protoenumutils"
	protos "www.rvb.com/protos"
)

func postgresStart() {
	db := connectToDB()
	defer db.Close()
}

//func connectToDB() *sql.DB {
//
//	// Open the connection
//	db, err := sql.Open("postgres", psqlInfo)
//	if err != nil {
//		log.Fatal("Error connecting to the database: ", err)
//	}
//
//	// Check the connection
//	err = db.Ping()
//	if err != nil {
//		log.Fatal("Error pinging the database: ", err)
//	}
//
//	fmt.Println("Successfully connected to the database!")
//	return db
//}

// FOR DOCKER
func connectToDB() *sql.DB {

	var psqlInfo string

	if restGrpcCfg.Mode == "dev" {
		//FOR LOCAL DEVELOPMENT
		psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
			"password=%s dbname=%s sslmode=disable",
			sqlDbCfg.Server, sqlDbCfg.Port, sqlDbCfg.Username, sqlDbCfg.Password, sqlDbCfg.DatabaseName)

		println("psqlInfo", psqlInfo)
	} else {
		// FOR DOCKER
		psqlInfo = "host=postgres-patients user=postgres " + "password=mysecretpassword dbname=patients_db sslmode=disable"
	}

	// Open the connection
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
	}

	// Check the connection
	err = db.Ping()
	if err != nil {
		log.Fatal("Error pinging the database: ", err)
	}

	fmt.Println("Successfully connected to the database!")
	return db
}

// handleRetrievePatientImage retrieves patient images along with tags
func retrievePatientImageHandlerDB(protoRequest *protos.ListPatientImagesRequest) (*protos.ListPatientImagesResponse, error) {
	var err error

	log.Info("Starting retrieval of patient images.")

	// Initialize the response
	response := &protos.ListPatientImagesResponse{
		Patient: &protos.Patient{
			Id:       "",
			Name:     "",
			UserName: "",
		},
		Images: nil,
		Result: &protos.Result{
			RequestResult: 0,
			Error: &protos.Error{
				Message:          "",
				ErrorRetryStatus: 0,
			},
		},
	}

	// Define retry status
	retryStatus, err := protoenumutils.GetErrorRetryStatusEnumValueFromString(protos.ErrorRetryStatus_RETRY.String())
	if err != nil {
		return response, nil
	}

	defer func() {
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("Failed to get patient list of images")
			response.Result.RequestResult = protos.RequestResult_REJECTED
			response.Result.Error = &protos.Error{
				Message:          err.Error(),
				ErrorRetryStatus: retryStatus,
			}
		} else {
			log.Info("Successfully retrieved patient list of images")
			response.Result.RequestResult = protos.RequestResult_ACCEPTED
		}
	}()

	// Connect to the database
	db := connectToDB()
	if db == nil {
		log.Error("Failed to connect to the database.")
		return nil, fmt.Errorf("failed to connect to the database")
	}
	defer db.Close()

	// Assign patient ID from request to response
	response.Patient.Id = protoRequest.Patient.Id

	// Query to retrieve patient images along with tags based on the patient ID
	query := `
    SELECT i.bucket_path, i.upload_date, i.name, i.description, ARRAY_AGG(t.name) AS tags
    FROM images i
    LEFT JOIN image_tags it ON i.id = it.image_id
    LEFT JOIN tags t ON it.tag_id = t.id
    WHERE i.patient_id = $1
    GROUP BY i.bucket_path, i.upload_date, i.name, i.description;
`

	rows, err := db.Query(query, protoRequest.Patient.Id)
	if err != nil {
		return nil, fmt.Errorf("error querying images and tags: %v", err)
	}
	defer rows.Close()

	// Iterate over the rows and populate the response with images and tags
	var currentImage *protos.Image
	for rows.Next() {
		var name string
		var description string
		var bucketPath string
		var uploadDate time.Time
		var tagString string // Define a string to hold comma-separated tags
		if err := rows.Scan(&bucketPath, &uploadDate, &name, &description, &tagString); err != nil {
			return nil, fmt.Errorf("error scanning image and tag row: %v", err)
		}
		trimmed := strings.Trim(tagString, "{}")
		tags := strings.Split(trimmed, ",")
		// TODO: check this :Check if the image is the same as the previous one
		if currentImage == nil || currentImage.BucketPath != bucketPath {
			// If it's a new image, create a new Image object
			currentImage = &protos.Image{
				Tags:        nil,
				Name:        name,
				Description: description,
				BucketPath:  bucketPath,
				Url:         "",
				UploadTime:  ptypes.TimestampNow(), // Assuming it's the current time
			}
			response.Images = append(response.Images, currentImage)
		}
		// Append tags to the current image's tags
		currentImage.Tags = append(currentImage.Tags, tags...)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over image and tag rows: %v", err)
	}

	log.Info("Successfully retrieved patient images with tags from the database")
	return response, nil
}

func uploadPatientImageHandlerDB(req *protos.UploadPatientImageRequest, bucketPath string) (*protos.UploadPatientImageResponse, error) {
	var err error

	log.Info("Starting retrieval of patient images.")

	// Initialize the response
	response := &protos.UploadPatientImageResponse{
		Result: &protos.Result{
			RequestResult: 0,
			Error: &protos.Error{
				Message:          "",
				ErrorRetryStatus: 0,
			},
		},
	}

	// Define retry status
	retryStatus, err := protoenumutils.GetErrorRetryStatusEnumValueFromString(protos.ErrorRetryStatus_RETRY.String())
	if err != nil {
		return response, nil
	}

	defer func() {
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("Failed to get patient list of images")
			response.Result.RequestResult = protos.RequestResult_REJECTED
			response.Result.Error = &protos.Error{
				Message:          err.Error(),
				ErrorRetryStatus: retryStatus,
			}
		} else {
			log.Info("Successfully uploaded image")
			response.Result.RequestResult = protos.RequestResult_ACCEPTED
		}
	}()

	db := connectToDB()
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	var imageID int
	err = tx.QueryRow("INSERT INTO images (patient_id, bucket_path, name, description, upload_date) VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP) RETURNING id",
		req.PatientImage.Patient.Id, bucketPath, req.PatientImage.Image.Name, req.PatientImage.Image.Description).Scan(&imageID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if req.Tags != nil && len(req.Tags.Tag) > 0 {
		// Insert tags and associate them with the image
		for _, tagName := range req.Tags.Tag {
			var tagID int
			// Insert the tag if it doesn't exist and get its id
			err = tx.QueryRow("INSERT INTO tags (name) VALUES ($1) ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name RETURNING id", tagName).Scan(&tagID)
			if err != nil {
				tx.Rollback()
				return nil, err
			}

			_, err = tx.Exec("INSERT INTO image_tags (image_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", imageID, tagID)
			if err != nil {
				tx.Rollback() // rollback in case of error
				return nil, err
			}
		}
	}

	// Commit
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return response, nil
}
