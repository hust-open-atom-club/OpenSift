package maven

import (
	"fmt"
	"io"
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	file, _ := os.Open("pom.xml")
	defer file.Close()
	data, _ := io.ReadAll(file)
	t.Run("Parse Pom.xml", func(t *testing.T) {
		pkg, deps, _ := Parse(string(data))
		fmt.Println(*pkg)
		for index, dep := range *deps {
			fmt.Println(index, dep)
		}
	})
}
