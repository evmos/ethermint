FROM --platform=$BUILDPLATFORM golang:1.17 AS build-env

ARG TARGETOS TARGETARCH

# Install dependencies
RUN apt-get update
RUN apt-get install git

# Set up env vars
ENV COSMOS_BUILD_OPTIONS nostrip

# Set working directory for the build
WORKDIR /src 

COPY . .
RUN go mod download

# Make the binary
RUN env GOOS=$TARGETOS GOARCH=$TARGETARCH make build 

# Final image
FROM debian

# Install ca-certificates
RUN apt-get update
RUN apt-get install jq -y

WORKDIR /root

COPY docker/entrypoint.sh .
COPY init.sh .

# Copy over binaries from the build-env
COPY --from=build-env /src/build/ethermintd /usr/bin/ethermintd

EXPOSE 26656 26657

ENTRYPOINT ["./entrypoint.sh"]
CMD ["ethermintd"]
