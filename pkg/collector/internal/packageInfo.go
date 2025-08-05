package collector

import (
	"log"

	"github.com/HUSTSecLab/OpenSift/pkg/storage"
	"github.com/HUSTSecLab/OpenSift/pkg/storage/repository"
)

type PackageInfoInterface interface {
	ParseDistPackage() *repository.DistPackage
	ParseDistLinkInfo() *repository.DistDependency
	GetGitlinkByPkg(ac storage.AppDatabaseContext)
	CalculateImpact(count int)
}

type PackageInfo struct {
	DirectDepends          []string `json:"Depends"`
	IndirectDepends        []string
	DependsCount           int
	Description            string
	Homepage               string `json:"URL"`
	Name                   string
	PageRank               float64
	Version                string
	Impact                 float64
	Gitlink                string
	Type                   repository.DistType
	DistPackageTablePrefix repository.DistPackageTablePrefix
}

type PackageURL []string

var (
	FedoraURL = PackageURL{
		"https://mirrors.aliyun.com/fedora/releases/41/Everything/source/tree/repodata/df7750a80c5a4e4ff04ff5a1a499d32b6379dd50680b29140638e6edb1d71d68-primary.xml.gz",
	}
	DebianURL = PackageURL{
		"https://mirrors.hust.edu.cn/debian/dists/stable/main/binary-amd64/Packages.gz",
	}
	CentOSURL = PackageURL{
		"https://mirrors.aliyun.com/centos/7/os/x86_64/repodata/2b479c0f3efa73f75b7fb76c82687744275fff78e4a138b5b3efba95f91e099e-primary.xml.gz",
	}
	GentooURL = PackageURL{
		"https://github.com/gentoo/gentoo.git",
	}
	HomebrewURL = PackageURL{
		"https://github.com/Homebrew/homebrew-core.git",
	}
	UbuntuURL = PackageURL{
		"https://mirrors.hust.edu.cn/ubuntu/dists/jammy/main/binary-amd64/Packages.gz",
		"https://mirrors.hust.edu.cn/ubuntu/dists/jammy/universe/binary-amd64/Packages.gz",
		"https://mirrors.hust.edu.cn/ubuntu/dists/jammy/multiverse/binary-amd64/Packages.gz",
		"https://mirrors.hust.edu.cn/ubuntu/dists/jammy/restricted/binary-amd64/Packages.gz",
	}
	AlpineURL = PackageURL{
		"https://mirrors.aliyun.com/alpine/v3.21/main/x86_64/APKINDEX.tar.gz",
	}
	ArchlinuxURL = PackageURL{
		"https://mirrors.hust.edu.cn/archlinux/community/os/x86_64/community.files.tar.gz",
		"https://mirrors.hust.edu.cn/archlinux/community-staging/os/x86_64/community-staging.files.tar.gz",
		"https://mirrors.hust.edu.cn/archlinux/community-testing/os/x86_64/community-testing.files.tar.gz",
		"https://mirrors.hust.edu.cn/archlinux/core/os/x86_64/core.files.tar.gz",
		"https://mirrors.hust.edu.cn/archlinux/core-staging/os/x86_64/core-staging.files.tar.gz",
		"https://mirrors.hust.edu.cn/archlinux/core-testing/os/x86_64/core-testing.files.tar.gz",
		"https://mirrors.hust.edu.cn/archlinux/extra/os/x86_64/extra.files.tar.gz",
		"https://mirrors.hust.edu.cn/archlinux/extra-staging/os/x86_64/extra-staging.files.tar.gz",
		"https://mirrors.hust.edu.cn/archlinux/extra-testing/os/x86_64/extra-testing.files.tar.gz",
		"https://mirrors.hust.edu.cn/archlinux/gnome-unstable/os/x86_64/gnome-unstable.files.tar.gz",
		"https://mirrors.hust.edu.cn/archlinux/kde-unstable/os/x86_64/kde-unstable.files.tar.gz",
		"https://mirrors.hust.edu.cn/archlinux/multilib/os/x86_64/multilib.files.tar.gz",
		"https://mirrors.hust.edu.cn/archlinux/multilib-staging/os/x86_64/multilib-staging.files.tar.gz",
		"https://mirrors.hust.edu.cn/archlinux/multilib-testing/os/x86_64/multilib-testing.files.tar.gz",
		"https://mirrors.hust.edu.cn/archlinux/staging/os/x86_64/staging.files.tar.gz",
		"https://mirrors.hust.edu.cn/archlinux/testing/os/x86_64/testing.files.tar.gz",
	}
	AurURL = PackageURL{
		"https://aur.archlinux.org/packages-meta-ext-v1.json.gz",
	}
	DeepinURL = PackageURL{
		"https://mirrors.hust.edu.cn/deepin/beige/dists/beige/main/binary-amd64/Packages.gz",
	}
	OpenEulerURL = PackageURL{
		"https://mirrors.hust.edu.cn/openeuler/openEuler-25.03/source/repodata/d2c8439b5d4ef77caf0aba57453b255c955625b95b5231266c0f668223524800-primary.xml.zst",
	}
	OpenKylinURL = PackageURL{
		"https://mirrors.hust.edu.cn/openkylin/dists/huanghe/main/binary-amd64/Packages.gz",
	}

	OpenAnolisURL = PackageURL{
		"https://mirrors.openanolis.cn/anolis/23/os/x86_64/os/repodata/b8deff7d46ad19c92ec958ea1113fccece011dcb718b29dbd238864e7e28760b-primary.xml.gz",
	}

	OpenHarmonyURL = PackageURL{
		"https://gitee.com/openharmony/manifest/raw/OpenHarmony-5.1.0-Release/ohos/ohos.xml",
	}

	OpenCloudOSURL = PackageURL{
		"https://mirrors.opencloudos.tech/opencloudos/8.6/BaseOS/x86_64/os/repodata/f2da23eddb3242a3f3e71dfba303726141e5c01662f6f86d0fedd6dffde830bd-primary.xml.gz",
		"https://mirrors.opencloudos.tech/opencloudos/9.4/BaseOS/x86_64/os/repodata/d7efd4aec2a6bf15c7e7240dc1e1972f5764cc8610613d8329fe7ce0319fffb0-primary.xml.gz",
	}
)

