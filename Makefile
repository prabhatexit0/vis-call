.PHONY: all build run clean

# Silents all the targets
.SILENT:

# Name of the executable
BINARY_NAME=viscall

# 'Don't tell and Do Nothing' Rule to do nothing
# for the command params (considered rules) passed
# to the run rule defined below.
%:
	@:

install:
	npm install

build:
	go build -o bin/$(BINARY_NAME) .

run: build
	go run . $(filter-out $@, $(MAKECMDGOALS))

clean:
	rm -rf node_modules bin

dbuild:
	docker build . -t $(BINARY_NAME)

drun:
	docker run -it -v "$(pwd)/:/app/" --name $(BINARY_NAME) $(filter-out $@, $(MAKECMDGOALS))

dclean:
	docker rm $(BINARY_NAME)
	docker rmi $(BINARY_NAME)