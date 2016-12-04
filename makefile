.PHONY: all fmt tags doc js

all:
	go install -v ./...
	gofmt -s -w -l .
	go install -v ./...
	e8chk -path="shanhu.io/third"
	golint ./...
	gotags -R . > tags

rall:
	touch `find . -name "*.go"`
	go install -v ./...

fmt:
	gofmt -s -w -l .

tags:
	gotags -R . > tags

test:
	gofmt -s -w -l .
	go test ./...

testv:
	go test -v ./...

testc:
	go test -cover ./...

lc:
	wc -l `find . -name "*.go"`

doc:
	godoc -http=localhost:8000

lint:
	golint ./...

fmtchk:
	gofmt -d -l .
