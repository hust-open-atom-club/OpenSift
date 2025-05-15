package util

import (
	"encoding/csv"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/url"
)

func GetCSVInput(path string) ([][]string, error) {
	file, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	defer file.Close()
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	urls, err := reader.ReadAll()

	if err != nil {
		return nil, err
	}

	return urls, nil
}

func Save2CSV(outputPath string, content [][]string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.WriteAll(content)

	return nil
}

func GetGitRepositoryPath(storagePath string, u *url.RepoURL) (string, error) {
	pathnames := strings.Split(u.Pathname, "/")
	if len(pathnames) < 2 {
		return "", fmt.Errorf("bad pathname: %s", u.Pathname)
	}

	username := pathnames[1]
	var prefix string
	if len(username) < 4 {
		prefix = strconv.Itoa(len(username)) + "/" + username[0:1]
	} else {
		prefix = username[0:2] + "/" + username[2:4]
	}

	// join path
	return path.Join(storagePath, u.Resource, prefix, u.Pathname), nil
}

func GetGitRepositoryPathFromURL(storagePath string, gitLink string) (string, error) {
	u := url.ParseURL(gitLink)
	return GetGitRepositoryPath(storagePath, &u)
}
