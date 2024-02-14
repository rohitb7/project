package main

import (
	"flag"
	"github.com/dlintw/goconf"
	log "github.com/sirupsen/logrus"
	"time"
	"www.rvb.com/blob-service/s3-manager"
)

var devEnv bool

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

	// Define flags
	dev := flag.Bool("dev", false, "Use development configuration file")
	flag.Parse()

	devEnv = *dev

	var configFile string

	// Use development configuration file if specified, otherwise use Docker configuration file
	if *dev {
		log.Info("running on dev env")
		configFile = "./patient_image_service_dev.conf"
	} else {
		log.Info("running on docker env")
		configFile = "./patient_image_service_docker.conf"
	}

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
