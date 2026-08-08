FROM golang:1.24-alpine AS build

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/webguard-server-agent ./cmd/webguard-server-agent

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/webguard-server-agent /usr/local/bin/webguard-server-agent

USER nonroot:nonroot

ENTRYPOINT ["/usr/local/bin/webguard-server-agent"]
