// This file is used to clone the git repository from the input csv file.
package main

import (
	"log"
	"os"
	"sync"
	"time"

	collector "github.com/HUSTSecLab/criticality_score/pkg/gitfile/collector"
	url "github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/url"
	gitUtil "github.com/HUSTSecLab/criticality_score/pkg/gitfile/util"
	"github.com/HUSTSecLab/criticality_score/pkg/logger"
	"github.com/bytedance/gopkg/util/gopool"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	const viperStorageKey = "storage"

	logger.ConfigAsCommandLineTool()

	pflag.Usage = func() {
		logger.Printf("This tool is used to clone the git repository from the input csv file.\n")
		logger.Printf("Usage: %s [options...] [path]\n", os.Args[0])
		pflag.PrintDefaults()
	}

	pflag.StringP(viperStorageKey, "s", "./storage", "path to git storage location")
	pflag.Parse()
	viper.BindPFlag(viperStorageKey, pflag.Lookup("storage"))
	viper.BindEnv(viperStorageKey, "STORAGE_PATH")

	if pflag.NArg() == 0 || pflag.NArg() > 1 {
		pflag.Usage()
		os.Exit(1)
	}

	path := pflag.Arg(0)

	urls, err := gitUtil.GetCSVInput(path)
	if err != nil {
		log.Fatalf("Failed to read %s", path)
	}
	var wg sync.WaitGroup
	wg.Add(len(urls))

	for index, input := range urls {
		if index%10 == 0 {
			time.Sleep(5 * time.Second)
		} else {
			time.Sleep(2 * time.Second)
		}

		gopool.Go(func() {
			defer wg.Done()
			u, _ := url.ParseURL(input[0])
			_, err := collector.Collect(&u, viper.GetString(viperStorageKey))
			if err != nil {
				logger.Panicf("Cloning %s Failed", input)
			} else {
				logger.Infof("%s Cloned", input)
			}
		})
	}

	wg.Wait()
}
