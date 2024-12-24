TEST_OPTIONS := -v -json -failfast -race -cover

.PHONY: all
all: help

# this is godly
# https://news.ycombinator.com/item?id=11939200
.PHONY: help
help:	### this screen. Keep it first target to be default
ifeq ($(UNAME), Linux)
	@grep -P '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
else
	@# this is not tested, but prepared in advance for you, Mac drivers
	@awk -F ':.*###' '$$0 ~ FS {printf "%15s%s\n", $$1 ":", $$2}' \
		$(MAKEFILE_LIST) | grep -v '@awk' | sort
endif


.PHONY: fmt
fmt: ### Executes the formatting pipeline on the project
	$(info: Make: Format)
	go fmt ./...
	goimports -w .
	golines --max-len=120 --reformat-tags -w .



.PHONY: test
test: ### Runs the test suite
	go test ${TEST_OPTIONS} ./... | tparse -all -progress
