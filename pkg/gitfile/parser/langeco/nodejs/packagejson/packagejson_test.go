package packagejson

import (
	"fmt"
	"io"
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	file, _ := os.Open("package.json")
	defer file.Close()
	data, _ := io.ReadAll(file)
	t.Run("Parse Package.json", func(t *testing.T) {
		pkg, deps, _ := Parse(string(data))
		fmt.Println(*pkg)
		for index, dep := range *deps {
			fmt.Println(index, dep)
		}
	})
}
