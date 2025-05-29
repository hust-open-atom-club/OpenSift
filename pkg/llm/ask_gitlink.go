package llm

import (
	"bytes"
	"context"
	"fmt"
	"iter"
	"os"
	"text/template"

	"google.golang.org/genai"
)

const askGitLinkPromptTemplate = `You are an expert in software package management and distribution systems. Your task is to provide the Git repository link for a given package in a specific Linux distribution.

## Package Information

Distribution: {{.Distribution}}
Package Name: {{.PackageName}}
Description: {{.Description}}
Homepage: {{.Homepage}}

## Output Format

Please provide the Git repository link for the package in the following format:
` + "```" + `
Git Link: <link>
Confidence: <confidence>
` + "```" + `

### Details

Git Link should be the most official Git repository link for the package. If you cannot find a Git repository, please set Git Link to "NA". Git Link should be a valid URL starting with "http://" or "https://", or "git://". Links in GitHub and gitlab should have no ".git" suffix, e.g. "https://github/neovim/neovim", while links in other platforms are supposed to keep its original format, such as "git://gcc.gnu.org/git/gcc.git".

Confidence should be a number between 0 and 1, where 0 means no confidence and 1 means high confidence. If you are unsure, provide the best guess based on the available information. If you are certain that the link is correct, or the link does not exist, set Confidence to 1.

## Example

The following is an output example of package neovim in the Ubuntu distribution:

` + "```" + `
Git Link: https://github.com/neovim/neovim
Confidence: 1.0
` + "```" + `

## Important Notes

1. There are many repos for a package, you should find the most official one, pay attention to following common cases:
   - The link is a github mirror, not the original repo, for example, the right official link for ` + "`gcc`" + ` is "git://gcc.gnu.org/git/gcc.git", but not "https://github.com/gcc-mirror/gcc"
   - *Distributions may have their own forks or mirrors*, such as Debian's neovim repo at https://salsa.debian.org/vim-team/neovim, which is a bad answer for it is not the official neovim repo, while the official one is https://github.com/neovim/neovim

2. Some packages may not be under development with git, but SVN, mercurial, or other version control systems. In such cases, set Git Link to "NA" and Confidence to 1.

3. Search websites like Google to find the official Git repository link if you cannot find it in the package's homepage or description. You can use the following search query format: "<package name> official git repo", and try not to use the "<package name> <distribution> git repo" format, as it may lead to wrong results, such as the Debian neovim repo mentioned above.

4. Homepage usually points to the package's official website, which may contain links to the Git repository, it is a good place to start your search.
`

// 3. Search websites like Google to find the official Git repository link if you cannot find it in the package's homepage or description. You can use the following search query format: "<package name> official git repo".

func AskGitLinkPrompt(distribution, packageName, description, homepage string) (iter.Seq2[*genai.GenerateContentResponse, error], error) {
	apiKey := os.Getenv("GEMINI_API_KEY")

	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is not set")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, err
	}

	googleSearchTool := genai.Tool{
		GoogleSearch: &genai.GoogleSearch{},
	}

	askPromptArgs := map[string]interface{}{
		"Distribution": distribution,
		"PackageName":  packageName,
		"Description":  description,
		"Homepage":     homepage,
	}
	// generate from template
	templ, err := template.New("askGitLinkPrompt").Parse(askGitLinkPromptTemplate)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := templ.Execute(&buf, askPromptArgs); err != nil {
		return nil, err
	}
	prompt := buf.String()

	res := client.Models.GenerateContentStream(ctx, "gemini-2.0-flash",
		genai.Text(prompt),
		&genai.GenerateContentConfig{
			// with google search results
			Tools: []*genai.Tool{
				&googleSearchTool,
			},
			ResponseModalities: []string{
				string(genai.ModalityText),
			},
		},
	)
	return res, nil
}
