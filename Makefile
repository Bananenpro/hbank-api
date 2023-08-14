OUT_DIR=bin
BIN_NAME=h-bank

.PHONY: backend frontend run-backend run-frontend init clean

backend: frontend
	CGO_ENABLED=0 go build -o ${OUT_DIR}/${BIN_NAME} ./cmd/hbank-api

frontend:
	npm run --prefix frontend build

frontend/dist:
	npm run --prefix frontend build

run-backend: frontend/dist
	@which wgo &> /dev/null || (echo "Installing wgo..." && go install github.com/bokwoon95/wgo@latest)
	wgo run -file config.json ./cmd/hbank-api

run-frontend:
	npm run --prefix frontend serve

init:
	go mod download
	npm install --prefix frontend
 
clean:
	go clean
	rm -rf frontend/dist
	rm ./${OUT_DIR}
