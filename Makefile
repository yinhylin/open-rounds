.PHONY: compile

all:
	clang-format -style=google -i proto/*.proto
	protoc --go_out=. proto/*.proto
