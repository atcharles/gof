## 参数定义
APP_ROOT=./apps
APP_NAME=clt
## 编译类型: 0.Mac编译 1.Linux编译 2.Linux +race
Type=0
Upx =0
ARG =
BLDDIR=$(APP_ROOT)/$(APP_NAME)/build
## 编译信息
VERSION=1.0.1
ifeq ($(Type), 2)
VERSION:=$(VERSION) -race
endif
LDFlags=-s -w -X 'main.version=$(VERSION)'
LDFlags+= -X 'main.buildTime=`date '+%Y-%m-%d_%H:%M:%S %Z/%p'`'
LDFlags+= -X 'main.goVersion=`go version`'
LDFlags+= -X 'main.gitBranch=`git rev-parse --abbrev-ref HEAD`'
LDFlags+= -X 'main.gitHash=`git rev-parse HEAD`'

## 编译
.PHONY:build clean
build:clean
ifeq ($(Type), 0)
	@go build -ldflags="$(LDFlags)" $(ARG) -o $(BLDDIR)/$(APP_NAME).out $(APP_ROOT)/$(APP_NAME)/*.go
else ifeq ($(Type), 1)
	@CGO_ENABLED=0 GOOS="linux" GOARCH="amd64" \
	go build -ldflags="$(LDFlags)" $(ARG) -o $(BLDDIR)/$(APP_NAME).out $(APP_ROOT)/$(APP_NAME)/*.go
else ifeq ($(Type), 2)
	@CGO_ENABLED=1 GOOS="linux" GOARCH="amd64" \
	go build -ldflags="$(LDFlags) -extldflags -static" -race $(ARG) -o $(BLDDIR)/$(APP_NAME).out $(APP_ROOT)/$(APP_NAME)/*.go
endif
## 压缩
ifneq ($(Upx), 0)
	@upx -9 $(BLDDIR)/$(APP_NAME).out
endif

$(BLDDIR)/%:
	mkdir -p $@

## 清理
clean:
	@rm -rf $(BLDDIR)/*logs
	@rm -rf $(BLDDIR)/*.out
	@rm -rf $(BLDDIR)/*.log
