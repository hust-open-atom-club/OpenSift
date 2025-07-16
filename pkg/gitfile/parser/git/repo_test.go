package git

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/HUSTSecLab/OpenSift/pkg/gitfile/collector"
	url "github.com/HUSTSecLab/OpenSift/pkg/gitfile/parser/url"
)

func TestParseGitRepo(t *testing.T) {
	tests := []struct {
		input    string
		expected Repo
	}{
		{
			input:    "https://github.com/gin-gonic/gin.git",
			expected: Repo{},
		},
		{
			input:    "https://gitee.com/mirrors/Proxy-Go.git",
			expected: Repo{},
		},
		{
			input:    "https://gitlab.com/Sasha-Zayets/nx-ci-cd.git",
			expected: Repo{},
		},
	}
	for n, test := range tests {
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			u, _ := url.ParseURL(test.input)
			r, err := collector.EzCollect(&u)
			if err != nil {
				t.Fatal(err)
			}
			repo, err := ParseRepo(r)
			if err != nil {
				t.Fatal(err)
			}
			repo.Show()
			//require.Equal(t, test.expected, *repo)
		})
	}
}

func TestParseLocalRepo(t *testing.T) {
	tests := []struct {
		input    string
		expected Repo
	}{
		{
			input:    "",
			expected: Repo{},
		},
	}
	for n, test := range tests {
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			r, err := collector.Open(test.input)
			if err != nil {
				t.Fatal(err)
			}
			repo, err := ParseRepo(r)
			if err != nil {
				t.Fatal(err)
			}
			//* repo.Show()
			for k, _ := range repo.EcoDeps {
				if k != nil {
					fmt.Println(*k)
				}
				//if v != nil {
				//	for _, dep := range *v {
				//		fmt.Println(dep)
				//	}
				//}
			}
			//require.Equal(t, test.expected, *repo)
		})
	}

}
