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
RUN make tools deps build

# Final image
FROM alpine

# Install ca-certificates
RUN apk add --update ca-certificates
WORKDIR /root

# Copy over binaries from the build-env
COPY --from=build-env /go/src/github.com/cosmos/ethermint/build/emintd /usr/bin/emintd

# Run emintd by default
CMD ["emintd"]
