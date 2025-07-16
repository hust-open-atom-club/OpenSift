#!/bin/bash
# The script helps to set up the app
set -e
set -o pipefail

COMP_DIR="${BASH_SOURCE%/*}"
cd "$COMP_DIR"

if [ -n "$(git status --porcelain)" ]; then
    GIT_STATUS="(dirty)"
fi

about() {
    echo "Criticality Score Setup tool"
    echo "============================"
    echo "Homepage: https://github.com/HUSTSecLab/OpenSift"
    echo "Version :" "$(git rev-parse HEAD)" "$GIT_STATUS"
    echo
}

help() {
    echo "Usage: $0 [options...]"
    echo "Options:"
    echo " -h              Show this help message and exit"
    echo " -a <api_port>   The port for the api server, default is 8081"
    echo " -s <storage_dir> The directory to store the git repositories, default is ./data/git"
    echo " -d <data_dir>   The directory to store the data, default is ./data"
    echo " -p <db_passwd>  The password for the database, default is randomly generated"
    echo " -w <web_port>   The port for the web server, default is 8080"
    echo " -b <db_port>    The port for the database, default is 5432"
}

echo_red() {
    echo -e "\033[31m$*\033[0m"
}

########## Init ##########

about

DB_HOST_PORT="5432"
WEB_HOST_PORT="8080"
APISERVER_HOST_PORT="8081"
STORAGE_DIR="./data/git"

while getopts "s:a:d:p:w:b:h" opt; do
    case $opt in
    a)
        APISERVER_HOST_PORT="$OPTARG"
        ;;
    d)
        DATA_DIR="$OPTARG"
        ;;
    p)
        DB_PASSWD="$OPTARG"
        ;;
    w)
        WEB_HOST_PORT="$OPTARG"
        ;;
    b)
        DB_HOST_PORT="$OPTARG"
        ;;
    s)
        STORAGE_DIR="$OPTARG"
        ;;
    h)
        help
        exit 0
        ;;
    \?)
        echo "Invalid option: -$OPTARG" >&2
        help
        exit 1
        ;;
    esac
done

shift $((OPTIND - 1))

if [ -z "$DATA_DIR" ]; then
    DATA_DIR="./data"
fi

if [ -z "$DB_PASSWD" ]; then
    DB_PASSWD=$(openssl rand -base64 12)
fi

if [ -f "$DATA_DIR"/DB_PASSWD ]; then
    echo_red "Password file already exists, -p will be ignored"
    DB_PASSWD=$(cat "$DATA_DIR"/DB_PASSWD)
else
    mkdir -p "$DATA_DIR"
    echo "$DB_PASSWD" >"$DATA_DIR"/DB_PASSWD
fi

########## Process ##########

if [ -f ".env" ]; then
    echo_red "It seems that the app is already set up."
    echo_red "If you want to upgrade, please run "
    echo_red "    docker compose build & docker compose up -d"
    echo
    echo -n "Do you want to continue setup again? [y/N] "

    read -r answer
    if [ "$answer" != "y" ] && [ "$answer" != "Y" ]; then
        exit 0
    fi
fi

if [ -z "$GITHUB_TOKEN" ]; then
    echo_red "It seems that you haven't set the GITHUB_TOKEN."
    echo_red "enumerate_github will not work without it."
    echo
    echo -n "Do you want to ignore it and continue setup? [y/N] "

    read -r answer
    if [ "$answer" != "y" ] && [ "$answer" != "Y" ]; then
        exit 0
    fi
fi

if [ -z "$GENTOO_PREFIX_DIR" ]; then
    echo_red "It seems that you haven't set the GENTOO_PREFIX_DIR."
    echo_red "enumerate_gentoo will not work without it."
    echo
    echo 'About how to set GENTOO_PREFIX_DIR, please refer to the `docs/setup/gentoo.md`.'
    echo
    echo -n "Do you want to ignore it and continue setup? [y/N] "

    read -r answer
    if [ "$answer" != "y" ] && [ "$answer" != "Y" ]; then
        exit 0
    fi
    GENTOO_PREFIX_DIR="$DATA_DIR/gentoo"
fi

# 1. Create dirs and files

echo "Setting up files..."

mkdir -p "$DATA_DIR/db" "$DATA_DIR/rec" "$DATA_DIR/config" "$DATA_DIR/git" "$DATA_DIR/log" "$DATA_DIR/gentoo"

cat <<EOF >"$DATA_DIR/config/config.json"
{
    "database": "criticality_score",
    "host": "db",
    "user": "postgres",
    "password": "$DB_PASSWD",
    "port": "5432",
    "GitHubToken": "$GITHUB_TOKEN"
}
EOF

cat <<EOF >".env"
DATA_DIR=$DATA_DIR
DB_HOST_PORT=$DB_HOST_PORT
DB_PASSWD=$DB_PASSWD
WEB_HOST_PORT=$WEB_HOST_PORT
APISERVER_HOST_PORT=$APISERVER_HOST_PORT
STORAGE_DIR=$STORAGE_DIR
GITHUB_TOKEN=$GITHUB_TOKEN
GENTOO_PREFIX_DIR=$GENTOO_PREFIX_DIR
EOF

# 2. Start docker compose

echo "Setting up app..."

docker compose build
docker compose up -d

# 3. Create database and tables

echo "Waiting for database to start..."
sleep 5

docker compose cp ./schema.sql db:/tmp/schema.sql
docker compose exec db psql -h localhost -U postgres -f /tmp/schema.sql
docker compose exec db rm /tmp/schema.sql

# 3. Run first time collector
echo "Running workflow for the first time..."
docker compose exec app /workflow/update.sh -C /data/rec package

echo_red "========== NOTICE =========="
echo_red "git link could only be updated manually."
echo_red "Try following steps to update git link:"
echo_red "    1. use home2git tool to find the git link"
echo_red "    2. update the git link in database, database password is $DB_PASSWD"
echo_red "    3. run 'docker compose exec app /workflow/update.sh gitlink'"

echo
echo "Done!"
