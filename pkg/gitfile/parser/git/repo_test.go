/*
 * @Author: 7erry
 * @Date: 2024-08-31 03:50:13
 * @LastEditTime: 2025-01-07 19:01:54
 * @Description:
 */
package git

import (
	"strconv"
	"testing"

	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/collector"
	url "github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/url"
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
