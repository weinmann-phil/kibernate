test:
	./scripts/run-all-tests.sh

docker-build:
	./scripts/docker-build.sh linux/amd64,linux/arm64

docker-push-ghcr:
ifndef IMAGE_TAGS
	$(error IMAGE_TAGS is undefined)
endif
	./scripts/docker-push-ghcr.sh $(IMAGE_TAGS)

helm-package:
	./scripts/helm-package.sh

helm-push-ghcr:
	./scripts/helm-push-ghcr.sh