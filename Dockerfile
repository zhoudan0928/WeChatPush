FROM golang:1.21-alpine

WORKDIR /app

# 设置Go环境变量
ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct \
    CGO_ENABLED=0 \
    GOOS=linux

# 安装基本工具
RUN apk add --no-cache git tzdata

# 复制go.mod和go.sum
COPY go.mod go.sum ./

# 下载依赖并整理
RUN go mod download && go mod tidy

# 复制源代码
COPY . .

# 确保删除.env文件
RUN rm -f .env

# 构建应用
RUN go build -o main .

# 暴露端口
EXPOSE 8080

# 设置时区
ENV TZ=Asia/Shanghai

CMD ["./main"]
