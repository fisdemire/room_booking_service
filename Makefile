.PHONY: up down mock help load-test

BASE_URL ?= http://localhost:8080
ROOM_ID ?=
TEST_DATE ?=

up:
	docker compose up -d --build

down:
	docker compose down

mock:
	mockery

load-test:
	k6 run \
		--env BASE_URL=$(BASE_URL) \
		--env ROOM_ID=$(ROOM_ID) \
		--env TEST_DATE=$(TEST_DATE) \
		load_test.js

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-12s %s\n", $$1, $$2}'