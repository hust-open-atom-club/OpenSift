package url

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseURL(t *testing.T) {
	tests := []struct {
		input    string
		expected RepoURL
		hasError bool
	}{
		{input: "http://ionicabizau.net/blog", expected: RepoURL{
			Protocols: []string{"http"},
			Protocol:  "http",
			Port:      nil,
			Resource:  "ionicabizau.net",
			User:      "",
			Pathname:  "/blog",
			Hash:      "",
			Search:    "",
			URL:       "http://ionicabizau.net/blog",
			Query:     map[string][]string{},
		}},
		{input: "    http://ionicabizau.net/blog   ", expected: RepoURL{
			Protocols: []string{"http"},
			Protocol:  "http",
			Port:      nil,
			Resource:  "ionicabizau.net",
			User:      "",
			Pathname:  "/blog",
			Hash:      "",
			Search:    "",
			URL:       "http://ionicabizau.net/blog",
			Query:     map[string][]string{},
		}},
		{input: "http://domain.com/path/name?foo=bar&bar=42#some-hash", expected: RepoURL{
			Protocols: []string{"http"},
			Protocol:  "http",
			Port:      nil,
			Resource:  "domain.com",
			User:      "",
			Pathname:  "/path/name",
			Hash:      "some-hash",
			Search:    "foo=bar&bar=42",
			URL:       "http://domain.com/path/name?foo=bar&bar=42#some-hash",
			Query:     map[string][]string{"bar": {"42"}, "foo": {"bar"}},
		}},
		{input: "http://domain.com/path/name#some-hash?foo=bar&bar=42", expected: RepoURL{
			Protocols: []string{"http"},
			Protocol:  "http",
			Port:      nil,
			Resource:  "domain.com",
			User:      "",
			Pathname:  "/path/name",
			Hash:      "some-hash?foo=bar&bar=42",
			Search:    "",
			URL:       "http://domain.com/path/name#some-hash?foo=bar&bar=42",
			Query:     map[string][]string{},
		}},
		{input: "http://domain.com/path/name?foo=bar&bar=42#some-hash", expected: RepoURL{
			Protocols: []string{"http"},
			Protocol:  "http",
			Port:      nil,
			Resource:  "domain.com",
			User:      "",
			Pathname:  "/path/name",
			Hash:      "some-hash",
			Search:    "foo=bar&bar=42",
			URL:       "http://domain.com/path/name?foo=bar&bar=42#some-hash",
			Query:     map[string][]string{"bar": {"42"}, "foo": {"bar"}},
		}},
		{input: "git+ssh://git@host.xz/path/name.git", expected: RepoURL{
			Protocols: []string{"git", "ssh"},
			Protocol:  "git",
			Port:      nil,
			Resource:  "host.xz",
			User:      "git",
			Pathname:  "/path/name.git",
			Hash:      "",
			Search:    "",
			URL:       "git+ssh://git@host.xz/path/name.git",
			Query:     map[string][]string{},
		}},
		{input: "http://domain.com/path/name?foo=bar&bar=42#some-hash", expected: RepoURL{
			Protocols: []string{"http"},
			Protocol:  "http",
			Port:      nil,
			Resource:  "domain.com",
			User:      "",
			Pathname:  "/path/name",
			Hash:      "some-hash",
			Search:    "foo=bar&bar=42",
			URL:       "http://domain.com/path/name?foo=bar&bar=42#some-hash",
			Query:     map[string][]string{"bar": {"42"}, "foo": {"bar"}},
		}},
		{input: "git@github.com:IonicaBizau/git-stats.git", expected: RepoURL{
			Protocols: []string{},
			Protocol:  "ssh",
			Port:      nil,
			Resource:  "github.com",
			User:      "git",
			Pathname:  "/IonicaBizau/git-stats.git",
			Hash:      "",
			Search:    "",
			URL:       "git@github.com:IonicaBizau/git-stats.git",
			Query:     map[string][]string{},
		}},
		{input: "/path/to/file.png", expected: RepoURL{
			Protocols: []string{},
			Protocol:  "file",
			Port:      nil,
			Resource:  "",
			User:      "",
			Pathname:  "/path/to/file.png",
			Hash:      "",
			Search:    "",
			URL:       "/path/to/file.png",
			Query:     map[string][]string{},
		}},
		{input: "./path/to/file.png", expected: RepoURL{
			Protocols: []string{},
			Protocol:  "file",
			Port:      nil,
			Resource:  "",
			User:      "",
			Pathname:  "./path/to/file.png",
			Hash:      "",
			Search:    "",
			URL:       "./path/to/file.png",
			Query:     map[string][]string{},
		}},
		{input: "./.path/to/file.png", expected: RepoURL{
			Protocols: []string{},
			Protocol:  "file",
			Port:      nil,
			Resource:  "",
			User:      "",
			Pathname:  "./.path/to/file.png",
			Hash:      "",
			Search:    "",
			URL:       "./.path/to/file.png",
			Query:     map[string][]string{},
		}},
		{input: ".path/to/file.png", expected: RepoURL{
			Protocols: []string{},
			Protocol:  "file",
			Port:      nil,
			Resource:  "",
			User:      "",
			Pathname:  ".path/to/file.png",
			Hash:      "",
			Search:    "",
			URL:       ".path/to/file.png",
			Query:     map[string][]string{},
		}},
		{input: "path/to/file.png", expected: RepoURL{
			Protocols: []string{},
			Protocol:  "file",
			Port:      nil,
			Resource:  "",
			User:      "",
			Pathname:  "path/to/file.png",
			Hash:      "",
			Search:    "",
			URL:       "path/to/file.png",
			Query:     map[string][]string{},
		}},
		{input: "git@github.com:9IonicaBizau/git-stats.git", expected: RepoURL{
			Protocols: []string{},
			Protocol:  "ssh",
			Port:      nil,
			Resource:  "github.com",
			User:      "git",
			Pathname:  "/9IonicaBizau/git-stats.git",
			Hash:      "",
			Search:    "",
			URL:       "git@github.com:9IonicaBizau/git-stats.git",
			Query:     map[string][]string{},
		}},
		{input: "git@github.com:0xABC/git-stats.git", expected: RepoURL{
			Protocols: []string{},
			Protocol:  "ssh",
			Port:      nil,
			Resource:  "github.com",
			User:      "git",
			Pathname:  "/0xABC/git-stats.git",
			Hash:      "",
			Search:    "",
			URL:       "git@github.com:0xABC/git-stats.git",
			Query:     map[string][]string{},
		}},
		{input: "https://attacker.com\\@example.com", expected: RepoURL{
			Protocols: []string{"https"},
			Protocol:  "https",
			Port:      nil,
			Resource:  "attacker.com",
			User:      "",
			Pathname:  "/@example.com",
			Hash:      "",
			Search:    "",
			URL:       "https://attacker.com\\@example.com",
			Query:     map[string][]string{},
		}},
		{input: "jav\r\nascript://%0aalert(1)", expected: RepoURL{
			Protocols: []string{"javascript"},
			Protocol:  "javascript",
			Port:      nil,
			Resource:  "%0aalert(1)",
			User:      "",
			Pathname:  "",
			Hash:      "",
			Search:    "",
			URL:       "javascript://%0aalert(1)",
			Query:     map[string][]string{},
		}},
		{input: "https://abcabc/ /abc", hasError: true},
		{input: "git://????????/abc", hasError: true},
	}

	for n, test := range tests {
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			ret, err := ParseURL(test.input)
			require.True(t, (test.hasError) == (err != nil))
			if !test.hasError {
				require.Equal(t, test.expected, ret)
			}
		})
	}
}

