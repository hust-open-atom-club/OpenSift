package setup

import (
	"fmt"
	"io"
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	file, _ := os.Open("setup.py")
	defer file.Close()
	data, _ := io.ReadAll(file)
	t.Run("Parse setup.py", func(t *testing.T) {
		pkg, deps, err := Parse(string(data))
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(*pkg)
		if deps != nil {
			for index, dep := range *deps {
				fmt.Println(index, dep)
			}
		}
	})
}
