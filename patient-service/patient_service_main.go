package main

import (
	"github.com/dlintw/goconf"
	log "github.com/sirupsen/logrus"
	"time"
	"www.rvb.com/blob-service/s3-manager"
)

const (
	DefaultMaxRetry              = 1
	DefaultRetryTimeout          = 2 * time.Second
	DefaultTaskQueueSize         = 500
	DefaultWorkPoolSize          = 50
	DefaultGrpcServerAddr        = ":9798"
	DefaultMode                  = "dev"
	DefaultRestServerAddr        = ":9797"
	DefaultGrpcRestServerRetries = 3
)

func main() {
	//// Get the current working directory
	//currentDir, err := os.Getwd()
	//if err != nil {
	//	log.Fatalf("Failed to get the current working directory: %v", err)
	//}
	//fmt.Println("Current directory:", currentDir)

	configFile := "./patient_image_service.conf"

	c, err := goconf.ReadConfigFile(configFile)
	if err != nil {
		log.Error("confInitTest: ReadConfigFile failed: %v", err)
		return
	}
	// configure required sections. postgres, rest, grpc, workers , job managers
	configErr := configureSections(c)
	if configErr != nil {
		log.WithFields(log.Fields{"error": configErr}).Error("Config error")
		return
	}
	initJobManger()
	JobManagerInstance.Start()
	//postgresStart()
	// start REST/gRPC gateway
	if err = restGrpcCfg.Start(); err != nil {
		log.WithFields(log.Fields{"error": err}).Error("REST/gRPC server would not start")
		return
	}
	// init the south bound image service
	err = initBlobStorageService(c)
	if err != nil {
		log.Error("initBlobStorageService failed: %v", err)
		return
	}
	log.Info("WAITING FOR REQUESTS!!!!!!!")
	select {}
}

// initBlobStorageService
func initBlobStorageService(s3conf *goconf.ConfigFile) error {
	var err error
	serverCfg.storageManagerInterface, err = s3_manager.NewProductionStorage(s3conf)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

// CloseAll Close all the connections
func CloseAll() {
	restGrpcCfg.Close()
	serverCfg.Close()
	JobManagerInstance.Close()
	workPoolCfg.Close()
}