func NewPackageInfo() PackageInfoInterface {
	return &PackageInfo{}
}
func (pkg *PackageInfo) ParseDistPackage() *repository.DistPackage {
	return &repository.DistPackage{
		Package:      &pkg.Name,
		Description:  &pkg.Description,
		HomePage:     &pkg.Homepage,
		Version:      &pkg.Version,
		DependsCount: &pkg.DependsCount,
	}
}

func (pkg *PackageInfo) ParseDistLinkInfo() *repository.DistDependency {
	return &repository.DistDependency{
		GitLink:   &pkg.Gitlink,
		DepImpact: &pkg.Impact,
		Type:      &pkg.Type,
		DepCount:  &pkg.DependsCount,
		PageRank:  &pkg.PageRank,
	}
}

func (pkg *PackageInfo) GetGitlinkByPkg(ac storage.AppDatabaseContext) {
	repo := repository.NewDistPackageRepository(ac, pkg.DistPackageTablePrefix)
	pkgInfo, err := repo.GetByName(pkg.Name)
	if err != nil {
		log.Println("Error getting package info from database:", err)
	}
	if pkgInfo.GitLink != nil {
		pkg.Gitlink = *pkgInfo.GitLink
	} else {
		log.Println("Error getting package info from database:", err)
	}
}

func (pkg *PackageInfo) CalculateImpact(count int) {
	pkg.Impact = float64(pkg.DependsCount) / float64(count)
}
