///usr/bin/true; exec /usr/bin/env go run "$0" "$@"

package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/HUSTSecLab/criticality_score/pkg/config"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/url"
	"github.com/HUSTSecLab/criticality_score/pkg/logger"
	"github.com/HUSTSecLab/criticality_score/pkg/storage"
	"github.com/bytedance/gopkg/util/gopool"
	"github.com/h2non/filetype"
	"github.com/spf13/pflag"
)

type task struct {
	gitLink string
	version string
	url     string
	output  string
}

func getDir(outputDir, gitLink string) (string, error) {
	u, _ := url.ParseURL(gitLink)

	var pathName string
	splitItems := strings.Split(u.Pathname, "/")

	if len(splitItems) < 2 {
		return "", fmt.Errorf("Illegal pathname")
	}
	orgName := splitItems[1]

	if len(orgName) < 4 {
		pathName = path.Join(strconv.Itoa(len(orgName)), orgName[0:1], orgName)
	} else {
		pathName = path.Join(orgName[0:2], orgName[2:4], orgName)
	}
	pathName = pathName + "/" + path.Join(splitItems[2:]...)

	return path.Join(outputDir, u.Resource, pathName), nil
}

func getExtension(url string, header http.Header, filename string) (string, error) {
	if strings.HasSuffix(url, ".tar.gz") || strings.HasSuffix(url, ".tgz") {
		return "tar.gz", nil
	}
	if strings.HasSuffix(url, ".tar.xz") {
		return "tar.xz", nil
	}
	if strings.HasSuffix(url, ".tar.bz2") {
		return "tar.bz2", nil
	}

	if cd := header.Get("Content-Disposition"); cd != "" {
		if _, params, err := mime.ParseMediaType(cd); err == nil {
			if filename, ok := params["filename"]; ok {
				ext := path.Ext(filename)
				if ext != "" {
					if strings.HasSuffix(filename, ".tar"+ext) {
						return "tar" + ext, nil
					}
					return ext[1:], nil // Remove the leading dot
				}
			}
		}
	}

	if contentType := header.Get("Content-Type"); contentType != "" {
		exts, _ := mime.ExtensionsByType(contentType)
		if len(exts) > 0 {
			return exts[0][1:], nil // Remove the leading dot
		}
	}

	// if not exsits, use filetype to judge
	kind, err := filetype.MatchFile(filename)
	if err != nil {
		return "", err
	}
	return kind.Extension, nil
}

func doTask(outputDir string, t *task) error {
	dirName, err := getDir(outputDir, t.gitLink)
	if err != nil {
		return err
	}

	// download the file
	if err := os.MkdirAll(dirName, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dirName, err)
	}

	httpClient := &http.Client{}
	resp, err := httpClient.Get(t.url)
	if err != nil {
		return err
	}

	splitItems := strings.Split(t.gitLink, "/")
	softwareName := splitItems[len(splitItems)-1]

	filenameWithoutExt := softwareName + "-" + t.version
	if softwareName == "" {
		filenameWithoutExt = t.version
	}

	// pipe resp to file
	fileName := path.Join(dirName, filenameWithoutExt+".downloading")
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	defer resp.Body.Close()
	_, err = io.Copy(f, resp.Body)

	if err != nil {
		return fmt.Errorf("Error writing file %s: %v", fileName, err)
	}

	ext, err := getExtension(t.url, resp.Header, fileName)

	if err != nil {
		return fmt.Errorf("Could not judge extension: %s", err)
	}

	newFileName := path.Join(dirName, filenameWithoutExt+"."+ext)

	// rename file
	if err := os.Rename(fileName, newFileName); err != nil {
		return fmt.Errorf("failed to rename file from %s to %s: %w", fileName, newFileName, err)
	}

	t.output = newFileName

	return nil
}

func main() {
	config.RegistCommonFlags(pflag.CommandLine)

	var flagJobs int32
	var outputDir string
	var wg sync.WaitGroup
	pflag.Int32VarP(&flagJobs, "jobs", "j", 10, "Number of jobs to run concurrently")
	pflag.StringVarP(&outputDir, "output", "o", "", "Output directory")

	config.ParseFlags(pflag.CommandLine)

	if outputDir == "" {
		logger.Fatal("Output directory is required")
	}

	ctx := storage.GetDefaultAppDatabaseContext()
	rows, err := ctx.Query(`SELECT "git_link", "软件版本", "版本下载地址" FROM "__tempdata" WHERE "版本下载地址" is not null`)
	if err != nil {
		panic(err)
	}

	pool := gopool.NewPool("fetch-releases", flagJobs, &gopool.Config{})

	for rows.Next() {
		var t task
		if err := rows.Scan(&t.gitLink, &t.version, &t.url); err != nil {
			logger.Errorf("Error scanning row: %v", err)
		}
		wg.Add(1)
		pool.Go(func() {
			// t.gitLink = "https://github.com/google/snappy"
			// t.url = "https://api.github.com/repos/google/snappy/tarball/1.2.1"
			defer wg.Done()
			l := logger.WithFields(map[string]any{
				"git_link": t.gitLink,
				"url":      t.url,
			})
			l.Infof("Downloading...")
			err := doTask(outputDir, &t)
			if err != nil {
				l.Error(err)
			}
			l.WithFields(map[string]any{
				"output_filename": t.output,
			}).Infof("Success")
		})
	}
	wg.Wait()
}
