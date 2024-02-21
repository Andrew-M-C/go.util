#

SUB_GO_MOD_FILES=$(shell find . -name 'go.mod')
SUB_GO_MOD_DIRS=$(sort $(dir $(SUB_GO_MOD_FILES)))

.PHONY: nothing
nothing:
	@echo 欢迎下载 go.util 代码

.PHONY: mod
mod: $(SUB_GO_MOD_DIRS)
	go work sync

.PHONY: $(SUB_GO_MOD_DIRS)
$(SUB_GO_MOD_DIRS):
	@cd $@ && echo go mod tidy $@ && go mod tidy
