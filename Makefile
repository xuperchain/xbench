ifeq ($(OS),Windows_NT)
	PLATFORM="Windows"
else
	ifeq ($(shell uname),Darwin)
		PLATFORM="MacOS"
	else
	    PLATFORM="Linux"
	endif
endif

all: build

build:
	sh script/build.sh

.PHONY: test 
