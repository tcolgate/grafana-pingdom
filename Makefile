


all: deploy-production

PHONY: grafana-pingdom docker

grafana-pingdom: main.go go.sum go.mod
	GOOS=linux GOARCH=amd64 GO111MODULE=on CGO_ENABLED=0 go build .

deploy-production: grafana-pingdom
	baton -e production -d

