package main

import (
	"sync"
	"time"

	"github.com/HUSTSecLab/criticality_score/cmd/git-metadata-collector/internal/schedule"
	"github.com/HUSTSecLab/criticality_score/cmd/git-metadata-collector/internal/task"
	"github.com/HUSTSecLab/criticality_score/cmd/git-metadata-collector/rpc"
	"github.com/HUSTSecLab/criticality_score/pkg/config"
	"github.com/HUSTSecLab/criticality_score/pkg/logger"
	"github.com/spf13/pflag"
)

var flagJobsCount = pflag.IntP("jobs", "j", 256, "jobs count")
var flagRpcPort = pflag.IntP("port", "p", 20324, "rpc server port")
var flagDisableCollect = pflag.Bool("no-collect", false, "if set no, clone only but do not collect git metrics")

func main() {
	config.RegistCommonFlags(pflag.CommandLine)
	config.RegistGitStorageFlags(pflag.CommandLine)
	config.ParseFlags(pflag.CommandLine)
	logger.SetContext("git-metadata-collector")

	go rpc.RunServer(*flagRpcPort)

	// psql.CreateTable(db)

	logger.Infof("Launching %d go routines...", *flagJobsCount)

	schedule.SetFetchOptions(*flagJobsCount*10, *flagJobsCount*2)

	var wg sync.WaitGroup

	for i := 0; i < *flagJobsCount; i++ {
		wg.Add(1)
		go func() {
			cnt := 0
			for {
				t, err := schedule.GetTask()
				if err != nil {
					logger.Fatalf("Failed to get task: %s", err)
				}

				// // begin sleep trick
				if cnt%10 == 0 {
					<-time.After(5 * time.Second)
				} else {
					<-time.After(2 * time.Second)
				}
				cnt++
				// end sleep trick
				task.Collect(t, *flagDisableCollect)

				schedule.FinishTask(t)

			}
		}()
	}
	wg.Wait()
}
