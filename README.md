

# Patient-Image-Service


## Tech Stack Breakdown

- **Golang**
- **Docker**
- **MinIO**: An object storage solution that mimics AWS S3, providing a reliable and scalable storage option for files and data. Consider as a dummy S3
- **PostgreSQL**
- **Protobuf**: Facilitates streamlined data serialization and communication, making data exchanges more efficient.
- **Prometheus**: Initially considered for system monitoring but currently not in use.


_the below scripts adds default data to minio and postgres_

## To Run locally
**Go to Project terminal- <br />**
export WORKSPACE_DIR=$(pwd) <br />
echo "Workspace set to $WORKSPACE_DIR" <br />
chmod +x ./build/etc/dev/build_all.sh <br />
./build/etc/dev/build_all.sh <br />
./bin/patient_service <br />


## To Run On Docker (can be seen as Prod env)
**Go to Project terminal- <br />**
chmod +x docker_build_all.sh <br />
./docker_build_all.sh <br />


## Current Setup: Patient-Service
The current implementation is a single service implementation called patient-service which talks to PostgreSQL and another sub-service called blob-service.
It exposes both a REST and gRPC endpoint.
Blob-service, currently part of the patient-service.
In the future, we can remove the blob-service from patient-service and it can be a complete separate service having its own database.
Since we have a single service now, i.e., the patient-service, I have kept the PostgreSQL schema simple (per service db)

### Blob Service
The blob service uses AWS SDK 2 for file upload/download/create bucket operations.
It talks to MINIO which is s3 compatible storage.
The blob service is named so because it can be used to upload blobs of any types not just images.
The bucket can be configurable by the client. (There is a createBucket method exposed, currently hardcoding to "mybucket").
This service can be enhanced further for far more S3 operations which are supported by AWS SDK.

### Database Simplicity
Files might require NoSQL because of their nature of having metadata, especially after processing.
The creation and searching of tags, especially with SQL, is limited since we need to create the tag first,
unlike NoSQL where we can assign tags dynamically.
Something like NoSQL/key-value like DynamoDB or something like MongoDB.

### GRPC Gateway Integration
A GRPC gateway is employed to bridge RESTful services with GRPC communication,
enabling both external REST access and internal GRPC messaging.
This dual communication strategy enhances flexibility.
NOTE. introduces complexities in error handling that can be further addressed.

## Future Directions

- **Blob Service Independence**: Blob service can evolve into a standalone entity, equipped with its own database (I can think of nosql right now)
  This move aims to increase the system's modularity and capacity for handling diverse file types.

- **Database Evolution**: Transitioning to a NoSQL database model for blobs would be better to accommodate
  the nuanced requirements of file metadata and tagging, offering a more adaptable framework for data management.

- **GRPC Gateway Refinement**: Efforts to streamline the GRPC gateway's error handling processes.
  Further work is need right now especially for error handling scenario.

##  Current databsae schema

![Screenshot 2024-02-13 at 10.30.29 AM.png](..%2F..%2F..%2F..%2Fvar%2Ffolders%2Fxk%2Fnzbhh1590qngs752x7tjhmxc0000gn%2FT%2FTemporaryItems%2FNSIRD_screencaptureui_bUO7cM%2FScreenshot%202024-02-13%20at%2010.30.29%E2%80%AFAM.png)

##  Current endpoints
Using GRPC gateway as seen in the diagram using grpc-gateway so we can expose our ports as REST and on grpc for internode service communication
drawback is error handling is not smooth, needs some extra work there.


![Screenshot 2024-02-13 at 10.30.51 AM.png](..%2F..%2F..%2F..%2Fvar%2Ffolders%2Fxk%2Fnzbhh1590qngs752x7tjhmxc0000gn%2FT%2FTemporaryItems%2FNSIRD_screencaptureui_P0BEjL%2FScreenshot%202024-02-13%20at%2010.30.51%E2%80%AFAM.png)

##  Alternate consideration
AWS Cloud Development Kit (AWS CDK), AWS Lambda, DynamoDB, and S3.


# Endpoints
Please check the Swagger s3_service.swagger.json  and the s3_service.proto file.
[Link to the swagger file]('./protos/s3_service.swagger.json')


# Asynchronous operations
For asynchronous operations, there's an async_runner (check async_runner file).
File uploads occur asynchronously now instead of using presigned URLs.
The choice depends on the use case. It's preferable not to block the client,
especially for large file uploads. This can be controlled by a flag setting.
This approach might increase the server's workload since files are temporarily stored on the server
instead of being delegated to MinIO/S3. However, essential file processing, such as data extraction using worker nodes, is a popular use case.

# Synchronous operations
File downloads occur synchronously using a presigned URL.


# ---------------------------------------------

## How did you choose your design/architecture and what characteristics did you look for?
mentioned above...  <br />

Future considerations <br />
-Nature of data <br />
-Scalability <br />
-Single point of failure <br />
-loosely coupling  <br />
if we separate the 2 services

## Why did you choose the particular technology/framework/coding language?
Unlike traditional threads, Goroutines are cheap to create and have minimal overhead,
making it practical to use thousands or even millions of Goroutines within a single application.
I work on a system where at a once we are creating more than 20K threads and still see very minimum cpu and memory utilization.
Any sql would do,I have familiarity with postgres

##  If you had another week of prototyping time, what functionality would you want to add?
As discussed above separate the patient-service and make patient-service and blob service

ideally....  <br />
patient-service =>  mysql for patients  <br />
blob-service =>  nosql for file metadata + s3

currently....  <br />
patient-service =>  mysql for patients + blob-service  <br />
blob-service =>  s3

Unit tests <br />
Scale testing  <br />
Defining limits for cpu and memory utilization best and worst cases  <br />
Meta monitor / prometheus counters  <br />
Redis?? couple if cases I have (not implemented) where redis can play a huge role  <br />
Jobs table in db, which will store each jobs status (wrote code see job_manger but does not store in db)  <br />
Events table in db. which will store progress events (wrote code but does not store in db)  <br />
IAM, aws keys storage consideration <br />
DB indexes  <br />
Setup a NATS client and test<br />
Logs in elastic logstash  <br />
UI for all above  <br />

## If we wanted to deploy this to a real hospital setting, what features would we need to add and what other development activities would you want to do before declaring it ready?

Adding fare more REST endpoints required in an enterprise application <br />
Handling scalability using Data replication, CDN, async like kafka, caching
Authentication <br />
Authorization <br />
Dynamic tags <br />
Health checks <br />
Searching. No search implemented currently in any api. <br />
Also enhance the blob-service for more features. AWS sdk helps in that. It can be a service for other services. <br />

And all the above mentioned
