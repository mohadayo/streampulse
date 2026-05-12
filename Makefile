.PHONY: up down build test test-python test-go test-ts lint clean

up:
	docker compose up --build -d

down:
	docker compose down

build:
	docker compose build

test: test-python test-go test-ts

test-python:
	cd event-collector && pip install -r requirements.txt -q && pytest -v

test-go:
	cd event-processor && go test -v ./...

test-ts:
	cd api-gateway && npm install --silent && npm test

lint: lint-python lint-go lint-ts

lint-python:
	cd event-collector && flake8 --max-line-length=120 app.py

lint-go:
	cd event-processor && go vet ./...

lint-ts:
	cd api-gateway && npx eslint src/

logs:
	docker compose logs -f

restart:
	docker compose restart

status:
	docker compose ps

clean:
	docker compose down -v --rmi local
	rm -rf api-gateway/node_modules api-gateway/dist
	rm -rf event-collector/__pycache__ event-collector/.pytest_cache
