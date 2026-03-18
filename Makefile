.PHONY: setup web api dev stop test build lint clean prepare-rootfs

setup:
	cd apps/web && npm install

api:
	cd apps/api && go run ./cmd/server

web:
	cd apps/web && npm run dev

dev:
	./scripts/dev.sh

stop:
	./scripts/stop.sh

test:
	cd apps/api && go test ./...
	cd apps/web && npm run test

build:
	cd apps/api && go test ./...
	cd apps/web && npm run build

lint:
	cd apps/web && npm run lint

clean:
	cd apps/web && rm -rf dist

prepare-rootfs:
	./scripts/prepare-rootfs.sh
