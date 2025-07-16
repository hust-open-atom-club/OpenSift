package llm

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/HUSTSecLab/OpenSift/pkg/config"
	"github.com/HUSTSecLab/OpenSift/pkg/storage"
	"github.com/PuerkitoBio/goquery"
)

type RepoInfo struct {
	Description string   `json:"description"`
	Topics      []string `json:"topics"`
}

func IndustryID(ctx storage.AppDatabaseContext, url string, batchSize int, outputCsv string) {
	db, _ := ctx.GetDatabaseConnection()
	if db == nil {
		return
	}
	defer db.Close()
	file, err := os.OpenFile(outputCsv, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	gitlinks := fetchGitLink(db)
	gitHubToken := config.GetGithubToken()
	gitIndustry := make(map[string]string)
	for _, gitLink := range gitlinks {
		repoInfo := fetchDesTopic(gitLink, gitHubToken)
		readme := getReadmeText(gitLink)
		var prompt string
		if repoInfo == nil {
			prompt = fmt.Sprintf(PROMPT["industry_idx"], gitLink, readme, "", "")
		} else {
			prompt = fmt.Sprintf(PROMPT["industry_idx"], gitLink, readme, repoInfo.Description, strings.Join(repoInfo.Topics, ","))
		}
		fullResponse := InvokeModel(prompt, url)
		gitIndustry[gitLink] = fullResponse
		if err := writer.Write([]string{gitLink, fullResponse}); err != nil {
			log.Fatal("Error writing to CSV:", err)
		}
		writer.Flush()
		log.Println("gitLink:", gitLink, "industry:", fullResponse)
	}
	err = UpdateIdxBatch(db, batchSize, gitIndustry)
	if err != nil {
		log.Printf("Failed to insert batch: %v", err)
	}
	return
}
func fetchGitLink(db *sql.DB) []string {
	var gitLink []string
	rows, err := db.Query("SELECT git_link FROM git_repositories")
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var link string
		if err := rows.Scan(&link); err != nil {
			panic(err)
		}
		gitLink = append(gitLink, link)
	}
	return gitLink
}

func fetchDesTopic(gitLink string, GitHubToken string) *RepoInfo {
	var owner, repo string
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	if strings.HasSuffix(gitLink, ".git") {
		gitLink = strings.TrimSuffix(gitLink, ".git")
	}
	re := regexp.MustCompile(`^(https?://|git://)?(www\.)?`)
	gitLink = re.ReplaceAllString(gitLink, "")
	parts := strings.Split(gitLink, "/")
	if len(parts) == 3 {
		owner = parts[1]
		repo = parts[2]
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return nil
	}
	req.Header.Set("Accept", "application/vnd.github.mercy-preview+json")
	req.Header.Set("Authorization", "token "+GitHubToken)
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("Request failed with status: %s", resp.Status)
		return nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %s", err)
		return nil
	}
	var repoInfo RepoInfo
	err = json.Unmarshal(body, &repoInfo)
	if err != nil {
		log.Println("Error unmarshaling JSON:", err)
		return nil
	}
	if repoInfo.Description == "" {
		return nil
	}
	if len(repoInfo.Topics) == 0 {
		return nil
	}
	return &repoInfo
}

func getReadmeText(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Failed to fetch page: %v", err)
		return ""
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("Failed to parse HTML: %v", err)
		return ""
	}

	readmeLink := ""
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists && regexp.MustCompile(`^/.*/blob/.*?/README.*$`).MatchString(href) {
			readmeLink = "https://raw.githubusercontent.com" + href
			return
		}
	})

	if readmeLink != "" {
		readmeResp, err := http.Get(strings.Replace(readmeLink, "/blob/", "/", 1))
		if err != nil {
			log.Printf("Failed to fetch README: %v", err)
			return ""
		}
		defer readmeResp.Body.Close()

		if readmeResp.StatusCode == http.StatusOK {
			readmeText, err := ioutil.ReadAll(readmeResp.Body)
			if err != nil {
				log.Printf("Failed to read README content: %v", err)
				return ""
			}

			maxLen := 1500
			if len(readmeText) > maxLen {
				readmeText = readmeText[:maxLen]
			}

			return string(readmeText)
		} else {
			log.Printf("Failed to fetch README, status code: %d", readmeResp.StatusCode)
			return ""
		}
	} else {
		log.Println("README link not found.")
		return ""
	}
}

func UpdateIdxBatch(db *sql.DB, batchSize int, gitIndustry map[string]string) error {
	var gitIndustryList []struct {
		GitLink    string
		IndustryID string
	}
	for gitLink, industryID := range gitIndustry {
		gitIndustryList = append(gitIndustryList, struct {
			GitLink    string
			IndustryID string
		}{GitLink: gitLink, IndustryID: industryID})
	}
	for i := 0; i < len(gitIndustryList); i += batchSize {
		end := i + batchSize
		if end > len(gitIndustryList) {
			end = len(gitIndustryList)
		}

		query := "UPDATE git_repositories SET industry = CASE "
		valueArgs := make([]interface{}, 0, 2*(end-i))

		for idx, item := range gitIndustryList[i:end] {
			query += fmt.Sprintf("WHEN git_link = $%d THEN $%d ", 2*idx+1, 2*idx+2)
			valueArgs = append(valueArgs, item.GitLink, item.IndustryID)
		}

		query += "END WHERE git_link IN ("
		valueStrings := make([]string, 0, end-i)
		for idx := range gitIndustryList[i:end] {
			valueStrings = append(valueStrings, fmt.Sprintf("$%d", 2*idx+1))
		}
		query += strings.Join(valueStrings, ",") + ")"

		_, err := db.Exec(query, valueArgs...)
		if err != nil {
			return err
		}
	}

	return nil
}
