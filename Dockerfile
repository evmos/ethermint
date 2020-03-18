FROM golang:alpine AS build-env

# Set up dependencies
ENV PACKAGES git build-base

# Set working directory for the build
WORKDIR /go/src/github.com/Chainsafe/ethermint

# Install dependencies
RUN apk add --update $PACKAGES

# Add source files
COPY . .

# Make the binary
RUN make build

# Final image
FROM alpine

# Install ca-certificates
RUN apk add --update ca-certificates
WORKDIR /root

# Copy over binaries from the build-env
COPY --from=build-env /go/src/github.com/Chainsafe/ethermint/build/emintd /usr/bin/emintd
COPY --from=build-env /go/src/github.com/Chainsafe/ethermint/build/emintcli /usr/bin/emintcli

# Run emintd by default
CMD ["emintd"]
