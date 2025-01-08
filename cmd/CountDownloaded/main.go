/*
 * @Date: 2023-11-11 22:44:26
 * @LastEditTime: 2025-01-07 19:14:21
 * @Description: Just Count downloaded repos
 */
package main

import (
	"fmt"
	"os"

	config "github.com/HUSTSecLab/criticality_score/pkg/collector_git/config"
	url "github.com/HUSTSecLab/criticality_score/pkg/collector_git/parser/url"
	"github.com/HUSTSecLab/criticality_score/pkg/collector_git/utils"
	"github.com/HUSTSecLab/criticality_score/pkg/logger"

	"github.com/go-git/go-git/v5"
)

func main() {
	var path string
	if len(os.Args) == 2 {
		path = os.Args[1]
	} else {
		path = ""
	}
	inputs, err := utils.GetCSVInput(path)
	if err != nil {
		logger.Fatalf("Reading %s Failed", path)
	}
	var count int
	for _, input := range inputs {
		u := url.ParseURL(input[0])
		path = config.STORAGE_PATH + u.Pathname
		_, err := git.PlainOpen(path)
		if err == git.ErrRepositoryNotExists {
			logger.Infof("%s Not Collected", input[0])
			continue
		}
		count++
	}
	fmt.Println(count)
}
