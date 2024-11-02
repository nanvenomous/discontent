run: templ
	go run ./...

templ:
	templ generate ./ui/...

up:
	docker compose -f docker-compose.yml up --build --force-recreate
