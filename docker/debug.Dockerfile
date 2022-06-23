FROM golang:1.17 AS build-env

# Install dependencies
RUN apt-get update
RUN apt-get install git

# Set up env vars
ENV COSMOS_BUILD_OPTIONS nostrip

# Set working directory for the build
WORKDIR /src 

COPY . .
RUN go mod download

# Build Delve
RUN go install github.com/go-delve/delve/cmd/dlv@latest

# Make the binary
RUN LEDGER_ENABLED=false make build 

# Final image
FROM debian

# Install ca-certificates
RUN apt-get update
RUN apt-get install jq -y

WORKDIR /root

COPY docker/entrypoint-debug.sh .
COPY init.sh .

# Copy over binaries from the build-env
COPY --from=build-env /src/build/ethermintd /usr/bin/ethermintd
COPY --from=build-env /go/bin/dlv /usr/bin/dlv

EXPOSE 26656 26657

ENTRYPOINT ["./entrypoint-debug.sh"]
CMD ["ethermintd"]