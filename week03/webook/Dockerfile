FROM golang:1.21.0 as builder
ENV GOPROXY https://goproxy.cn,direct
WORKDIR /go/src/project
COPY . .
RUN CGO_ENABLED=0 OOS=linux GOARCH=amd64 go build -tags=k8s -o /opt/webook

FROM scratch
COPY --from=builder /opt/webook /app/webook
WORKDIR /app
ENTRYPOINT ["/app/webook"]