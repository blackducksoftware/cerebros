COMPONENTS = blackduck-cli blackduck-cli-single-scan get-docker-image git-repo-fetcher idir-gen idir-scanner kube-metrics polaris-api-load-gen polaris-api-token polaris-cli polaris-download-and-capture queue scan-queue scan-queue-client-example service-load-gen

IMAGE_BASE=gcr.io/eng-dev/blackducksoftware/cerebros/
IMAGE_TAG=master

CURRENT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
OUTDIR=_output

.PHONY: test ${OUTDIR} ${COMPONENTS}

all: compile

compile: ${OUTDIR} ${COMPONENTS}

${COMPONENTS}:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./cmd/$@/$@ ./cmd/$@
	docker build -t $(IMAGE_BASE)$@:$(IMAGE_TAG) ./cmd/$@
	mv cmd/$@/$@ $(OUTDIR)
	## gcloud docker -- push $(IMAGE_BASE)$@:$(IMAGE_TAG)

docker-image: $(COMPONENTS)
	$(foreach p,${COMPONENTS},cd ${CURRENT_DIR}/cmd/$p; docker build -t $(IMAGE_BASE)${p}:$(IMAGE_TAG) .;)

gcloud-docker-push: docker-image
	$(foreach p,${COMPONENTS},gcloud docker -- push $(IMAGE_BASE)${p}:$(IMAGE_TAG);)

test:
	go test ./pkg/...

clean:
	rm -rf ${OUTDIR}
	$(foreach p,${COMPONENTS},rm -f cmd/$p/$p;)

${OUTDIR}:
	mkdir -p ${OUTDIR}

install:
	go get -u golang.org/x/lint/golint

lint:
	go fmt ./pkg/... ./cmd/...
	#./hack/verify-gofmt.sh
	# ./hack/verify-golint.sh
	./hack/verify-govet.sh

fmt:
	go fmt ./cmd/... ./pkg/...

vet:
	go vet ./cmd/... ./pkg/...
