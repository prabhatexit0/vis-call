.PHONY: all build run clean

# Name of the executable
BINARY_NAME=viscall

# 'Don't tell and Do Nothing' Rule to do nothing
# for the command params (considered rules) passed
# to the run rule defined below.
%:
	@:

install: packages.json
	npm install

build:
	go build -o bin/$(BINARY_NAME) .

run: build
	go run . $(filter-out $@,$(MAKECMDGOALS))

dbuild:
	docker build . -t $(BINARY_NAME)

drun:
	docker run -it -v "$(pwd)/:/app/" --name $(BINARY_NAME) $(filter-out $@,$(MAKECMDGOALS))

dclean:
	docker rm $(BINARY_NAME)
	docker rmi $(BINARY_NAME)