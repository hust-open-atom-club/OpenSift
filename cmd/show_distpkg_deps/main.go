package main

import (
	"flag"
	"fmt"

	"github.com/HUSTSecLab/criticality_score/pkg/collector/alpine"
	"github.com/HUSTSecLab/criticality_score/pkg/collector/archlinux"
	"github.com/HUSTSecLab/criticality_score/pkg/collector/aur"
	"github.com/HUSTSecLab/criticality_score/pkg/collector/centos"
	"github.com/HUSTSecLab/criticality_score/pkg/collector/debian"
	"github.com/HUSTSecLab/criticality_score/pkg/collector/deepin"
	"github.com/HUSTSecLab/criticality_score/pkg/collector/fedora"
	"github.com/HUSTSecLab/criticality_score/pkg/collector/gentoo"
	"github.com/HUSTSecLab/criticality_score/pkg/collector/homebrew"
	"github.com/HUSTSecLab/criticality_score/pkg/collector/nix"
	"github.com/HUSTSecLab/criticality_score/pkg/collector/ubuntu"
	"github.com/HUSTSecLab/criticality_score/pkg/storage"
)

var (
	flagConfigPath = flag.String("config", "config.json", "path to the config file")
	flagType       = flag.String("type", "", "type of the distribution")
	flagGenDot     = flag.String("gendot", "", "output dot file")
	workerCount    = flag.Int("worker", 1, "number of workers")
	batchSize      = flag.Int("batch", 1000, "batch size")
)

func main() {
	flag.Parse()
	storage.InitializeDefaultAppDatabase(*flagConfigPath)

	switch *flagType {
	case "archlinux":
		archlinux.Archlinux(*flagGenDot)
	case "debian":
		debian.Debian(*flagGenDot)
	case "deepin":
		deepin.Deepin(*flagGenDot)
	case "ubuntu":
		ubuntu.Ubuntu(*flagGenDot)
	case "nix":
		if *flagGenDot == "" {
			fmt.Errorf("Nix not support gendot")
		}
		nix.Nix(*workerCount, *batchSize)
	case "homebrew":
		homebrew.Homebrew(*flagGenDot)
	case "gentoo":
		gentoo.Gentoo(*flagGenDot)
	case "fedora":
		fedora.Fedora(*flagGenDot)
	case "centos":
		centos.Centos(*flagGenDot)
	case "alpine":
		alpine.Alpine(*flagGenDot)
	case "aur":
		aur.Aur(*flagGenDot)
	}
}
