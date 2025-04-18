/*
 * @Author: 7erry
 * @Date: 2024-08-31 03:50:13
 * @LastEditTime: 2025-01-07 19:01:54
 * @Description:
 */
package git

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/collector"
	url "github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/url"

	"github.com/stretchr/testify/require"
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
			u := url.ParseURL(test.input)
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

func TestGetURL(t *testing.T) {
	tests := []struct {
		url string
	}{
		{"https://gitee.com/teocloud/teo-docs-search-engine.git"},
		{"https://gitee.com/Open-Brother/pzstudio.git"},
		{"https://gitee.com/mirrors/Proxy-Go.git"},
		{"https://gitcode.com/lovinpanda/DirectX.git"},
		{"https://gitlab.com/Sasha-Zayets/nx-ci-cd.git"},
		{"https://salsa.debian.org/med-team/kmer.git"},
	}
	for n, test := range tests {
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			u := url.ParseURL(test.url)
			r, err := collector.EzCollect(&u)
			if err != nil {
				t.Fatal(err)
			}
			result, err := GetURL(r)
			if err != nil {
				t.Fatal(err)
			}
			require.Equal(t, test.url, result)
		})
	}
}

func TestEco(t *testing.T) {
	tests := []struct {
		input    string
		expected Repo
	}{
		{
			input:    "https://github.com/gin-gonic/gin.git", //* Go
			expected: Repo{},
		},
		{
			input:    "https://github.com/jquery/jquery.git", //* NPM
			expected: Repo{},
		},
		{
			input:    "https://github.com/pallets/flask.git", //* PyPI dependency type not totally solved
			expected: Repo{},
		},
		{
			input:    "https://github.com/serde-rs/json.git", //* Cargo
			expected: Repo{},
		},
		{
			input:    "https://github.com/junit-team/junit4.git", //* Maven
			expected: Repo{},
		},
	}
	for n, test := range tests {
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			u := url.ParseURL(test.input)
			r, err := collector.EzCollect(&u)
			if err != nil {
				t.Fatal(err)
			}
			repo, err := ParseRepo(r)
			if err != nil {
				t.Fatal(err)
			}
			if repo.EcoDeps == nil {
				fmt.Println("Not found")
				return
			}
			for k, v := range repo.EcoDeps {
				if k != nil {
					fmt.Println(*k)
				}
				if v != nil {
					for _, dep := range *v {
						fmt.Println(dep)
					}
				}
			}
		})
	}
}
