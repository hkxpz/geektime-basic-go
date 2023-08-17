FROM golang:1.21.0 as builder
ENV GOPROXY https://goproxy.cn,direct
ENV GOOS=linux
ENV GOARCH=arm

WORKDIR /go/src/project
COPY . /go/src/project
RUN go build -o /opt/webook

FROM scratch
COPY --from=builder /opt/webook /app/webook
WORKDIR /app
ENTRYPOINT ["/app/webook"]
