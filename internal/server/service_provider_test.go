package server

import (
	"aipc/byze/internal/api/dto"
	"context"
	"reflect"
	"testing"
)

func TestServiceProviderImpl_CreateServiceProvider(t *testing.T) {
	type args struct {
		ctx     context.Context
		request *dto.CreateServiceProviderRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *dto.CreateServiceProviderResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ServiceProviderImpl{}
			got, err := s.CreateServiceProvider(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateServiceProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateServiceProvider() got = %v, want %v", got, tt.want)
			}
		})
	}
}
