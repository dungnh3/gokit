hello:
	echo hello

update:
	GOSUMDB=off go mod tidy

.PHONY: update