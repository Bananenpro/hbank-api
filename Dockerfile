# Build frontend
FROM --platform=$BUILDPLATFORM node:18 AS frontend-build
ARG BUILDPLATFORM
WORKDIR /app
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build


# Build backend
FROM --platform=$BUILDPLATFORM golang:alpine AS backend-build
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN rm -rf frontend

COPY --from=frontend-build /app/dist ./frontend/dist

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o bin/h-bank ./cmd/h-bank

# Run
FROM alpine AS h-bank
ARG BUILDPLATFORM
WORKDIR /
COPY --from=backend-build /app/bin/h-bank /h-bank

EXPOSE 80

CMD [ "/h-bank" ]
