# ============ STAGE 1: Build ============
FROM golang:1.25-alpine AS builder
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.google.cn
RUN apk add --no-cache git ca-certificates
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o main .

# ============ STAGE 2: Final tiny image (Run Time)============
FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Africa/Cairo
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone
RUN adduser -D -g '' appuser
WORKDIR /app
COPY --from=builder /app/main .
RUN chown appuser:appuser main
USER appuser
EXPOSE 8080
CMD ["./main"]