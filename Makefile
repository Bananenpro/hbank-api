API_BINARY_NAME=hbank-api
PAYMENT_PLANS_BINARY_NAME=hbank-payment-plans

API_MAIN=cmd/hbank-api/main.go
PAYMENT_PLANS_MAIN=cmd/hbank-payment-plans/main.go


run_api:
	go run ${API_MAIN}

build_api:
	go build -o ${API_BINARY_NAME} ${API_MAIN}


run_payment_plans:
	go run ${PAYMENT_PLANS_MAIN}
 
build_payment_plans:
	go build -o ${PAYMENT_PLANS_BINARY_NAME} ${PAYMENT_PLANS_MAIN}
 
test:
	go test ./...
 
clean:
	go clean
	rm ${API_BINARY_NAME}
	rm ${PAYMENT_PLANS_BINARY_NAME}
