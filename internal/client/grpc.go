package client

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"intel.com/aog/internal/client/grpc/grpc_client"
	"intel.com/aog/internal/types"
)

type GRPCClient struct {
	conn *grpc.ClientConn
}

func NewGRPCClient(target string) (*GRPCClient, error) {
	conn, err := grpc.Dial(target, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return &GRPCClient{conn: conn},
		nil
}

func (c *GRPCClient) Close() error {
	return c.conn.Close()
}

func (c *GRPCClient) ServerLive() (*grpc_client.ServerLiveResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client := grpc_client.NewGRPCInferenceServiceClient(c.conn)
	serverLiveRequest := grpc_client.ServerLiveRequest{}
	serverLiveResponse, err := client.ServerLive(ctx, &serverLiveRequest)
	if err != nil {
		return nil, err
	}
	return serverLiveResponse, nil
}

func (c *GRPCClient) ServerReady() (*grpc_client.ServerReadyResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client := grpc_client.NewGRPCInferenceServiceClient(c.conn)
	serverReadyRequest := grpc_client.ServerReadyRequest{}
	serverReadyResponse, err := client.ServerReady(ctx, &serverReadyRequest)
	if err != nil {
		return nil, err
	}
	return serverReadyResponse, nil
}

func (c *GRPCClient) ModelMetadata(modelName string, modelVersion string) (*grpc_client.ModelMetadataResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client := grpc_client.NewGRPCInferenceServiceClient(c.conn)
	modelMetadataRequest := grpc_client.ModelMetadataRequest{
		Name:    modelName,
		Version: modelVersion,
	}
	modelMetadataResponse, err := client.ModelMetadata(ctx, &modelMetadataRequest)
	if err != nil {
		return nil, err
	}
	return modelMetadataResponse, nil
}

func (c *GRPCClient) ModelInfer(inferInputs []*grpc_client.ModelInferRequest_InferInputTensor, serviceType, modelName, modelVersion string) (*grpc_client.ModelInferResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	client := grpc_client.NewGRPCInferenceServiceClient(c.conn)

	// Set different Outputs according to different serviceType.
	inferOutputs := make([]*grpc_client.ModelInferRequest_InferRequestedOutputTensor, 0)
	if serviceType == types.ServiceTextToImage {
		inferOutputs = append(inferOutputs, &grpc_client.ModelInferRequest_InferRequestedOutputTensor{
			Name: "image",
		})
	}

	modelInferRequest := grpc_client.ModelInferRequest{
		ModelName: modelName,
		Inputs:    inferInputs,
		Outputs:   inferOutputs,
	}

	modelInferResponse, err := client.ModelInfer(ctx, &modelInferRequest)
	if err != nil {
		return nil, err
	}

	return modelInferResponse, nil
}
