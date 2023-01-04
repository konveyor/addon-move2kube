BUILD_DIR=bin
BIN_NAME=addon-move2kube

.PHONY: clean
clean:
	rm -rf ${BUILD_DIR}/${BIN_NAME}

.PHONY: build
build: ${BUILD_DIR}/${BIN_NAME}

${BUILD_DIR}/${BIN_NAME}:
	go build -o ${BUILD_DIR}/${BIN_NAME} github.com/konveyor/addon-move2kube/cmd

.PHONY: cbuild
cbuild:
	docker build -t quay.io/konveyor/addon-move2kube:latest -f Dockerfile .

.PHONY: cpush
cpush:
	docker push quay.io/konveyor/addon-move2kube:latest
