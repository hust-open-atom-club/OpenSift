package git

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/HUSTSecLab/OpenSift/pkg/gitfile/collector"
	"github.com/HUSTSecLab/OpenSift/pkg/gitfile/parser/langeco"
	url "github.com/HUSTSecLab/OpenSift/pkg/gitfile/parser/url"
)

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
			input:    "https://github.com/pallets/flask.git", //* PyPI dependency
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
	rTest := NewRepo()
	rTest.EcoDeps = make(map[*langeco.Package]*langeco.Dependencies)
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
			if repo.EcoDeps == nil {
				fmt.Println("Not found")
				return
			}
			for k, v := range repo.EcoDeps {
				if k != nil {
					fmt.Println(*k)
				}
				if v != nil {
					rTest.EcoDeps[k] = v

					for _, dep := range *v {
						fmt.Println(dep)
					}
				}
			}
		})
	}
	led := LangEcoDeps{}
	led.Merge(&rTest)
	for k, v := range rTest.EcoDeps {
		if k != nil {
			fmt.Println(*k)
		}
		if v != nil {

			for _, dep := range *v {
				fmt.Println(dep)
			}
		}
	}
}

func TestMerge(t *testing.T) {
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
			input:    "https://github.com/pallets/flask.git", //* PyPI dependency
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
			u, _ := url.ParseURL(test.input)
			rTest := NewRepo()
			rTest.EcoDeps = make(map[*langeco.Package]*langeco.Dependencies)
			led := NewLangEcoDeps(&rTest)

			r, err := collector.EzCollect(&u)
			if err != nil {
				t.Fatal(err)
			}

			r1, err := ParseRepo(r)
			if err != nil {
				t.Fatal(err)
			}
			r2, err := ParseRepo(r)
			if err != nil {
				t.Fatal(err)
			}

			for k, v := range r1.EcoDeps {
				rTest.EcoDeps[k] = v
			}
			fmt.Println("Original")
			for k, v := range r2.EcoDeps {
				if k != nil {
					fmt.Println(*k)
				}
				if v != nil {
					rTest.EcoDeps[k] = v

					for _, dep := range *v {
						fmt.Println(dep)
					}
				}
			}
			fmt.Println("Doubled")
			for k, v := range rTest.EcoDeps {
				if k != nil {
					fmt.Println(*k)
				}
				if v != nil {
					for _, dep := range *v {
						fmt.Println(dep)
					}
				}
			}
			led.Merge(&rTest)
			fmt.Println("Merged")
			for k, v := range rTest.EcoDeps {
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
