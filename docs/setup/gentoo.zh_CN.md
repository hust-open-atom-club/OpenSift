# 如何设置 Gentoo prefix 以建立系统

## 为 Gentoo prefix 准备一个目录

```bash
mkdir -p <你想要放置 Gentoo prefix 的位置>
```

## 准备一个 Docker 容器

```bash
docker run --name build-gentoo -v <你想要放置 Gentoo prefix 的位置>:/root -it ubuntu:latest
```

## 安装必要的软件包

```bash
apt update
apt install -y curl build-essential git
export GENTOO_MIRRORS=http://mirrors.hust.edu.cn/gentoo
curl -L https://gitweb.gentoo.org/repo/proj/prefix.git/plain/scripts/bootstrap-prefix.sh | sed '2690,+10d' > build-gentoo.sh
./build-gentoo.sh
```

运行 build-gentoo.sh 时，脚本会询问一些问题，你可以直接按回车键使用默认值。

脚本将花费数小时才能完成，所以你可以留在那里做其他事情。

### 构建完成后

退出容器， Gentoo prefix 目录是 `你想要放置 Gentoo prefix 的位置/gentoo`。
你可以设置环境变量 `$GENTOO_PREFIX_DIR` 并运行设置脚本来建立系统。