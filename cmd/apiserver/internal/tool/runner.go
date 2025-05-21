package tool

import (
	"encoding/json"
	"os"
	"path"
	"time"

	"github.com/HUSTSecLab/criticality_score/pkg/config"
	"github.com/google/uuid"
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
	ID             string    `json:"id"`
	ToolID         string    `json:"toolId"`
	ToolName       string    `json:"toolName"`
	LaunchUserName string    `json:"launchUserName"`
	StartTime      time.Time `json:"startTime"`
	EndTime        time.Time `json:"endTime"`
	Ret            int       `json:"ret"`
	Err            string    `json:"err"`
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
	metaFileName := path.Join(dir, id+".json")

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
		StartTime:      time.Now(),
	}

	// write meta to file
	d, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return nil, err
	}
	err = os.WriteFile(metaFileName, d, 0644)
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
		meta.EndTime = time.Now()
		meta.Ret = ret
		if err != nil {
			meta.Err = err.Error()
		}

		// write meta to file
		d, err := json.MarshalIndent(meta, "", "  ")
		if err != nil {
			return
		}
		os.WriteFile(metaFileName, d, 0644)
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

func GetHistoryInstances() ([]*ToolInstanceHistory, error) {
	dir := config.GetWebToolHistoryDir()
	files, err := os.ReadDir(dir)

	if err != nil {
		return nil, err
	}

	var instances []*ToolInstanceHistory
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if path.Ext(file.Name()) != ".json" {
			continue
		}
		fileName := path.Join(dir, file.Name())
		data, err := os.ReadFile(fileName)
		if err != nil {
			return nil, err
		}
		var instance ToolInstanceHistory
		err = json.Unmarshal(data, &instance)
		if err != nil {
			return nil, err
		}
		instances = append(instances, &instance)
	}
	return instances, nil
}

func GetLog(id string) (string, error) {
	dir := config.GetWebToolHistoryDir()
	logFileName := path.Join(dir, id+".log")
	logFile, err := os.OpenFile(logFileName, os.O_RDONLY, 0666)
	if err != nil {
		return "", err
	}
	defer logFile.Close()
	data, err := os.ReadFile(logFileName)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (inst *ToolInstance) TerminateInstance() {
	inst.Kill <- 2
}
