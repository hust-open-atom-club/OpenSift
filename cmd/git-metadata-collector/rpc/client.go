package rpc

import "net/rpc"

type RpcServiceClient struct {
	client *rpc.Client
}

// AddManualTask implements RpcService.
func (r *RpcServiceClient) AddManualTask(req struct{ GitLink string }, resp *struct{}) error {
	return r.client.Call("Collector.AddManualTask", req, resp)
}

// QueryCurrent implements RpcService.
func (r *RpcServiceClient) QueryCurrent(req struct{}, resp *StatusResp) error {
	return r.client.Call("Collector.QueryCurrent", req, resp)
}

// Start implements RpcService.
func (r *RpcServiceClient) Start(req struct{}, resp *struct{}) error {
	return r.client.Call("Collector.Start", req, resp)
}

// Stop implements RpcService.
func (r *RpcServiceClient) Stop(req struct{}, resp *struct{}) error {
	return r.client.Call("Collector.Stop", req, resp)
}

func (r *RpcServiceClient) Close() {
	if r.client != nil {
		r.client.Close()
	}
}

var _ RpcService = (*RpcServiceClient)(nil)

func NewRpcServiceClient(addr string) (*RpcServiceClient, error) {
	client, err := rpc.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &RpcServiceClient{client: client}, nil
}
