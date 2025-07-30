FROM golang:1.23.9-alpine

# 필요한 패키지 설치
RUN apk add --no-cache ca-certificates tzdata git

# 작업 디렉토리 설정
WORKDIR /app

# Go 모듈 파일 복사
COPY go.mod go.sum ./
RUN go mod download

# 소스 코드 복사
COPY . .

# 펌웨어 디렉토리 생성
RUN mkdir -p firmware && \
    echo "dummy firmware content" > firmware/latest.bin

# 애플리케이션 빌드
RUN go build -o main . && \
    ls -l main && \
    chmod +x main && \
    ls -l main

# 포트 노출
EXPOSE 8080

# 애플리케이션 실행
CMD ["./main"]