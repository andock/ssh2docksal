BINARIES ?=	ssh2docksal
GODIR ?=	github.com/andock/ssh2docksal
RUN_ARGS ?=	-V --local-user=local-user


PKG_BASE_DIR ?=	.
CONVEY_PORT ?=	9042
SOURCES :=	$(shell find . -type f -name "*.go")
COMMANDS :=	$(shell go list ./... | grep -v /vendor/ | grep /cmd/)
PACKAGES :=	$(shell go list ./... | grep -v /vendor/ | grep -v /cmd/)
REL_COMMANDS := $(subst $(GODIR),./,$(COMMANDS))
REL_PACKAGES := $(subst $(GODIR),./,$(PACKAGES))
GOENV ?=	GO15VENDOREXPERIMENT=1
GO ?=		$(GOENV) go
USER ?=		andockio


all:	build


.PHONY: run
run: build
	./$(BINARIES) $(RUN_ARGS)


.PHONY: build
build:	$(BINARIES)


$(BINARIES):	$(SOURCES)
	dep ensure
	golint
	$(GO) build -o $@ main/main.go


.PHONY: test
test:
	#$(GO) get -t ./...
	$(GO) test -i $(PACKAGES) $(COMMANDS)
	$(GO) test -v $(PACKAGES) $(COMMANDS)


.PHONY: install
install:
	$(GO) install $(COMMANDS)


.PHONY: clean
clean:
	rm -f $(BINARIES)


.PHONY: re
re:	clean all


.PHONY: convey
convey:
	$(GO) get github.com/smartystreets/goconvey
	goconvey -cover -port=$(CONVEY_PORT) -workDir="$(realpath $(PKG_BASE_DIR))" -depth=1


.PHONY:	cover
cover:	profile.out


profile.out:	$(SOURCES)
	rm -f $@
	find . -name profile.out -delete
	for package in $(REL_PACKAGES); do \
	  rm -f $$package/profile.out; \
	  $(GO) test -covermode=count -coverpkg=$(PKG_BASE_DIR) -coverprofile=$$package/profile.out $$package; \
	done
	echo "mode: count" > profile.out.tmp
	cat `find . -name profile.out` | grep -v mode: | sort -r | awk '{if($$1 != last) {print $$0;last=$$1}}' >> profile.out.tmp
	mv profile.out.tmp profile.out


docker:
	docker build -t andockio/ssh2docksal .
                      \


.PHONY: docker-ps
docker-ps:
	docker ps --filter=label=ssh2docksal -a
