/**
 *  MindLab
 *
 *  Create by songli on 2020/10/23
 *  Copyright © 2021 imind.tech All rights reserved.
 */

package template

import (
	"os"
	"text/template"
)

// 生成docker
func CreateBuild(data *Data) error {
	// 生成Makefile
	var tpl = `GOPATH := $(shell go env GOPATH)
VERSION := 0.0.1.1

gengo:
	@protoc -I. --proto_path ../../application/proto \
  --go_out ../../application/proto --go_opt paths=source_relative \
  --micro_out ../../application/proto --micro_opt paths=source_relative {{.Service}}/{{.Service}}.proto
	@protoc-go-inject-tag -input=../../application/proto/{{.Service}}/{{.Service}}.pb.go

depend:
	go get ../../...

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o {{.Service}} ../../main.go ../../plugins.go

test:
	go test -v ../... -cover

docker: build health
	docker build -f ./Dockerfile -t 348681422678.dkr.ecr.ap-southeast-1.amazonaws.com/{{.Project}}/{{.Service}}:$(VERSION) ../
	#docker push 348681422678.dkr.ecr.ap-southeast-1.amazonaws.com/{{.Project}}/{{.Service}}:$(VERSION)
	rm -rf {{.Service}} grpc-health-probe

health:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o grpc-health-probe ../pkg/grpc-health-probe/main.go

deploy: docker
	kubectl apply -f ns-rbac.yaml
	kubectl apply -f {{.Service}}.yaml

helm:
	helm install {{.Service}} ./helm --set image.tag=$(VERSION)

clean:
	docker rmi 348681422678.dkr.ecr.ap-southeast-1.amazonaws.com/{{.Project}}/{{.Service}}:$(VERSION)

k8s: docker
	kubectl set image deployment/{{.Service}} {{.Service}}=348681422678.dkr.ecr.ap-southeast-1.amazonaws.com/{{.Project}}/{{.Service}}:$(VERSION)

.PHONY: gengo depend build test docker health deploy helm clean k8s
`

	t, err := template.New("makefile").Parse(tpl)
	if err != nil {
		return err
	}

	dir := "./" + data.Domain + "/" + data.Project + "/" + data.Service + "/infrastructure/build/"

	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	fileName := dir + "Makefile"

	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	err = t.Execute(f, data)
	if err != nil {
		return err
	}
	f.Close()

	// 生成Dockerfile
	tpl = `FROM golang:alpine as builder
RUN apk --no-cache add git
WORKDIR /go/src/{{.Domain}}/{{.Project}}/{{.Service}}/
COPY . .
ENV GOPROXY=https://goproxy.cn,direct
RUN go get ./...
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o {{.Service}} main.go plugins.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o grpc-health-probe infrastructure/pkg/grpc-health-probe/main.go

FROM alpine:latest
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories
RUN apk add --no-cache tzdata \
    && ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && rm -rf /var/cache/apk/* /tmp/* /var/tmp/* $HOME/.cache

WORKDIR .
ADD conf /infrastructure/conf
COPY --from=builder /go/src/{{.Domain}}/{{.Project}}/{{.Service}}/{{.Service}} /go/src/{{.Domain}}/{{.Project}}/{{.Service}}/grpc-health-probe /bin/
ENTRYPOINT [ "/bin/{{.Service}}" ]
`

	t, err = template.New("dockerfile").Parse(tpl)
	if err != nil {
		return err
	}

	fileName = dir + "Dockerfile"

	f, err = os.Create(fileName)
	if err != nil {
		return err
	}
	err = t.Execute(f, data)
	if err != nil {
		return err
	}
	f.Close()

	return nil
}
