.PHONY: up down mock help

up:
	docker compose up -d --build

down:
	docker compose down

mock:
	mockery

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-12s %s\n", $$1, $$2}'