func TestProtocols(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{name: "multiple protocols", input: "git+ssh://git@some-host.com/and-the-path/name", expected: []string{"git", "ssh"}},
		{name: "no protocols", input: "//foo.com/bar.js", expected: []string{}},
		{name: "one protocol", input: "ssh://git@some-host.com/and-the-path/name", expected: []string{"ssh"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, Protocols(test.input), test.expected)
		})
	}
}

func TestIsSsh(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		// Secure Shell Transport Protocol (SSH)
		{"ssh://user@host.xz:port/path/to/repo.git/", true},
		{"ssh://user@host.xz/path/to/repo.git/", true},
		{"ssh://host.xz:port/path/to/repo.git/", true},
		{"ssh://host.xz/path/to/repo.git/", true},
		{"ssh://user@host.xz/path/to/repo.git/", true},
		{"ssh://host.xz/path/to/repo.git/", true},
		{"ssh://user@host.xz/~user/path/to/repo.git/", true},
		{"ssh://host.xz/~user/path/to/repo.git/", true},
		{"ssh://user@host.xz/~/path/to/repo.git", true},
		{"ssh://host.xz/~/path/to/repo.git", true},
		{"user@host.xz:/path/to/repo.git/", true},
		{"user@host.xz:~user/path/to/repo.git/", true},
		{"user@host.xz:path/to/repo.git", true},
		{"host.xz:/path/to/repo.git/", true},
		{"host.xz:path/to/repo.git", true},
		{"host.xz:~user/path/to/repo.git/", true},
		{"rsync://host.xz/path/to/repo.git/", true},

		// Git Transport Protocol
		{"git://host.xz/path/to/repo.git/", false},
		{"git://host.xz/~user/path/to/repo.git/", false},

		// HTTP/S Transport Protocol
		{"http://host.xz/path/to/repo.git/", false},
		{"https://host.xz/path/to/repo.git/", false},
		{"http://host.xz:8000/path/to/repo.git/", false},
		{"https://host.xz:8000/path/to/repo.git/", false},
		// Local (Filesystem) Transport Protocol
		{"/path/to/repo.git/", false},
		{"path/to/repo.git/", false},
		{"~/path/to/repo.git", false},
		{"file:///path/to/repo.git/", false},
		{"file://~/path/to/repo.git/", false},
	}

	for n, test := range tests {
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			require.Equal(t, test.expected, IsSsh(test.input))
		})
	}
}
