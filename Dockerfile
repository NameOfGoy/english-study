FROM registry.cn-shenzhen.aliyuncs.com/qzes/builder:td3330 AS builder
WORKDIR /app

# 将所有源代码添加到构建上下文
ADD . /app

# 设置环境变量
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct
ENV GOOS=linux
ENV GOCACHE=/opt/go/.cache/go-build
ENV CGO_ENABLED=1

# 运行构建命令
RUN --mount=type=cache,target=/opt/go/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go mod tidy && \
    go build -gcflags '-N -l' -ldflags '-s -w -X main.Version=1112' -o ./bin/englishstudy /app

FROM ubuntu:22.04

# 替换为国内源并安装依赖
RUN sed -i 's|http://archive.ubuntu.com|http://mirrors.aliyun.com|g' /etc/apt/sources.list && \
    sed -i 's|http://security.ubuntu.com|http://mirrors.aliyun.com|g' /etc/apt/sources.list && \
    apt-get update && \
    apt-get install -y ca-certificates vim && \
    update-ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/bin/englishstudy /app/bin/englishstudy

CMD ["/app/bin/englishstudy", "-f", "/app/config/config.yaml"]
