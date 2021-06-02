GOLANG_IMAGE ?= golang:1.15

.PHONY: fmt deps-up

fmt:
	@echo "Formatting files..."
	@docker run --rm \
		-v $(CURDIR):/workspace \
		--entrypoint gofmt \
		$(GOLANG_IMAGE) -w -l -s \
		.

deps-up:
	@echo "Updating all dependencies..."
	@docker run --rm \
		-v $(CURDIR):/workspace \
		--workdir /workspace \
		$(GOLANG_IMAGE) /bin/sh -c "go get -u all && go mod tidy"
