#

GO_VER_MIN=1.19
SUB_GO_MOD_FILES=$(shell find . -name 'go.mod')
SUB_GO_MOD_DIRS=$(subst ./,, $(sort $(dir $(SUB_GO_MOD_FILES))))

.PHONY: nothing
nothing:
	@echo 欢迎下载 go.util 代码

# go mod tidy
_MOD_TGTS=$(addprefix _MOD, $(SUB_GO_MOD_DIRS))
.PHONY: mod
mod: $(_MOD_TGTS)
	go work sync

.PHONY: $(_MOD_TGTS)
$(_MOD_TGTS):
	@for dir in $(subst _MOD,, $@); do \
		cd $$dir; \
		echo ======== go mod tidy $$dir ========; \
		go mod tidy; \
	done

# go test
_TEST_TGTS=$(addprefix _TEST, $(SUB_GO_MOD_DIRS))
.PHONY: test
test: $(_TEST_TGTS)

.PHONY: $(_TEST_TGTS)
$(_TEST_TGTS):
	@for dir in $(subst _TEST,, $@); do \
		cd $$dir; echo ======== TEST $$dir ========; go test -v ./...; \
	done

# go get -u
_UP_TGTS=$(addprefix _UP, $(SUB_GO_MOD_DIRS))
.PHONY: up
up: $(_UP_TGTS)

.PHONY: $(_UP_TGTS)
$(_UP_TGTS):
	@for dir in $(subst _UP,, $@); do \
		cd $$dir; \
		rm -f go.mod go.sum; \
		echo "module github.com/Andrew-M-C/go.util/$$dir" > go.mod; \
		echo "" >> go.mod; \
		echo "go $(GO_VER_MIN)" >> go.mod; \
		echo ======== go mod tidy $$dir ========; \
		go mod tidy; \
	done

