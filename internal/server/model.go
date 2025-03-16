package server

import (
	"context"
	"errors"
	"log/slog"

	"intel.com/aog/internal/api/dto"
	"intel.com/aog/internal/datastore"
	"intel.com/aog/internal/provider"
	"intel.com/aog/internal/types"
	"intel.com/aog/internal/utils/bcode"
)

type Model interface {
	CreateModel(ctx context.Context, request *dto.CreateModelRequest) (*dto.CreateModelResponse, error)
	DeleteModel(ctx context.Context, request *dto.DeleteModelRequest) (*dto.DeleteModelResponse, error)
	GetModels(ctx context.Context, request *dto.GetModelsRequest) (*dto.GetModelsResponse, error)
}

type ModelImpl struct{}

func NewModel() Model {
	return &ModelImpl{}
}

func (s *ModelImpl) CreateModel(ctx context.Context, request *dto.CreateModelRequest) (*dto.CreateModelResponse, error) {
	ds := datastore.GetDefaultDatastore()
	sp := new(types.ServiceProvider)
	if request.ProviderName != "" {
		sp.ProviderName = request.ProviderName
	} else {
		// 获取service 默认provider
		// todo 暂只支持chat、generate服务拉取模型
		if request.ServiceName != types.ServiceChat && request.ServiceName != types.ServiceGenerate {
			return nil, bcode.ErrServer
		}

		service := &types.Service{}
		service.Name = request.ServiceName

		err := ds.Get(ctx, service)
		if err != nil {
			return nil, err
		}

		if request.ServiceSource == types.ServiceSourceLocal && service.LocalProvider != "" {
			sp.ProviderName = service.LocalProvider
		} else if request.ServiceSource == types.ServiceSourceRemote && service.RemoteProvider != "" {
			sp.ProviderName = service.RemoteProvider
		}
	}

	sp.ServiceName = request.ServiceName
	sp.ServiceSource = request.ServiceSource

	err := ds.Get(ctx, sp)
	if err != nil && !errors.Is(err, datastore.ErrEntityInvalid) {
		// todo debug log output
		return nil, bcode.ErrServer
	} else if errors.Is(err, datastore.ErrEntityInvalid) {
		return nil, bcode.ErrServiceRecordNotFound
	}

	m := new(types.Model)
	m.ProviderName = sp.ProviderName
	m.ModelName = request.ModelName

	err = ds.Get(ctx, m)
	if err != nil && !errors.Is(err, datastore.ErrEntityInvalid) {
		// todo debug log output
		return nil, bcode.ErrServer
	}

	modelEngine := provider.GetModelEngine(sp.Flavor)
	stream := false
	pullReq := &types.PullModelRequest{
		Model:  request.ModelName,
		Stream: &stream,
	}
	_, err = modelEngine.PullModel(ctx, pullReq, nil)
	if err != nil {
		slog.Error("Pull model error: ", err.Error())
		return nil, bcode.ErrEnginePullModel
	}
	slog.Info("Pull model %s completed ..." + request.ModelName)

	m.Status = "downloaded"
	err = ds.Add(ctx, m)
	if err != nil {
		return nil, bcode.ErrAddModel
	}

	return &dto.CreateModelResponse{
		Bcode: *bcode.ModelCode,
	}, nil
}

func (s *ModelImpl) DeleteModel(ctx context.Context, request *dto.DeleteModelRequest) (*dto.DeleteModelResponse, error) {
	ds := datastore.GetDefaultDatastore()
	sp := new(types.ServiceProvider)
	sp.ProviderName = request.ProviderName

	err := ds.Get(ctx, sp)
	if err != nil && !errors.Is(err, datastore.ErrEntityInvalid) {
		// todo err debug log output
		return nil, bcode.ErrServer
	} else if errors.Is(err, datastore.ErrEntityInvalid) {
		return nil, bcode.ErrServiceRecordNotFound
	}

	m := new(types.Model)
	m.ProviderName = request.ProviderName
	m.ModelName = request.ModelName

	err = ds.Get(ctx, m)
	if err != nil && !errors.Is(err, datastore.ErrEntityInvalid) {
		// todo err debug log output
		return nil, bcode.ErrServer
	} else if errors.Is(err, datastore.ErrEntityInvalid) {
		return nil, bcode.ErrModelRecordNotFound
	}

	// 3. 调用engin delete model
	modelEngine := provider.GetModelEngine(sp.Flavor)
	deleteReq := &types.DeleteRequest{
		Model: request.ModelName,
	}

	err = modelEngine.DeleteModel(ctx, deleteReq)
	if err != nil {
		// todo err debug log output
		return nil, bcode.ErrEngineDeleteModel
	}

	err = ds.Delete(ctx, m)
	if err != nil {
		// todo err debug log output
		return nil, err
	}
	return &dto.DeleteModelResponse{
		Bcode: *bcode.ModelCode,
	}, nil
}

func (s *ModelImpl) GetModels(ctx context.Context, request *dto.GetModelsRequest) (*dto.GetModelsResponse, error) {
	ds := datastore.GetDefaultDatastore()
	m := &types.Model{}
	if request.ModelName != "" {
		m.ModelName = request.ModelName
	}
	if request.ProviderName != "" {
		m.ProviderName = request.ProviderName
	}
	list, err := ds.List(ctx, m, &datastore.ListOptions{
		Page:     0,
		PageSize: 1000,
	})
	if err != nil {
		return nil, err
	}

	respData := make([]dto.Model, 0)
	for _, v := range list {
		tmp := new(dto.Model)
		dsModel := v.(*types.Model)

		tmp.ModelName = dsModel.ModelName
		tmp.ProviderName = dsModel.ProviderName
		tmp.Status = dsModel.Status
		tmp.CreatedAt = dsModel.CreatedAt
		tmp.UpdatedAt = dsModel.UpdatedAt

		respData = append(respData, *tmp)
	}

	return &dto.GetModelsResponse{
		Bcode: *bcode.ModelCode,
		Data:  respData,
	}, nil
}
