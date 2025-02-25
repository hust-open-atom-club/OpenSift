package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"sync"

	"github.com/HUSTSecLab/criticality_score/cmd/log-parser/parsers"
	"github.com/HUSTSecLab/criticality_score/pkg/config"
	"github.com/HUSTSecLab/criticality_score/pkg/storage"
	"github.com/spf13/pflag"
)

type CountMap map[string]int

var inputPath = pflag.String("i", "", "input log file")
var outputPath = pflag.String("o", "", "output file")
var nginxCombinedRe = regexp.MustCompile(
	`^(\S+) - (.+) \[([^]]+)\] "([^ ]+ )?(.+)( HTTP/[\d.]+)" (\d+) (\d+) "((?:[^\"]|\\.)*)" "((?:[^\"]|\\.)*)" "((?:[^\"]|\\.)*)" - (.+)$`)

// processChunk 处理文件的一个块
func processChunk(lines []string, resultChan chan map[string]CountMap, wg *sync.WaitGroup) {
	defer wg.Done()
	downloadCount := make(map[string]CountMap)

	distParsers := []parsers.Parser{
		parsers.NewDebianParser(nil),
		parsers.NewArchLinuxParser(nil),
		parsers.NewPypiParser(nil),
	}

	for _, line := range lines {
		matches := nginxCombinedRe.FindStringSubmatch(line)
		if matches == nil {
			fmt.Printf("invalid line: %s\n", line)
			continue
		}
		if matches[7] != "200" {
			continue
		}
		info := parsers.MatchInfo{
			Url: matches[5],
			Ua:  matches[11],
		}
		for _, parser := range distParsers {
			if parser.IsMatch(info) {
				err := parser.ParseLine(info.Url)
				if err != nil {
					fmt.Printf("parse line error: %s\n", err)
				}
				continue
			}
		}
	}

	for _, parser := range distParsers {
		downloadCount[parser.Tag()] = parser.GetResult()
	}

	resultChan <- downloadCount
}

func main() {
	config.RegistCommonFlags(pflag.CommandLine)
	config.ParseFlags(pflag.CommandLine)

	db := storage.GetDefaultAppDatabaseContext()
	db.GetDatabaseConnection()

	numCPU := runtime.NumCPU()
	var wg sync.WaitGroup
	var resultWg sync.WaitGroup

	sem := make(chan struct{}, numCPU)
	resultChan := make(chan map[string]CountMap, numCPU)
	downloadCount := make(map[string]CountMap)

	if *inputPath == "" {
		panic("input path is required")
	}

	file, err := os.Open(*inputPath)
	if err != nil {
		panic(err)
	}

	reader := bufio.NewReaderSize(file, 20*1024*1024)
	scanner := bufio.NewScanner(reader)
	var lines []string
	chunkSize := 100000

	resultWg.Add(1)
	go func() {
		defer resultWg.Done()
		for result := range resultChan {
			for dist, countMap := range result {
				if _, ok := downloadCount[dist]; ok {
					for packageName, count := range countMap {
						if _, ok := downloadCount[dist][packageName]; ok {
							downloadCount[dist][packageName] += count
						} else {
							downloadCount[dist][packageName] = count
						}
					}
				} else {
					downloadCount[dist] = countMap
				}
			}
		}
	}()

	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
		if len(lines) == chunkSize {
			// 获取信号量，控制并发
			sem <- struct{}{}
			wg.Add(1)
			go func(lines []string) {
				defer func() {
					// 释放信号量
					<-sem
				}()
				processChunk(lines, resultChan, &wg)
			}(lines)
			lines = nil
		}
	}

	if len(lines) > 0 {
		sem <- struct{}{}
		wg.Add(1)
		go func(lines []string) {
			defer func() {
				<-sem
			}()
			processChunk(lines, resultChan, &wg)
		}(lines)
	}

	wg.Wait()
	close(sem)
	close(resultChan)

	resultWg.Wait()
	file.Close()

	for dist, countMap := range downloadCount {
		resultPath := fmt.Sprintf("%s_%s.csv", *outputPath, dist)
		outputFile, err := os.Create(resultPath)
		if err != nil {
			panic(err)
		}
		defer outputFile.Close()
		writer := bufio.NewWriter(outputFile)
		for packageName, count := range countMap {
			writer.WriteString(fmt.Sprintf("%s,%d\n", packageName, count))
		}
		writer.Flush()
	}

}
