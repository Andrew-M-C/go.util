#

GO_VER_MIN=1.19
SUB_GO_MOD_FILES=$(shell find . -name 'go.mod')
SUB_GO_MOD_DIRS=$(subst ./,, $(sort $(dir $(SUB_GO_MOD_FILES))))

.PHONY: nothing
nothing:
	@echo 欢迎下载 go.util 代码

_MOD_DIRS=$(addprefix _MOD, $(SUB_GO_MOD_DIRS))
.PHONY: mod
mod: $(_MOD_DIRS)
	go work sync

.PHONY: $(_MOD_DIRS)
$(_MOD_DIRS):
	@for dir in $(subst _MOD,, $@); do \
		cd $$dir; echo go mod tidy $$dir; go mod tidy; \
	done

_TEST_DIRS=$(addprefix _TEST, $(SUB_GO_MOD_DIRS))
.PHONY: test
test: $(_TEST_DIRS)

.PHONY: $(_TEST_DIRS)
$(_TEST_DIRS):
	@for dir in $(subst _TEST,, $@); do \
		cd $$dir; echo ======== TEST $$dir ========; go test -v ./...; \
	done
