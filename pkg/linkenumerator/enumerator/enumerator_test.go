package enumerator

import (
	"testing"

	"github.com/HUSTSecLab/OpenSift/pkg/linkenumerator/writer"
)

func Test_enumerateBitbucket(t *testing.T) {
	t.Run("Bitbucket", func(t *testing.T) {
		c := NewBitBucketEnumerator(1000)
		c.SetWriter(writer.NewStdOutWriter())
		c.Enumerate()
	})
}

func Test_enumerateGitlab(t *testing.T) {
	t.Run("Gitlab", func(t *testing.T) {
		c := NewGitlabEnumerator(1000, 4)
		c.SetWriter(writer.NewStdOutWriter())
		c.Enumerate()
	})
}
