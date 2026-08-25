FROM docker.m.daocloud.io/library/golang:1.26.3-bookworm

ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.google.cn

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /app/task225-timberjoint ./cmd/task225-timberjoint

EXPOSE 8080
ENTRYPOINT ["/app/task225-timberjoint"]
CMD ["--smoke-test"]
