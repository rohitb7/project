package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/dlintw/goconf"
	grpcruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"
	"www.rvb.com/blob-service/core"
	"www.rvb.com/commonutils"
	"www.rvb.com/patient-service/workerpool"
	protos "www.rvb.com/protos"
)

type NodeConfiguration interface {
	configure(nodeName string, c *goconf.ConfigFile) error
}

// Processed sections and configurations ,
// in the order to be processed

var MandatoryConfSections = []string{
	"server",
	"worker_pool",
	"sqldb",
	"http_server",
	"s3_config",
}

// Processed sections and corresponding function handlers
var (
	configurationConfModule = map[string]func() NodeConfiguration{ //nolint
		"server":      serverConfigurationNew,
		"sqldb":       sqlDbConfigurationNew,
		"worker_pool": workerPoolConfigurationNew,
		"http_server": restGrpcConfigurationNew,
		"s3_config":   S3BucketConfigurationNew,
	}
)

// Section Config variables , with Config state if any

type ServerConfiguration struct {
	maxRetry                int
	retryTime               time.Duration
	dbQueryTimeout          time.Duration
	uploadsLocation         string
	downloadsLocation       string
	dataLocation            string
	serverContext           context.Context
	serverCancel            context.CancelFunc
	testMode                bool
	storageManagerInterface core.StorageManager
	mux                     *sync.Mutex
}

// TODO need to update this struct to have the right items to be saved for the REST and GRPC gateway server
type S3BucketConfiguration struct {
	bucket string
}

type RestGrpcConfiguration struct {
	grpcPortAddress string
	httpPortAddress string
	retries         int
	allowCrud       bool
	Rest            RestServer
	GRPC            GRPCServer
	Mode            string
}

type SqldbConfiguration struct {
	Server       string
	Port         int
	Username     string
	Password     string
	DatabaseName string
}

type WorkerPoolConfiguration struct {
	masterPool       *workerpool.TaskManager
	highPriorityPool *workerpool.TaskManager
	workersCount     int
	queueSize        int
	totalQueueSize   int // master pool + high priority pool
	totalWorkerCount int // master pool + high priority pool
}

type RestServer struct {
	ConnContext context.Context
	GRPCMUX     *grpcruntime.ServeMux
	GRPCPort    string
	GRPCOption  []grpc.DialOption
}

type GRPCServer struct {
	Server *grpc.Server
}

func serverConfigurationNew() NodeConfiguration {
	return &ServerConfiguration{}
}

func restGrpcConfigurationNew() NodeConfiguration {
	return &RestGrpcConfiguration{}
}

func S3BucketConfigurationNew() NodeConfiguration {
	return &S3BucketConfiguration{}
}

func sqlDbConfigurationNew() NodeConfiguration {
	return &SqldbConfiguration{}
}

func workerPoolConfigurationNew() NodeConfiguration {
	return &WorkerPoolConfiguration{}
}

// Global Config objects per section
var (
	s3Cfg       *S3BucketConfiguration
	serverCfg   *ServerConfiguration
	restGrpcCfg *RestGrpcConfiguration
	sqlDbCfg    *SqldbConfiguration
	workPoolCfg *WorkerPoolConfiguration
)

// Configure each section from conf file, return error if any sections are not properly configured
func configureSections(c *goconf.ConfigFile) error { //nolint
	var configErr error = nil
	for _, section := range MandatoryConfSections {
		log.Info("Configuring ", section)
		if funcConfig, ok := configurationConfModule[section]; ok {
			err := funcConfig().configure(section, c)
			if err != nil {
				log.Error(log.Fields{"error": err, "section": section}, "Configuring section failed")
				configErr = err
			}
		} else {
			log.Error(log.Fields{"section": section}, "section configuration found in config module, "+"but not configurable")
		}
	}
	if configErr != nil {
		log.WithFields(log.Fields{"error": configErr}).Error("Config processing failed")
	} else {
		log.Info("Mandatory  Config sections processed successfully")
	}
	log.Debug("Mandatory Config sections processing done")

	return configErr
}

