FROM golang
RUN mkdir -p /go/src/Pipeline
WORKDIR /go/src/Pipeline
ADD /cmd/main.go .
ADD go.mod .

FROM alpine:latest
WORKDIR /root/
COPY --from=0 /go/bin/Pipeline .
ENTRYPOINT ./Pipeline
EXPOSE 8080