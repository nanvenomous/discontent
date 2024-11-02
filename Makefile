run: templ
	go run ./...

templ:
	templ generate ./ui/...

revive:
	revive -config .revive.toml -formatter stylish ./...

up:
	docker compose -f docker-compose.yml up --build --force-recreate
