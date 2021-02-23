package transport

import (
	"github.com/ClessLi/bifrost/api/protobuf-spec/bifrostpb"
	"github.com/ClessLi/bifrost/internal/pkg/bifrost/endpoint"
	"golang.org/x/net/context"
	"google.golang.org/grpc/health/grpc_health_v1"
	"time"
)

type watchResponseInfo struct {
	bytesResponseChan chan *bifrostpb.BytesResponse
	signalChan        chan int
}

func (wr watchResponseInfo) Respond() <-chan *bifrostpb.BytesResponse {
	return wr.bytesResponseChan
}

func (wr watchResponseInfo) Close() error {
	select {
	case wr.signalChan <- 9:
		return nil
	case <-time.After(time.Second * 30):
		return ErrWatcherCloseTimeout
	}
}

func newWatchResponseInfo(epWatchResponseInfo endpoint.WatchResponseInfo) *watchResponseInfo {
	responseInfo := &watchResponseInfo{
		bytesResponseChan: make(chan *bifrostpb.BytesResponse),
		signalChan:        make(chan int),
	}
	go func() {
		for {
			select {
			case bytesResponseInfo := <-epWatchResponseInfo.Respond():
				responseInfo.bytesResponseChan <- &bifrostpb.BytesResponse{
					Ret: bytesResponseInfo.Respond(),
					Err: bytesResponseInfo.Error(),
				}
			case signal := <-responseInfo.signalChan:
				if signal == 9 {
					epWatchResponseInfo.Close()
					return
				}
			}
		}
	}()
	return responseInfo
}

func EncodeViewResponse(_ context.Context, r interface{}) (interface{}, error) {
	if resp, ok := r.(endpoint.BytesResponseInfo); ok {
		return &bifrostpb.BytesResponse{
			Ret: resp.Respond(),
			Err: resp.Error(),
		}, nil
	}
	return nil, ErrUnknownResponse
}

func EncodeUpdateResponse(_ context.Context, r interface{}) (interface{}, error) {
	if resp, ok := r.(endpoint.ErrorResponseInfo); ok {
		return &bifrostpb.ErrorResponse{Err: resp.Error()}, nil
	}
	return nil, ErrUnknownResponse
}

func EncodeWatchResponse(_ context.Context, r interface{}) (interface{}, error) {
	if resp, ok := r.(endpoint.WatchResponseInfo); ok {
		return newWatchResponseInfo(resp), nil
	}
	return nil, ErrUnknownResponse
}

func EncodeHealthCheckResponse(_ context.Context, r interface{}) (interface{}, error) {
	resp := r.(endpoint.HealthResponse)
	status := grpc_health_v1.HealthCheckResponse_NOT_SERVING
	if resp.Status {
		status = grpc_health_v1.HealthCheckResponse_SERVING
	}
	return &grpc_health_v1.HealthCheckResponse{
		Status: status,
	}, nil
}
