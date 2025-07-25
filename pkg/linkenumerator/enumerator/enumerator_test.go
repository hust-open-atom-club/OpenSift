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

func Test_enumerateCargo(t *testing.T) {
	t.Run("Cargo", func(t *testing.T) {
		c := NewCargoEnumerator(1000)
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

func Test_enumerateGO(t *testing.T) {
	t.Run("Go", func(t *testing.T) {
		c := NewGoEnumerator(100)
		c.SetWriter(writer.NewStdOutWriter())
		c.Enumerate()
	})
}

func Test_enumerateHaskell(t *testing.T) {
	t.Run("Haskell", func(t *testing.T) {
		c := NewHaskellEnumerator(500)
		c.SetWriter(writer.NewStdOutWriter())
		c.Enumerate()
	})
}

func Test_enumerateNpm(t *testing.T) {
	t.Run("Npm", func(t *testing.T) {
		c := NewNpmEnumerator(1000)
		c.SetWriter(writer.NewStdOutWriter())
		c.Enumerate()
	})
}

func Test_enumerateNuget(t *testing.T) {
	t.Run("Nuget", func(t *testing.T) {
		c := NewNugetEnumerator(1000)
		c.SetWriter(writer.NewStdOutWriter())
		c.Enumerate()
	})
}

func Test_enumeratePackagist(t *testing.T) {
	t.Run("Packagist", func(t *testing.T) {
		c := NewPackagistEnumerator(500)
		c.SetWriter(writer.NewStdOutWriter())
		c.Enumerate()
	})
}

func Test_enumeratePypi(t *testing.T) {
	t.Run("Pypi", func(t *testing.T) {
		c := NewPypiEnumerator(1000)
		c.SetWriter(writer.NewStdOutWriter())
		c.Enumerate()
	})
}

func Test_enumerateRuby(t *testing.T) {
	t.Run("Ruby", func(t *testing.T) {
		c := NewRubyEnumerator(1000)
		c.SetWriter(writer.NewStdOutWriter())
		c.Enumerate()
	})
}
