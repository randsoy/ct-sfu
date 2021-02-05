GOPROXY=https://goproxy.cn,direct
GOCMD=GO111MODULE=on GOPROXY=$(GOPROXY) go
GOBUILD=$(GOCMD) build -mod=vendor
GOTEST=$(GOCMD) test

build: init
	rm -rf target
	mkdir target/
	cp cmd/meet/meet-example.toml target/meet.toml
	$(GOBUILD) -o target/meet cmd/meet/main.go

init:
	go mod vendor


run:
	nohup target/meet -conf=target/meet.toml 2>&1 > target/meet.log &

stop:
	pkill -f target/meet

clean:
	rm -rf target/