/*
 * @Author: 7erry
 * @Date: 2024-09-29 14:41:35
 * @LastEditTime: 2025-03-03 17:47:30
 * @Description:
 */
package collector

import (
	"testing"

	url "github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/url"
	"github.com/stretchr/testify/require"
)

func TestCollect(t *testing.T) {
	tests := []struct {
		input    string
		expected error
	}{
		{input: "https://github.com/gin-gonic/gin", expected: nil},
		{input: "https://gitee.com/teocloud/teo-docs-search-engine.git", expected: nil},
		{input: "https://gitlab.com/Sasha-Zayets/nx-ci-cd.git", expected: nil},
		{input: "https://salsa.debian.org/med-team/kmer.git", expected: nil},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			u, _ := url.ParseURL(test.input)
			_, err := Collect(&u, t.TempDir(), nil)
			require.Equal(t, test.expected, err)
		})
	}
}

func TestEzCollect(t *testing.T) {
	tests := []struct {
		input    string
		expected error
	}{
		{input: "https://github.com/gin-gonic/gin", expected: nil},
		{input: "https://gitee.com/teocloud/teo-docs-search-engine.git", expected: nil},
		{input: "https://gitlab.com/Sasha-Zayets/nx-ci-cd.git", expected: nil},
		{input: "https://salsa.debian.org/med-team/kmer.git", expected: nil},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			u, _ := url.ParseURL(test.input)
			_, err := EzCollect(&u)
			require.Equal(t, test.expected, err)
		})
	}
}

func TestCheck(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"https://github.com/gin-gonic/gin", true},
		{"https://salsa.debian.org/ruby-team/ruby-debian.git", true},
		{"https://github.com/Homebrew/homebrew-core/blob/a8921327fe4f7b6214763769d50d8e6614270089/Formula/p/python-tk@3.10.rb.git", false},
		{"https://sourceforge.net/p/bmagic/code/ci/master/tree/.git", false},
		{"https://github.com/haskell-compat/base-orphans#readme.git", false},
		{"https://github.com/conda-forge/perl-dist-checkconflicts-feedstock?tab=readme-ov-file#readme.git", false},
		{"https://svn.code.sf.net/p/refit/code/trunk refit-code.git", false},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			u, _ := url.ParseURL(test.input)
			require.Equal(t, test.expected, Check(&u) == nil)
		})
	}
}
