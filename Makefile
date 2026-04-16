setup:
	test -f .env || cp samples/.env.sample .env
	test -f docker-compose.yaml || cp samples/docker-compose.yaml.sample docker-compose.yaml
	go mod download && go mod tidy
