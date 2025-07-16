package llm

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/HUSTSecLab/OpenSift/pkg/storage"
)

var visitedLinks = make(map[string]bool)

func Home2git(ctx storage.AppDatabaseContext, repolist []string, url string, batchSize int, outputCsv string) {
	db, err := ctx.GetDatabaseConnection()
	if err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		return
	}
	if url == "" {
		fmt.Println("Please provide a LLM URL.")
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
	defer writer.Flush()
	fmt.Println("Starting to process the links...")
	resultMap := make(map[string]map[string]string)
	PackageListMap, _ := FetchAllLinks(db, repolist)
	for repo := range PackageListMap {
		resultMap[repo] = make(map[string]string)
		PackageList := PackageListMap[repo]
		for i := 0; i < len(PackageList); i++ {
			homepageURL := PackageList[i][1]
			packageName := PackageList[i][0]
			htmlContent, err := DownloadHTML(homepageURL)
			if err != nil {
				continue
			}

			links, _ := FindLinksInHTML(homepageURL, htmlContent, 1)
			githubURL := ProcessHomepage(packageName, links, homepageURL, url)
			res := Check(githubURL)
			if res == "" {
				continue
			}
			resultMap[repo][packageName] = res
			if err := writer.Write([]string{repo, packageName, res}); err != nil {
				fmt.Println("Error writing to CSV:", err)
			}
			writer.Flush()
			fmt.Println("repo:", repo, "packageName:", packageName, "res:", res)
		}
	}
	UpdateBatch(db, batchSize, resultMap)
}
func Check(githubURL string) string {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(githubURL)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return githubURL
	}
	return ""
}

func FetchAllLinks(db *sql.DB, repolist []string) (map[string][][]string, error) {
	links := make(map[string][][]string)
	var query string
	for _, repo := range repolist {
		query = fmt.Sprintf("SELECT package, homepage FROM %s_packages WHERE (git_link = '' or git_link = NULL) and homepage != ''", repo)
		rows, err := db.Query(query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var packageName, link string
			if err := rows.Scan(&packageName, &link); err != nil {
				return nil, err
			}
			links[repo] = append(links[repo], []string{packageName, link})
		}
	}
	return links, nil
}

func DownloadHTML(url string) (string, error) {
	visitedLinks[url] = true

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP request failed with status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func FindLinksInHTML(url string, htmlContent string, depth int) ([]string, error) {
	// if depth == 0 {
	// 	return nil, nil
	// }
	// doc, err := html.Parse(strings.NewReader(htmlContent))
	// if err != nil {
	// 	return nil, fmt.Errorf("Error parsing HTML: %v", err)
	// }

	var links []string
	// var currentDomain = strings.Split(url, "/")[2] // Get the domain from the URL
	for _, pattern := range gitLinkPatterns {
		if matches := pattern.FindAllString(htmlContent, -1); matches != nil {
			for _, match := range matches {
				if !visitedLinks[match] {
					visitedLinks[match] = true
					links = append(links, match)
				}
			}
		}
	}
	// var traverse func(*html.Node)
	// traverse = func(n *html.Node) {
	// 	if n.Type == html.TextNode {
	// 		for _, pattern := range gitLinkPatterns {
	// 			if matches := pattern.FindAllString(htmlContent, -1); matches != nil {
	// 				for _, match := range matches {
	// 					if !visitedLinks[match] {
	// 						visitedLinks[match] = true
	// 						links = append(links, match)
	// 					}
	// 				}
	// 			}
	// 		}
	// 	}
	// 	for c := n.FirstChild; c != nil; c = c.NextSibling {
	// 		traverse(c)
	// 	}
	// 	if n.Type == html.ElementNode && n.Data == "a" {
	// 		for _, a := range n.Attr {
	// 			if a.Key == "href" {
	// 				href := strings.TrimSpace(a.Val)
	// 				if strings.HasPrefix(href, "/") {
	// 					href = strings.Split(url, "/")[0] + "//" + currentDomain + href
	// 				} else if !strings.HasPrefix(href, "http://") &&
	// 					!strings.HasPrefix(href, "https://") &&
	// 					!strings.HasPrefix(href, "//") {
	// 					href = url + href
	// 				} else {
	// 					linkDomain := strings.Split(href, "/")[2]
	// 					if linkDomain != currentDomain {
	// 						continue // Skip non-same-origin full URLs
	// 					}
	// 				}
	// 				if !visitedLinks[href] {
	// 					visitedLinks[href] = true
	// 					deeperContent, err := DownloadHTML(href)
	// 					if err == nil && depth > 1 {
	// 						FindLinksInHTML(href, deeperContent, depth-1)
	// 					}
	// 				}
	// 			}
	// 		}
	// 	}
	// }

	// traverse(doc)
	return links, nil
}

func FindGitRepository(homepageURL string, links []string, packageName string, attempts int, url string) string {
	var prompt string
	linksString := strings.Join(links, ", ")
	if len(links) > 0 {
		prompt = fmt.Sprintf(PROMPT["home2git_link"], linksString, packageName)
	} else {
		prompt = fmt.Sprintf(PROMPT["home2git_nolink"], homepageURL)
	}
	fullResponse := InvokeModel(prompt, url)
	if strings.Contains(fullResponse, "does not exist") {
		if attempts < 3 {
			return FindGitRepository(homepageURL, links, packageName, attempts+1, url)
		}
		return "does not exist"
	} else if strings.Contains(fullResponse, "URL is:") {
		potentialURL := strings.Split(strings.Split(fullResponse, "URL is:")[1], "\n")[0]
		patten := gitLinkPatterns[0]
		if matches := patten.FindStringSubmatch(potentialURL); len(matches) > 0 {
			return matches[0]
		}
		return potentialURL
	}
	return "does not exist"
}

func ProcessHomepage(packageName string, links []string, homepageURL string, url string) string {
	githubURL := FindGitRepository(homepageURL, links, packageName, 3, url)
	return githubURL
}

func UpdateBatch(db *sql.DB, batchSize int, resultMap map[string]map[string]string) error {
	for repo, packages := range resultMap {
		var updateList []struct {
			PackageName string
			GitLink     string
		}

		for packageName, gitLink := range packages {
			updateList = append(updateList, struct {
				PackageName string
				GitLink     string
			}{PackageName: packageName, GitLink: gitLink})
		}

		for i := 0; i < len(updateList); i += batchSize {
			end := i + batchSize
			if end > len(updateList) {
				end = len(updateList)
			}

			query := "UPDATE " + repo + " SET git_link = CASE "
			valueArgs := make([]interface{}, 0, 2*(end-i))

			for idx, item := range updateList[i:end] {
				query += fmt.Sprintf("WHEN package = $%d THEN $%d ", 2*idx+1, 2*idx+2)
				valueArgs = append(valueArgs, item.PackageName, item.GitLink)
			}

			query += "END WHERE package IN ("
			valueStrings := make([]string, 0, end-i)
			for idx := range updateList[i:end] {
				valueStrings = append(valueStrings, fmt.Sprintf("$%d", 2*idx+1))
			}
			query += strings.Join(valueStrings, ",") + ")"

			valueArgs = append([]interface{}{repo}, valueArgs...)

			_, err := db.Exec(query, valueArgs...)
			if err != nil {
				return fmt.Errorf("failed to execute batch update for repo %s: %v", repo, err)
			}
		}
	}

	return nil
}
