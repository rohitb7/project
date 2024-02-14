package main

import (
	"context"
	"fmt"
	protos "www.rvb.com/protos"
)

type PatientImageServiceGRPCServer struct {
}

// UploadPatientImage
// Asynchronous call
// this method stores data in db, and uploads file asynchronously.
// this does not use presignurl as a general use case might be to process the file and its validate
// it also keeps tracks of upload status and its keeps track of its progress (currently logging but not stored in db)
// basically a worker get assinged from a worker pool and do the file upload async
func (s *PatientImageServiceGRPCServer) UploadPatientImage(ctx context.Context, req *protos.UploadPatientImageRequest) (resp *protos.UploadPatientImageResponse, err error) {
	// handle gracefully: thisis handled with http and not grpc
	panic("implement me")
}

// ListPatientImages
// Synchronous call
// this method stores data in db, and gets the presignedurl from minio
// stores relevant tags as well (see the schems)
func (s *PatientImageServiceGRPCServer) ListPatientImages(ctx context.Context, req *protos.ListPatientImagesRequest) (resp *protos.ListPatientImagesResponse, err error) {
	resp = handleRetrievePatientImage(req, ctx)
	if resp.GetResult().GetError() != nil && len(resp.GetResult().GetError().GetMessage()) > 0 {
		return resp, fmt.Errorf("%v", resp.GetResult().GetError().GetMessage())
	}
	return resp, nil
}

func (s *PatientImageServiceGRPCServer) GetPatientImage(ctx context.Context, request *protos.RetrievePatientImageRequest) (*protos.Result, error) {
	//TODO implement me
	panic("implement me")
}

func (s *PatientImageServiceGRPCServer) UpdatePatientImageTags(ctx context.Context, tags *protos.RetrievePatientImageRequestByTags) (*protos.Result, error) {
	//TODO implement me
	panic("implement me")
}

func (s *PatientImageServiceGRPCServer) DeletePatientImage(ctx context.Context, request *protos.DeletePatientImageRequest) (*protos.Result, error) {
	//TODO implement me
	panic("implement me")
}

func (s *PatientImageServiceGRPCServer) RetrieveImagesByTag(ctx context.Context, tags *protos.RetrievePatientImageRequestByTags) (*protos.Result, error) {
	//TODO implement me
	panic("implement me")
}
