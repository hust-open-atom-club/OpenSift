package tool

import (
	"io"
	"os"
	"path"
	"time"

	"github.com/HUSTSecLab/criticality_score/pkg/config"
	"github.com/HUSTSecLab/criticality_score/pkg/logger"
	"github.com/google/uuid"
	"github.com/samber/lo"
)

type ResizeArg struct {
	Width  uint16
	Height uint16
}

type InstanceResult struct {
	Ret int
	Err error
}

type ToolInstance struct {
	ID     string
	Tool   *Tool
	Args   map[string]any
	Input  *ToolReader
	Output *ToolWriter
	// Error          *ToolWriter
	Resize         chan ResizeArg
	LaunchUserName string
	Result         chan InstanceResult
	Kill           chan int
	StartTime      time.Time
}

var runningInstances = make(map[string]*ToolInstance)

type ToolInstanceHistory struct {
	ID             string
	ToolID         string
	ToolName       string
	LaunchUserName string
	StartTime      *time.Time
	EndTime        *time.Time
	Ret            *int
	Err            *string
	// Not in database
	IsRunning bool
}

func RunningInstanceToHistory(inst *ToolInstance) *ToolInstanceHistory {
	return &ToolInstanceHistory{
		ID:             inst.ID,
		ToolID:         inst.Tool.ID,
		ToolName:       inst.Tool.Name,
		LaunchUserName: inst.LaunchUserName,
		StartTime:      lo.ToPtr(inst.StartTime),
		IsRunning:      true,
	}
}

func CreateAndRun(tool *Tool, args map[string]any, launchUser string) (*ToolInstance, error) {

	// generate a uuid
	uuid, err := uuid.NewRandom()

	if err != nil {
		return nil, err
	}
	id := uuid.String()

	dir := config.GetWebToolHistoryDir()
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, err
	}
	logFileName := path.Join(dir, id+".log")

	logFile, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	inst := &ToolInstance{
		ID:   id,
		Tool: tool,
		Args: args,
		Input: &ToolReader{
			buffer: make(chan []byte),
			// logFile: logFile,
		},
		Result: make(chan InstanceResult, 1),
		Output: &ToolWriter{
			logFile: logFile,
		},
		// Error: &ToolWriter{
		// 	logFile: logFile,
		// },
		Resize:    make(chan ResizeArg),
		Kill:      make(chan int),
		StartTime: time.Now(),
	}

	meta := &ToolInstanceHistory{
		ID:             id,
		ToolID:         tool.ID,
		ToolName:       tool.Name,
		LaunchUserName: launchUser,
		StartTime:      lo.ToPtr(time.Now()),
	}

	err = SaveToolInstanceHistory(meta)
	if err != nil {
		return nil, err
	}

	go func() {
		runningInstances[id] = inst
		defer func() {
			logFile.Close()
			delete(runningInstances, id)
		}()

		ret, err := inst.Tool.Run(inst.Args, inst.Input, inst.Output, inst.Kill, inst.Resize)
		// consume all Kill signal

		allConsumed := false
		for !allConsumed {
			// try to read from the channel
			// if there is no signal, break
			select {
			case <-inst.Kill:
			default:
				allConsumed = true
			}
		}
		close(inst.Kill)

		inst.Result <- struct {
			Ret int
			Err error
		}{ret, err}
		close(inst.Result)
		meta.EndTime = lo.ToPtr(time.Now())
		meta.Ret = lo.ToPtr(ret)
		if err != nil {
			meta.Err = lo.ToPtr(err.Error())
		}

		err = SaveToolInstanceHistory(meta)
		if err != nil {
			logger.Errorf("save tool instance history error: %v", err)
		}
	}()
	return inst, nil
}

func GetRunningInstances() map[string]*ToolInstance {
	return runningInstances
}

func GetRunningInstance(id string) (*ToolInstance, error) {
	inst, ok := runningInstances[id]
	if !ok {
		return nil, os.ErrNotExist
	}
	return inst, nil
}

func GetInstanceHistory(id string) (*ToolInstanceHistory, error) {
	inst, ok := runningInstances[id]
	if ok {
		return RunningInstanceToHistory(inst), nil
	}
	item, err := QueryToolInstanceHistory(id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, os.ErrNotExist
	}
	return item, nil
}

func GetInstanceHistories(running bool, skip, take int) ([]*ToolInstanceHistory, error) {
	if running {
		items := lo.MapToSlice(runningInstances, func(k string, v *ToolInstance) *ToolInstanceHistory {
			return RunningInstanceToHistory(v)
		})
		return lo.Slice(items, skip, skip+take), nil
	} else {
		items, err := QueryToolInstancesHistoryOrderByStartTime(skip, take)
		if err != nil {
			return nil, err
		}
		// set IsRunning
		for _, item := range items {
			if _, ok := runningInstances[item.ID]; ok {
				item.IsRunning = true
			}
		}
		return items, nil
	}
}

func GetLog(id string, all bool) ([]byte, error) {
	dir := config.GetWebToolHistoryDir()
	logFileName := path.Join(dir, id+".log")
	logFile, err := os.OpenFile(logFileName, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer logFile.Close()
	// if all is false and size > 1MB, return last 1MB
	if !all {
		fi, err := logFile.Stat()
		if err != nil {
			return nil, err
		}
		if fi.Size() > 1024*1024 {
			_, err = logFile.Seek(-1024*1024, io.SeekEnd)
			if err != nil {
				return nil, err
			}
			// read last 1MB
			data := make([]byte, 1024*1024)
			_, err = logFile.Read(data)
			if err != nil {
				return nil, err
			}
			return append([]byte("truncated."), data...), nil
		}
	}
	// read all
	data, err := os.ReadFile(logFileName)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (inst *ToolInstance) TerminateInstance(sig int) {
	inst.Kill <- sig
}
