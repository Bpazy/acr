BINDIR=bin
FULLNAME=acr-$@
LDFLAGS="-s -w -X github.com/Bpazy/acr.buildVer=${VERSION}"
GOBUILD=go build -ldflags=${LDFLAGS}
CMDPATH=.

all: linux-amd64 darwin-amd64 darwin-arm64 windows-amd64 # Most used

darwin-amd64:
	GOARCH=amd64 GOOS=darwin $(GOBUILD) -o $(BINDIR)/$(FULLNAME) $(CMDPATH)

darwin-arm64:
	GOARCH=amd64 GOOS=darwin $(GOBUILD) -o $(BINDIR)/$(FULLNAME) $(CMDPATH)

linux-amd64:
	GOARCH=amd64 GOOS=linux $(GOBUILD) -o $(BINDIR)/$(FULLNAME) $(CMDPATH)

windows-amd64:
	GOARCH=amd64 GOOS=windows $(GOBUILD) -o $(BINDIR)/$(FULLNAME).exe $(CMDPATH)

install:
	go install -ldflags=${LDFLAGS} $(CMDPATH)

clean:
	rm $(BINDIR)/*
