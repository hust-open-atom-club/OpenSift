package rpcserver

import (
	"net"
	"net/rpc"
	"strconv"

	"github.com/HUSTSecLab/criticality_score/cmd/git-metadata-collector/internal/schedule"

	"github.com/HUSTSecLab/criticality_score/cmd/git-metadata-collector/internal/task"
	rpcproto "github.com/HUSTSecLab/criticality_score/cmd/git-metadata-collector/rpc"
	"github.com/HUSTSecLab/criticality_score/pkg/logger"
)

type RpcServiceServer struct{}

// AddManualTask implements RpcService.
func (r *RpcServiceServer) AddManualTask(req struct{ GitLink string }, resp *struct{}) error {
	logger.WithFields(map[string]any{
		"link": req.GitLink,
	}).Info("manual task added")
	schedule.AddManualTask(req.GitLink)
	return nil
}

// QueryCurrent implements RpcService.
func (r *RpcServiceServer) QueryCurrent(req struct{}, resp *rpcproto.StatusResp) error {
	*resp = rpcproto.StatusResp{
		CurrentTasks: task.GetRunningTasks(),
		PendingTasks: schedule.GetPendingTasks(),
		IsRunning:    schedule.IsScheduleRunning(),
	}
	return nil
}

// Start implements RpcService.
func (r *RpcServiceServer) Start(req struct{}, resp *struct{}) error {
	schedule.StartScheduler()
	return nil
}

// Stop implements RpcService.
func (r *RpcServiceServer) Stop(req struct{}, resp *struct{}) error {
	schedule.StopScheduler()
	return nil
}

var _ rpcproto.RpcService = (*RpcServiceServer)(nil)

func RunServer(port int) {
	err := rpc.RegisterName("Collector", new(RpcServiceServer))
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
