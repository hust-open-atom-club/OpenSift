FROM golang:1.23 AS builder

ENV GO111MODULE=on

# 设置 Go 模块代理为 goproxy.cn（中国大陆加速），加快依赖下载速度
ENV GOPROXY=https://goproxy.cn,direct

# predownload go modules to speed up incremental builds
RUN mkdir -p /build
WORKDIR /build
COPY go.mod go.sum /build/
RUN go mod download && rm -rf /build

# build the app
COPY ./cmd /build/cmd
COPY ./migrations /build/migrations
COPY ./pkg /build/pkg
COPY ./Makefile ./go.mod ./go.sum /build/

RUN make && \
    cp -r bin /app

# final stage
FROM ubuntu:24.04

RUN apt-get update && apt-get install -y ca-certificates && \
    curl -sSfL http://mirrors.hust.edu.cn/get | sh -s -- deploy ubuntu && \
    apt-get update && apt-get install -y git cron make curl xz-utils build-essential

ENV TZ=Asia/Shanghai
ENV PATH=/app:$PATH

ENV CFG_FILE=/config/config.json
ENV APP_BIN=/app
ENV STORAGE_DIR=/storage

# # install nix package manager
# RUN bash -c 'sh <(curl -L https://nixos.org/nix/install) --daemon'

# # add nix to the path
# ENV NIX_PROFILES="/nix/var/nix/profiles/default /root/.nix-profile"
# ENV PATH=/root/.nix-profile/bin:$PATH

COPY --from=builder /app /app

CMD ["echo", "hello, please run executables under /app to continue"]


