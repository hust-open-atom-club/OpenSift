package lock

import (
	"fmt"
	"io"
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	file, _ := os.Open("cargo.lock")
	defer file.Close()
	data, _ := io.ReadAll(file)
	t.Run("Parse Cargo Lockfile", func(t *testing.T) {
		pkg, deps, err := Parse(string(data))
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(*pkg)
		for index, dep := range *deps {
			fmt.Println(index, dep)
		}
	})
}
