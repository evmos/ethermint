FROM golang:alpine AS build-env

# Set up dependencies
ENV PACKAGES make git curl build-base

# Set working directory for the build
WORKDIR /go/src/github.com/cosmos/ethermint

# Install dependencies
RUN apk add --update $PACKAGES

# Add source files
COPY . .

# Make the binary
RUN make update-tools get-vendor-deps build

# Final image
FROM alpine:edge

# Install ca-certificates
RUN apk add --update ca-certificates
WORKDIR /root

# Copy over binaries from the build-env
COPY --from=build-env /go/src/github.com/cosmos/ethermint/build/ethermint /usr/bin/ethermint

# Run ethermint by default
CMD ["ethermint"]
