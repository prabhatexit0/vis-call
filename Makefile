.PHONY: all build run clean 

# Name of the executable
BINARY_NAME=viscall

# 'Don't tell and Do Nothing' Rule to do nothing
# for the command params (considered rules) passed
# to the run rule defined below.
%:
	@:

build:
	go build -o bin/viscall .

run: npm i
	go run . $(filter-out $@,$(MAKECMDGOALS))

