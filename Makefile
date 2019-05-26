
deps:
	GOPATH=${HOME}/go:${PWD} \
	go get go.mongodb.org/mongo-driver/mongo \
	go get -d -v ...

build:
	GOPATH=${HOME}/go:${PWD} \
	go build ...