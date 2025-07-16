package llm_test

import (
	"fmt"
	"testing"

	"github.com/HUSTSecLab/OpenSift/pkg/llm"
)

func TestAskGitLink(t *testing.T) {

	ret, err := llm.AskGitLinkPrompt(
		"Ubuntu",
		"nginx",
		"Nginx is a high-performance HTTP server and reverse proxy, as well as an IMAP/POP3 proxy server.",
		"https://nginx.org/",
	)
	if err != nil {
		t.Fatalf("AskGitLinkPrompt failed: %v", err)
	}

	for resp, err := range ret {
		if err != nil {
			t.Fatalf("Error in response: %v", err)
		}
		if resp == nil {
			t.Fatal("Received nil response")
		}
		// if len(resp.Candidates) == 0 {
		// 	t.Fatal("No choices in response")
		// }
		str, _ := resp.MarshalJSON()
		fmt.Println(string(str))
	}

}
