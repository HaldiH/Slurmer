GO=go
MIN_UID=1000
MAX_UID=65535
SLURMER_UID=1000
PREFIX=/usr/local
LDFLAGS = -s -w # Remove debug symbols and info

all: slurmer executor

slurmer: cmd/slurmer/main.go
	${GO} build -o "$@" -ldflags="${LDFLAGS}" $^

executor: cmd/cliexecutor/main.go
	${GO} build -o "$@" -ldflags="${LDFLAGS} \
		-X 'main.slurmerUid=${SLURMER_UID}' \
		-X 'main.minUid=${MIN_UID}' \
		-X 'main.maxUid=${MAX_UID}'" $^

install: executor slurmer
	mkdir -p ${PREFIX}/bin
	cp slurmer executor ${PREFIX}/bin/
	chown root:root ${PREFIX}/bin/executor
	chmod u+s ${PREFIX}/bin/executor

clean:
	rm -rf slurmer executor
