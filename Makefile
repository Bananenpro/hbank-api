API_BINARY_NAME=hbank-api
PAYMENT_PLANS_BINARY_NAME=hbank-payment-plans

API_MAIN=./cmd/hbank-api
PAYMENT_PLANS_MAIN=./cmd/hbank-payment-plans

make: build_api build_payment_plans

run_api:
	go run ${API_MAIN}

build_api:
	CGO_ENABLED=0 go build -o ./bin/${API_BINARY_NAME} ${API_MAIN}


run_payment_plans:
	go run ${PAYMENT_PLANS_MAIN}
 
build_payment_plans:
	CGO_ENABLED=0 go build -o ./bin/${PAYMENT_PLANS_BINARY_NAME} ${PAYMENT_PLANS_MAIN}
 
test:
	go test ./...
 
clean:
	go clean
	rm ./bin/${API_BINARY_NAME}
	rm ./bin/${PAYMENT_PLANS_BINARY_NAME}
	rmdir ./bin