func (s3BucketConfig *S3BucketConfiguration) configure(section string, c *goconf.ConfigFile) error {
	var err error
	bucket, err := c.GetString(section, "bucket")
	if err != nil {
		log.WithFields(log.Fields{"error": err, "section": section}).Error("missing bucket")
		return err
	}
	s3Cfg = s3BucketConfig
	s3Cfg.bucket = bucket
	return nil
}

// Common server configuration
func (serverConfig *ServerConfiguration) configure(section string, c *goconf.ConfigFile) error {
	var err error
	var testMode bool
	var uploadsLocation string
	retryTime, err := c.GetString(section, "retry_time")
	if err != nil {
		serverConfig.retryTime = DefaultRetryTimeout
		err = nil //nolint
	} else {
		if serverConfig.retryTime, err = time.ParseDuration(retryTime); err != nil {
			log.WithFields(log.Fields{"retry_time": retryTime}).Error("Invalid retry_time, retry_time should be in format, example 2s or 2000ms, using defaults")
			serverConfig.retryTime = DefaultRetryTimeout
		}
	}
	log.WithFields(log.Fields{"retry_time": serverConfig.retryTime}).Info("Config Read retry_time")
	testMode, _ = c.GetBool(section, "test_mode")
	if testMode {
		log.Info("Test mode enabled")
	}
	//TODO  need to have default values in-case config is messed up so that we start up with some default vaules
	uploadsLocation, err = c.GetString(section, "uploads_location")
	if err != nil {
		log.WithFields(log.Fields{"error": err, "section": section}).Error("missing uploads location")
		return err
	}
	log.WithFields(log.Fields{"uploads_location": uploadsLocation}).Info("configured repo settings location")
	granularLabels, _ := c.GetBool("metamonitoring", "granular_labels")
	log.WithFields(log.Fields{"granular_labels": granularLabels}).Info("Config Read granular_labels")
	serverConfig.uploadsLocation = uploadsLocation
	serverConfig.serverContext, serverConfig.serverCancel = context.WithCancel(context.Background())
	serverConfig.testMode = testMode
	serverConfig.mux = new(sync.Mutex)
	serverCfg = serverConfig
	err = commonutils.CreateDirectoriesIfNotExist([]string{
		serverConfig.uploadsLocation,
	})
	if err != nil {
		log.WithFields(log.Fields{"section": section, "error": err}).Error("failed to setup directories for repo")
		return err
	}
	return nil
}

func (s *ServerConfiguration) Start() {
}

// Stop processing jobs from chan
func (s *ServerConfiguration) Close() {
	if serverCfg.serverCancel != nil {
		serverCfg.serverCancel()
	}
}

// Getters for directories
func (s *ServerConfiguration) GetDownloadsDirectory() string {
	if s != nil {
		return s.downloadsLocation
	}
	return ""
}

func (s *ServerConfiguration) GetUploadsLocation() string {
	if s != nil {
		return s.uploadsLocation
	}
	return ""
}

