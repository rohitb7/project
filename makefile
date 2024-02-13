# Makefile

# Variables
DOCKER_CLEANUP=docker-prune.sh
MINIO_SETUP=minio_setup.sh
MINIO_CLEANUP=minio_cleanup.sh
POSTGRES_SETUP=postgres_setup.sh
POSTGRES_CLEANUP=postgres_cleanup.sh
POSTGRES_SCHEMA=postgres_schema.sql
BUILD_PROTO=build_proto.sh
BUILD_S3_MANAGER=build_s3_manager.sh
CLEAN_ALL=clean_all.sh
BUILD_ALL=build_all.sh

.PHONY: all build_all build_proto build_s3_manager clean_all docker_cleanup minio_setup minio_cleanup postgres_setup postgres_cleanup pretest

all: build_all

build_all:
	@chmod +x $(BUILD_ALL)
	@./project/$(BUILD_ALL)

build_proto:
	@chmod +x $(BUILD_PROTO)
	@./$(BUILD_PROTO)

build_s3_manager:
	@chmod +x $(BUILD_S3_MANAGER)
	@./$(BUILD_S3_MANAGER)

clean_all:
	@chmod +x $(CLEAN_ALL)
	@./$(CLEAN_ALL)

docker_cleanup:
	@chmod +x $(DOCKER_CLEANUP)
	@./$(DOCKER_CLEANUP)

minio_setup:
	@chmod +x $(MINIO_SETUP)
	@./$(MINIO_SETUP)

minio_cleanup:
	@chmod +x $(MINIO_CLEANUP)
	@./$(MINIO_CLEANUP)

postgres_setup:
	@chmod +x $(POSTGRES_SETUP)
	@./$(POSTGRES_SETUP)

postgres_cleanup:
	@chmod +x $(POSTGRES_CLEANUP)
	@./$(POSTGRES_CLEANUP)
