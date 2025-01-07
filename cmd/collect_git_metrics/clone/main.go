/*
* @Date: 2023-11-11 22:44:26
 * @LastEditTime: 2025-01-07 19:13:21
* @Description: Just Clone
*/
package main

import (
	"log"
	"os"
	"sync"
	"time"

	collector "github.com/HUSTSecLab/criticality_score/pkg/collector_git/collector"
	"github.com/HUSTSecLab/criticality_score/pkg/collector_git/logger"
	url "github.com/HUSTSecLab/criticality_score/pkg/collector_git/parser/url"
	"github.com/HUSTSecLab/criticality_score/pkg/collector_git/utils"
	"github.com/bytedance/gopkg/util/gopool"
)

func main() {
	var path string
	if len(os.Args) == 2 {
		path = os.Args[1]
	} else {
		path = ""
	}
	urls, err := utils.GetCSVInput(path)
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
			u := url.ParseURL(input[0])
			_, err := collector.Collect(&u)
			if err != nil {
				logger.Panicf("Cloning %s Failed", input)
			} else {
				logger.Infof("%s Cloned", input)
			}
		})
	}

	wg.Wait()
}