// Worker masterPool config section, to setup worker masterPool
func (workPoolConfig *WorkerPoolConfiguration) configure(section string, c *goconf.ConfigFile) error {
	if c.HasSection(section) {
		workers, err := c.GetInt(section, "workers")
		if err != nil {
			workPoolConfig.workersCount = DefaultWorkPoolSize
			log.WithFields(log.Fields{"workers": workPoolConfig.workersCount}).Debug("using defaults")
		} else {
			if workers <= 0 || workers > runtime.GOMAXPROCS(-1) {
				workPoolConfig.workersCount = DefaultWorkPoolSize
				log.WithFields(log.Fields{"error": err, "worker": workPoolConfig.workersCount}).Error(fmt.Sprintf("using defaults, workers should be 1 <= workers <= %d", runtime.GOMAXPROCS(-1)))

			} else {
				workPoolConfig.workersCount = workers
			}
		}
		log.WithFields(log.Fields{"workers": workPoolConfig.workersCount}).Info("Worker count")

		queueSize, err := c.GetInt(section, "task_queue_size")
		if err != nil {
			workPoolConfig.queueSize = DefaultTaskQueueSize
		} else {
			if queueSize <= 10 {
				workPoolConfig.queueSize = DefaultTaskQueueSize
			} else {
				workPoolConfig.queueSize = queueSize
			}
		}
	} else {
		workPoolConfig.workersCount = DefaultWorkPoolSize
		workPoolConfig.queueSize = DefaultTaskQueueSize
		log.WithFields(log.Fields{"workers": workPoolConfig.workersCount}).Info("Worker count")
		log.WithFields(log.Fields{"task_queue_size": workPoolConfig.queueSize}).Info("Task Queue Size")
	}
	workPoolCfg = workPoolConfig
	workPoolCfg.masterPool = workerpool.NewTaskManager(workPoolCfg.workersCount, "manager-worker", workPoolCfg.queueSize)
	workPoolCfg.highPriorityPool = workerpool.NewTaskManager(workPoolCfg.workersCount/2, "manager-highPriorityPool-worker", workPoolCfg.queueSize/2)
	workPoolCfg.totalQueueSize = workPoolCfg.queueSize + workPoolCfg.queueSize/2
	workPoolCfg.totalWorkerCount = workPoolCfg.workersCount + workPoolCfg.workersCount/2
	workPoolCfg.masterPool.Start() // will start the task manager // will create a new instance
	workPoolCfg.highPriorityPool.Start()
	return nil
}

// Worker masterPool config section, to setup worker masterPool
func (workPoolConfig *WorkerPoolConfiguration) Close() {
	if workPoolCfg != nil {
		if workPoolCfg.masterPool != nil {
			workPoolCfg.masterPool.Stop()
		}
		if workPoolCfg.highPriorityPool != nil {
			workPoolCfg.highPriorityPool.Stop()
		}
	}
}

func (jb *JobManager) Close() {
	if jb != nil {
		jb.Stop()
	}
}

// configure configures the SQL database connection based on the provided section and config file.
func (sqlDBConfig *SqldbConfiguration) configure(section string, c *goconf.ConfigFile) error {
	var err error
	sqlDBConfig.Server, err = c.GetString(section, "server")
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("missing or invalid server configuration")
		return errors.New("missing or invalid server configuration")
	}

	//TODO: remove mode from the conf's
	if devEnv {
		//UN COMMENT FOR LOCAL DEVELOPMENT
		sqlDBConfig.Port, err = c.GetInt(section, "port")
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("missing or invalid server configuration")
			return errors.New("missing or invalid server configuration")
		}
	}

	sqlDBConfig.Username, err = c.GetString(section, "username")
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("missing or invalid server configuration")
		return errors.New("missing or invalid server configuration")
	}
	sqlDBConfig.Password, err = c.GetRawString(section, "password")
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("missing or invalid server configuration")
		return errors.New("missing or invalid server configuration")
	}
	sqlDBConfig.DatabaseName, err = c.GetString(section, "dbname")
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("missing or invalid server configuration")
		return errors.New("missing or invalid server configuration")
	}
	// Logging default values
	log.WithFields(log.Fields{
		"server":        sqlDBConfig.Server,
		"port":          sqlDBConfig.Port,
		"username":      sqlDBConfig.Username,
		"password":      sqlDBConfig.Password,
		"database_name": sqlDBConfig.DatabaseName,
	}).Info("SQL database configuration")
	sqlDbCfg = sqlDBConfig
	return err
}

