package rpcserver

import (
	"net"
	"net/rpc"
	"strconv"

	"github.com/HUSTSecLab/OpenSift/cmd/workflow-runner/internal/db"
	"github.com/HUSTSecLab/OpenSift/cmd/workflow-runner/internal/loop"
	rpcproto "github.com/HUSTSecLab/OpenSift/cmd/workflow-runner/rpc"
)

type RpcServiceServer struct{}

// StopCurrentRunning implements rpc.RpcService.
func (r *RpcServiceServer) StopCurrentRunning(req rpcproto.StopRunningReq, resp *struct{}) error {
	if req.Type == "kill" {
		loop.StopCurrentJob(true)
	} else {
		loop.StopCurrentJob(false)
	}
	return nil
}

// GetCurrentRoundID implements rpc.RpcService.
func (r *RpcServiceServer) GetCurrentRoundID(req struct{}, resp *rpcproto.RoundResp) error {
	id, err := db.GetMaxRoundID()
	if err != nil {
		return err
	}
	resp.CurrentRound = id
	return nil
}

// GetRound implements rpc.RpcService.
func (r *RpcServiceServer) GetRound(req rpcproto.GetRoundReq, resp *rpcproto.RoundDTO) error {
	round, err := db.GetRound(req.RoundID)
	if err != nil {
		return err
	}
	*resp = *round
	return nil
}

// Start implements rpc.RpcService.
func (r *RpcServiceServer) Start(req struct{}, resp *struct{}) error {
	loop.SetRunningState(true)
	return nil
}

// Stop implements rpc.RpcService.
func (r *RpcServiceServer) Stop(req struct{}, resp *struct{}) error {
	loop.SetRunningState(false)
	return nil
}

var _ rpcproto.RpcService = (*RpcServiceServer)(nil)

func Start(port int) {
	err := rpc.RegisterName("Runner", new(RpcServiceServer))
	if err != nil {
		panic(err)
	}
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeConn(conn)
	}

}
