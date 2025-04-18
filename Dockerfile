FROM oven/bun:latest AS builder

WORKDIR /build
COPY web/package.json .
RUN bun install
COPY ./web .
COPY ./VERSION .
RUN DISABLE_ESLINT_PLUGIN='true' VITE_REACT_APP_VERSION=$(cat VERSION) bun run build

FROM golang:alpine AS builder2

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux

WORKDIR /build

ADD go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=builder /build/dist ./web/dist
COPY --from=builder /public/webHtml ./public/webHtml
RUN go build -ldflags "-s -w -X 'one-api/common.Version=$(cat VERSION)'" -o one-api

FROM alpine

RUN apk update \
    && apk upgrade \
    && apk add --no-cache ca-certificates tzdata ffmpeg \
    && update-ca-certificates \
    && mkdir -p /usr/local/share/one-api/lang    # 使用标准的应用数据目录

# 从 builder2 阶段复制文件
COPY --from=builder2 /build/lang/*.json /usr/local/share/one-api/lang/
COPY --from=builder2 /build/one-api /
COPY --from=builder2 /public /public
EXPOSE 3000
WORKDIR /data
ENTRYPOINT ["/one-api"]
