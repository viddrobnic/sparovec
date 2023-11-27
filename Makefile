.PHONY: generate
generate:
	@pnpm tailwindcss -i ./assets/_global.css -o ./assets/global.css --minify

.PHONY: build
build: generate
	@go build
