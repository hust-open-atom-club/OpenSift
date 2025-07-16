package main

import (
	"encoding/csv"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/HUSTSecLab/OpenSift/pkg/config"
	"github.com/HUSTSecLab/OpenSift/pkg/storage"
	"github.com/HUSTSecLab/OpenSift/pkg/storage/repository"
	"github.com/samber/lo"
	"github.com/spf13/pflag"
)

var DistMap = map[repository.DistPackageTablePrefix]map[string]int{}

var (
	csvDir = pflag.String("CsvDir", "", "csv file directory")
)

func readCsv(csvFile string) ([][]string, error) {
	file, err := os.Open(csvFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	data, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return data, nil

}

func main() {
	config.RegistCommonFlags(pflag.CommandLine)
	config.ParseFlags(pflag.CommandLine)
	ac := storage.GetDefaultAppDatabaseContext()
	files, err := os.ReadDir(*csvDir)
	if err != nil {
		log.Fatalf("read csv directory failed: %v", err)
	}

	for _, file := range files {
		packageInfo := make(map[string]int)
		if !file.IsDir() {
			filePath := *csvDir + "/" + file.Name()
			data, err := readCsv(filePath)
			if err != nil {
				log.Fatalf("read csv file failed: %v", err)
			}
			for _, row := range data {
				value, err := strconv.Atoi(row[1])
				if err != nil {
					log.Fatalf("convert string to int failed: %v", err)
				}
				packageInfo[row[0]] = value
			}
			DistName := strings.Split(file.Name(), "_")[1]
			DistName = strings.Split(DistName, ".")[0]
			switch DistName {
			case "debian":
				if DistMap[repository.DistLinkTablePrefixDebian] == nil {
					DistMap[repository.DistLinkTablePrefixDebian] = packageInfo
				} else {
					Info := DistMap[repository.DistLinkTablePrefixDebian]
					for key, value := range packageInfo {
						if ok := Info[key]; ok != 0 {
							Info[key] += value
						} else {
							Info[key] = value
						}
					}
				}
			case "arch":
				if DistMap[repository.DistLinkTablePrefixArchlinux] == nil {
					DistMap[repository.DistLinkTablePrefixArchlinux] = packageInfo
				} else {
					Info := DistMap[repository.DistLinkTablePrefixArchlinux]
					for key, value := range packageInfo {
						if ok := Info[key]; ok != 0 {
							Info[key] += value
						} else {
							Info[key] = value
						}
					}
				}
			case "homebrew":
				if DistMap[repository.DistLinkTablePrefixHomebrew] == nil {
					DistMap[repository.DistLinkTablePrefixHomebrew] = packageInfo
				} else {
					Info := DistMap[repository.DistLinkTablePrefixHomebrew]
					for key, value := range packageInfo {
						if ok := Info[key]; ok != 0 {
							Info[key] += value
						} else {
							Info[key] = value
						}
					}
				}
			case "gentoo":
				if DistMap[repository.DistLinkTablePrefixGentoo] == nil {
					DistMap[repository.DistLinkTablePrefixGentoo] = packageInfo
				} else {
					Info := DistMap[repository.DistLinkTablePrefixGentoo]
					for key, value := range packageInfo {
						if ok := Info[key]; ok != 0 {
							Info[key] += value
						} else {
							Info[key] = value
						}
					}
				}
			}
		}
	}
	DistPackages := ParseData(ac, DistMap)
	for dist := range DistPackages {
		distRepo := repository.NewDistPackageRepository(ac, dist)
		for _, distPackage := range DistPackages[dist] {
			err := distRepo.InsertOrUpdate(distPackage)
			if err != nil {
				log.Fatalf("insert dist dependencies failed: %v", err)
			}
		}
	}
	for dist := range DistPackages {
		for _, distPackage := range DistPackages[dist] {
			if distPackage.GitLink != nil {
				distRepo := repository.NewDistDependencyRepository(ac)
				DistDepend, _ := distRepo.GetByLink(*distPackage.GitLink, int(ParsePrefix(dist)))
				if DistDepend != nil {
					DistDepend.Downloads_3m = distPackage.Downloads_3m
					err := distRepo.InsertOrUpdate(DistDepend)
					if err != nil {
						log.Fatalf("insert dist dependencies failed: %v", err)
					}
				}
			}
		}
	}

}

func ParseData(ac storage.AppDatabaseContext, DistMap map[repository.DistPackageTablePrefix]map[string]int) map[repository.DistPackageTablePrefix][]*repository.DistPackage {
	var DistPackages = make(map[repository.DistPackageTablePrefix][]*repository.DistPackage)
	for dist := range DistMap {
		DistPackages[dist] = make([]*repository.DistPackage, 0)
		distRepo := repository.NewDistPackageRepository(ac, dist)
		distPackages, err := distRepo.Query()
		if err != nil {
			log.Fatalf("query dist dependencies failed: %v", err)
		}
		for distPackage := range distPackages {
			if ok := DistMap[dist][*distPackage.Package]; ok != 0 {
				distPackage.Downloads_3m = lo.ToPtr(DistMap[dist][*distPackage.Package])
			}
			DistPackages[dist] = append(DistPackages[dist], distPackage)
		}
	}
	return DistPackages
}

func ParsePrefix(prefix repository.DistPackageTablePrefix) repository.DistType {
	switch prefix {
	case repository.DistLinkTablePrefixDebian:
		return repository.Debian
	case repository.DistLinkTablePrefixArchlinux:
		return repository.Arch
	case repository.DistLinkTablePrefixHomebrew:
		return repository.Homebrew
	case repository.DistLinkTablePrefixGentoo:
		return repository.Gentoo
	}
	return repository.Other
}
