FROM golang:1.26-alpine AS build
ENV RUN apk add --no-cache git
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go vet ./... && CGO_ENABLED=0 go test ./...
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/gw ./cmd/gw
FROM scratch
COPY --from=build /out/gw /gw
USER 1000:1000
EXPOSE 25565
ENTRYPOINT ["/gw"]
