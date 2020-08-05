FROM golang:stretch as build-env

# Install minimum necessary dependencies
ENV PACKAGES curl make git libc-dev bash gcc
RUN apt-get update && apt-get upgrade -y && \
    apt-get install -y $PACKAGES

# Set working directory for the build
WORKDIR /go/src/github.com/ChainSafe/ethermint

# Add source files
COPY . .

# build Ethermint
RUN make build-ethermint-linux

# Final image
FROM golang:1.14 as final

WORKDIR /

RUN apt-get update

# Copy over binaries from the build-env
COPY --from=build-env /go/src/github.com/ChainSafe/ethermint/build/ethermintd /usr/bin/ethermintd
COPY --from=build-env /go/src/github.com/ChainSafe/ethermint/build/ethermintcli /usr/bin/ethermintcli
COPY --from=build-env /go/src/github.com/ChainSafe/ethermint/scripts/start.sh /

EXPOSE 26656 26657 1317 8545 8546

# Run ethermintd by default, omit entrypoint to ease using container with ethermintcli
ENTRYPOINT ["/bin/bash", "-c"]