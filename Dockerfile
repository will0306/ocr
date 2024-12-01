FROM golang:1.23.3  as builder

WORKDIR /workspace
COPY . /workspace
RUN echo "init build workspace" \
&& mkdir /tmp/workspace && mkdir /tmp/workspace/config && go mod tidy && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath --ldflags "-s -w -extldflags '-static -L/usr/local/lib -ltdjson_static -ltdjson_private -ltdclient -ltdcore -ltdactor -ltddb -ltdsqlite -ltdnet -ltdutils -ldl -lm -lssl -lcrypto -lstdc++ -lz'" -o /tmp/workspace/app main.go && cp config/config.yaml /tmp/workspace/config/ && ls -al /tmp/workspace

FROM alpine:latest
WORKDIR /workspace
COPY --from=builder /tmp/workspace/ /workspace
RUN apk update && apk add tzdata
CMD nohup ./app -f /workspace/config/config.yaml