// REST/gRPC server configuration
func (restGrpcConfig *RestGrpcConfiguration) configure(section string, c *goconf.ConfigFile) error {
	var err error
	var port string
	var mode string

	mode, err = c.GetString(section, "mode")
	if err != nil {
		log.WithFields(log.Fields{
			"mode": DefaultMode,
		}).Info("gRPC gateway server address not found defaulting")
		mode = DefaultMode
		err = nil
	}
	restGrpcConfig.Mode = mode

	port, err = c.GetString(section, "grpc_endpoint")
	if err != nil {
		log.WithFields(log.Fields{
			"grpc_endpoint": DefaultGrpcServerAddr,
		}).Info("gRPC gateway server address not found defaulting")
		port = DefaultGrpcServerAddr
		err = nil
	}
	restGrpcConfig.grpcPortAddress = port

	port, err = c.GetString(section, "http_endpoint")
	if err != nil {
		log.WithFields(log.Fields{
			"grpc_endpoint": DefaultGrpcServerAddr,
		}).Info("gRPC gateway server address not found defaulting")
		port = DefaultRestServerAddr
		err = nil
	}
	restGrpcConfig.httpPortAddress = port

	serverRetries, err := c.GetInt(section, "server_retries")
	if err != nil {
		restGrpcConfig.retries = DefaultGrpcRestServerRetries
		err = nil //nolint
	} else {
		restGrpcConfig.retries = serverRetries
	}

	log.WithFields(log.Fields{
		"section":       section,
		"http_endpoint": restGrpcConfig.httpPortAddress,
		"grpc_endpoint": restGrpcConfig.grpcPortAddress,
	}).Info("Using HTTP / gRPC Gateway configuration")

	// set the global Cfg  object
	restGrpcCfg = restGrpcConfig
	return err
}

// Start serving Requests Over NATS
func (restGrpcConfig *RestGrpcConfiguration) Start() error {
	var err error

	log.Info("RestGrpcConfiguration start")

	// Start the gRPC server
	go func() {
		for i := 0; i < restGrpcConfig.retries; i++ {
			// create a listener on TCP port
			lis, err := net.Listen("tcp", restGrpcConfig.grpcPortAddress)
			if err != nil {
				log.WithFields(log.Fields{"error": err.Error()}).Error("gRPC server startup failed")
				log.WithFields(log.Fields{"retry": fmt.Sprintf("%d of %d retries", i+1, restGrpcConfig.retries)}).Info("Data Store Service : gRPC Server retry")
				time.Sleep(1 * time.Second)
				continue
			}

			// Creates a new gRPC server and save it to config instance
			s := grpc.NewServer()
			restGrpcConfig.GRPC.Server = s

			//register with GRPC gateway
			protos.RegisterPatientImageServiceServer(s, &PatientImageServiceGRPCServer{})

			log.Info("Starting gRPC server on: " + restGrpcConfig.grpcPortAddress)
			err = s.Serve(lis)
			if err != nil {
				log.WithFields(log.Fields{"error": err.Error(), "grpc port": restGrpcConfig.grpcPortAddress}).Error("Failed starting gRPC server")
				time.Sleep(1 * time.Second)
			}

			if err == nil {
				log.Info("Data Store Service: gRPC Server started")
				break
			} else {
				log.WithFields(log.Fields{"retry": fmt.Sprintf("%d of %d retries", i+1, restGrpcConfig.retries)}).Info("Data Store Service : gRPC Server retry")
			}
			time.Sleep(1 * time.Second)
		}
	}()

	log.Info("RestGrpcConfiguration start 2")

	if err != nil {
		return err
	}
	// Start the REST server gateway
	go func() {
		for i := 0; i < restGrpcConfig.retries; i++ {
			err := startRESTServer(restGrpcConfig)
			if err == nil {
				log.Info("Data Store Service: REST Server started")
				break
			} else {
				log.WithFields(log.Fields{"retry": fmt.Sprintf("%d of %d retries", i+1, restGrpcConfig.retries)}).Info("Cluster-Boot: REST Server retry")
			}
			time.Sleep(1 * time.Second)
		}
	}()
	log.Info("RestGrpcConfiguration start 3", err)
	return err
}

// CustomLogger is a custom logger that satisfies the cors.Logger interface
type CustomLogger struct{}

// Log logs messages from the CORS middleware
func (l *CustomLogger) Log(format string, args ...interface{}) {
	log.Printf("[CORS Middleware] "+format, args...)
}

