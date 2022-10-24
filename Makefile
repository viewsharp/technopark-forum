CURDIR=$(shell pwd)
BINDIR=${CURDIR}/bin
GOVER=$(shell go version | perl -nle '/(go\d\S+)/; print $$1;')
EASYJSON=${BINDIR}/easyjson_${GOVER}

install-easyjson:
	test -f ${EASYJSON} || \
		(GOBIN=${BINDIR} go install github.com/mailru/easyjson/...@latest && \
		mv ${BINDIR}/easyjson ${EASYJSON})


generate: install-easyjson
	${EASYJSON} -all <file>.go
