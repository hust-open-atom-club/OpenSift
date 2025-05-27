package rpc

import "net/rpc"

type RpcServiceClient struct {
	client *rpc.Client
}

// GetRound implements RpcService.
func (r *RpcServiceClient) GetRound(req GetRoundReq, resp *RoundDTO) error {
	return r.client.Call("Runner.GetRound", req, resp)
}

// StopCurrentRunning implements RpcService.
func (r *RpcServiceClient) StopCurrentRunning(req StopRunningReq, resp *struct{}) error {
	return r.client.Call("Runner.StopCurrentRunning", req, resp)
}

// GetCurrentRoundID implements RpcService.
func (r *RpcServiceClient) GetCurrentRoundID(req struct{}, resp *RoundResp) error {
	return r.client.Call("Runner.GetCurrentRoundID", req, resp)
}

// Start implements RpcService.
func (r *RpcServiceClient) Start(req struct{}, resp *struct{}) error {
	return r.client.Call("Runner.Start", req, resp)
}

// Stop implements RpcService.
func (r *RpcServiceClient) Stop(req struct{}, resp *struct{}) error {
	return r.client.Call("Runner.Stop", req, resp)
}

// StopRoundJob implements RpcService.
func (r *RpcServiceClient) StopRoundJob(req StopRunningReq, resp *struct{}) error {
	return r.client.Call("Runner.StopRoundJob", req, resp)
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
