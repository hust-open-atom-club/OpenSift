# How to setup Gentoo prefix for seting up the system


## Prepare a directory for the prefix

```bash
mkdir -p WHERE_YOU_WANT_TO_PUT_THE_PREFIX
```

## Prepare a docker container

```bash
docker run --name build-gentoo -v WHERE_YOU_WANT_TO_PUT_THE_PREFIX:/root -it ubuntu:latest
```

## Install the necessary packages

```bash
apt update
apt install -y curl build-essential git
export GENTOO_MIRRORS=http://mirrors.hust.edu.cn/gentoo
curl -L https://gitweb.gentoo.org/repo/proj/prefix.git/plain/scripts/bootstrap-prefix.sh | sed '2690,+10d' > build-gentoo.sh
./build-gentoo.sh
```

When running `build-gentoo.sh`, the script will ask some questions, you can just press enter to use the default value. 

The script will take hours to finish, so you can just leave it there and do something else.

## After building 

Quit the container and the prefix dir is `WHERE_YOU_WANT_TO_PUT_THE_PREFIX/gentoo`.
And you can set the environment `$GENTOO_PREFIX_DIR` and run setup script to setup the system.