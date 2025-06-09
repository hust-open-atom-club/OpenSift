package git

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/collector"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser"
	"github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/langeco"
	url "github.com/HUSTSecLab/criticality_score/pkg/gitfile/parser/url"
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
					for _, dep := range *v {
						fmt.Println(dep)
					}
				}
			}
		})
	}
}

func TestMerge(t *testing.T) {
	r := NewRepo()
	pkgs := []langeco.Package{
		{
			Name:    "test1",
			Version: "123",
			Eco:     parser.NPM,
		},
		{
			Name:    "test1",
			Version: "123",
			Eco:     parser.NPM,
		},
		{
			Name:    "test2",
			Version: "123",
			Eco:     parser.NPM,
		},
	}
	deps := []langeco.Dependencies{
		{
			{
				Name:    "deps1",
				Version: "123",
				Eco:     parser.NPM,
			},
			{
				Name:    "deps2",
				Version: "123",
				Eco:     parser.NPM,
			},
			{
				Name:    "deps3",
				Version: "123",
				Eco:     parser.NPM,
			},
		},
		{
			{
				Name:    "deps4",
				Version: "123",
				Eco:     parser.NPM,
			},
			{
				Name:    "deps5",
				Version: "123",
				Eco:     parser.NPM,
			},
			{
				Name:    "deps6",
				Version: "123",
				Eco:     parser.NPM,
			},
		},
		{
			{
				Name:    "deps1",
				Version: "123",
				Eco:     parser.NPM,
			},
			{
				Name:    "deps1",
				Version: "123",
				Eco:     parser.NPM,
			},
		},
	}

	r.EcoDeps[&pkgs[0]] = &deps[0]
	r.EcoDeps[&pkgs[1]] = &deps[1]
	r.EcoDeps[&pkgs[2]] = &deps[2]

	led := LangEcoDeps{
		languages: map[string]int64{
			parser.NPM: 0,
		},
	}
	fmt.Printf("%+v", r.EcoDeps)
	led.Merge(&r)
	fmt.Printf("%+v", r.EcoDeps)
}
