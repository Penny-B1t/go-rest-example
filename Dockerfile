FROM golang:1.23.9-alpine AS builder

# 빌드 환경 설정: 타겟 OS 및 아키텍처 지정
ENV GOOS=linux
ENV GOARCH=amd64

# 필요한 패키지 설치
RUN apk add --no-cache ca-certificates tzdata git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/main .

FROM alpine:latest

# 런타임에 필요한 패키지만 설치
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# 빌드 스테이지에서 컴파일된 실행 파일만 복사
COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]