// Printf forwards the Printf call to log.Printf
func (l *CustomLogger) Printf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func startRESTServer(rgc *RestGrpcConfiguration) (err error) {
	ctx := context.Background()
	conn, cancel := context.WithCancel(ctx)
	defer cancel()
	httpMux := http.NewServeMux()
	gwMux := grpcruntime.NewServeMux(grpcruntime.WithMarshalerOption(grpcruntime.MIMEWildcard, &grpcruntime.JSONPb{}))

	opts := []grpc.DialOption{grpc.WithInsecure()}
	// Update the server details
	rgc.Rest.ConnContext = conn
	rgc.Rest.ConnContext = conn
	rgc.Rest.GRPCMUX = gwMux
	rgc.Rest.GRPCOption = opts
	rgc.Rest.GRPCPort = rgc.grpcPortAddress

	// Register REST servers`
	err = protos.RegisterPatientImageServiceHandlerFromEndpoint(
		rgc.Rest.ConnContext,
		rgc.Rest.GRPCMUX, rgc.Rest.GRPCPort, rgc.Rest.GRPCOption)
	if err != nil {
		return err
	}
	if err != nil {
		log.WithFields(log.Fields{"Error": err, "Server": "Data Store Service"}).Error("Registration service handler from endpoint failed for REST")
	} else {
		log.WithFields(log.Fields{"Server": "Data Store Service"}).Info("Registration of service handler from end point success for REST")
	}

	// Setup CORS middleware
	corsHandler := cors.New(cors.Options{
		//AllowedOrigins: []string{"*"},
		//AllowedOrigins: []string{"http://localhost", "http://localhost:3001"},
		AllowedMethods:  []string{"OPTIONS", "GET", "POST"},
		AllowedOrigins:  []string{"*"},
		AllowOriginFunc: func(origin string) bool { return true },
		//AllowOriginRequestFunc:     nil,
		//AllowOriginVaryRequestFunc: nil,
		//AllowedMethods:             []string{"OPTIONS", "GET", "POST"}, // Ensure POST method is allowed
		AllowedHeaders:       []string{"*"},
		ExposedHeaders:       nil,
		MaxAge:               36000, // 1 hour
		AllowCredentials:     true,
		AllowPrivateNetwork:  true,
		OptionsPassthrough:   true,
		OptionsSuccessStatus: 0,
		Debug:                true,
		Logger:               &CustomLogger{},
	})

	// Listen and Serve http
	if err == nil {
		// Register the gRPC gateway with CORS middleware and set Referrer-Policy header

		///v1/patient/images/upload
		httpMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// TODO: FIX
			w.Header().Set("Referrer-Policy", "*")
			w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
			w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, DELETE, HEAD, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, ResponseType")
			w.Header().Set("Access-Control-Allow-Headers", "DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Content-Transfer-Encoding,Custom-Header-1,X-Accept-Content-Transfer-Encoding,X-Accept-Response-Streaming,X-User-Agent,X-Grpc-Web")

			log.Info("r.Method ....", r.Method)
			log.Info("r.URL ....", r.URL)

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			log.Info("*********ACCEPTED*********")

			// Check if the URL starts with "/v1/patient/images/upload" this is needed as grpc itself does not handle file upload
			if strings.HasPrefix(r.URL.Path, "/v1/patient/images/upload") {
				handleUploadPatientImageHttp(w, r)
				return
			}

			// For other routes, serve through CORS-enabled gRPC gateway
			corsHandler.Handler(gwMux).ServeHTTP(w, r)
		})

		log.Info("Data Store Service: starting HTTP/1.1 REST server on: " + rgc.httpPortAddress)
		err := http.ListenAndServe(rgc.httpPortAddress, httpMux)
		if err != nil {
			log.WithFields(log.Fields{"error": err.Error(), "http port": rgc.httpPortAddress}).Error("Data Store Server : Failed to start REST server")
		}
	}
	return err
}

// Clean up for natStreamCfg
func (restGrpcConfig *RestGrpcConfiguration) Close() {
	//TODO  need to see if we need to do anything else on this
	if restGrpcConfig != nil && restGrpcConfig.Rest.ConnContext != nil {
		log.Info("Data Store Service: closing HTTP/1.1 REST server on: " + restGrpcConfig.Rest.GRPCPort)
		restGrpcConfig.Rest.ConnContext.Done()
	}
	if restGrpcConfig != nil && restGrpcConfig.GRPC.Server != nil {
		log.Info("Data Store  Service: closing gRPC server on: " + restGrpcConfig.grpcPortAddress)
		restGrpcConfig.GRPC.Server.GracefulStop()
	}
}
