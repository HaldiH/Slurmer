GO=go
execMinUid=1000
execMaxUid=65535
execSlurmerUid=1000

all: slurmer executor

slurmer: cmd/slurmer/main.go
	${GO} build -o "$@" $^

executor: cmd/cliexecutor/main.go
	${GO} build -o "$@" -ldflags=" \
		-X 'main.slurmerUid=${execSlurmerUid}' \
		-X 'main.minUid=${execMinUid}' \
		-X 'main.maxUid=${execMaxUid}'" $^

install: executor slurmer
	chown root:root executor
	chmod u+s executor
	chmod g+s executor

clean:
	rm -rf slurmer executor