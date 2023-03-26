test:
	./scripts/run-all-tests.sh

docker-build:
	./scripts/docker-build.sh

docker-buildx-push-ghcr:
ifndef IMAGE_TAGS
	$(error IMAGE_TAGS is undefined)
endif
	./scripts/docker-buildx-push-ghcr.sh $(IMAGE_TAGS)

helm-package:
	./scripts/helm-package.sh
