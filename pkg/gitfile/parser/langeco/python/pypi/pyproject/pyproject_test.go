package pyproject

import (
	"fmt"
	"io"
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	file, _ := os.Open("pyproject.toml")
	defer file.Close()
	data, _ := io.ReadAll(file)
	pkg, deps, _ := Parse(string(data))
	t.Run("Parse Pyproject.toml", func(t *testing.T) {
		fmt.Println(*pkg)
		for index, dep := range *deps {
			fmt.Println(index, dep)
		}
	})
}
