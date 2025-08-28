# 前端构建阶段
FROM node:18-alpine AS builder

WORKDIR /build

# 设置 npm 和 yarn 的镜像源
RUN npm config set registry https://registry.npmmirror.com/ && \
    yarn config set registry https://registry.npmmirror.com/

# 设置 Node.js 内存限制
ENV NODE_OPTIONS="--max-old-space-size=4096"

# 只复制 package.json
COPY web/package.json ./

# 安装依赖
RUN yarn cache clean && \
    yarn install --network-timeout 1000000 && \
    yarn add antd --network-timeout 1000000

# 安装较旧版本的 vite (4.x 版本更稳定)
RUN yarn add vite@4.5.2 --dev --network-timeout 1000000

# 复制其他源文件
COPY ./web .
COPY ./VERSION .

# 构建
RUN NODE_ENV=production \
    DISABLE_ESLINT_PLUGIN='true' \
    VITE_REACT_APP_VERSION=$(cat VERSION) \
    yarn build

# Go 构建阶段
FROM golang:alpine AS builder2

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux

WORKDIR /build

ADD go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=builder /build/dist ./web/dist
RUN go build -ldflags "-s -w -X 'one-api/common.Version=$(cat VERSION)'" -o one-api

# 最终阶段
FROM alpine
WORKDIR /data
RUN apk update \
    && apk upgrade --no-cache \
    && apk add --no-cache ca-certificates tzdata ffmpeg \
    && update-ca-certificates \
    && mkdir -p /usr/local/share/one-api/lang

COPY --from=builder2 /build/lang/*.json /usr/local/share/one-api/lang/
COPY --from=builder2 /build/one-api /
COPY --from=builder /build/dist /data/public/webHtml
EXPOSE 3000

ENTRYPOINT ["/one-api"]