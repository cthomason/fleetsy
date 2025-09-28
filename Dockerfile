FROM golang:1.25-alpine AS builder

# Set the working directory inside the container.
WORKDIR /opt/fleetsy

# copy go.mod and grab the dependencies
COPY go.mod ./
RUN go mod download

# copy everything else
COPY . .

# build the binary
# -o /server specifies the output file name and location.
# CGO_ENABLED=0 disables Cgo, which is needed for a static binary.
# -ldflags="-s -w" strips debugging information, making the binary smaller.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /uplink .

# since we're after only a self-contained binary here, use the scratch image
FROM scratch

# copy the binary from before
COPY --from=builder /fleetsy /fleetsy

# expose the UDP port
EXPOSE 8080

# run the application
ENTRYPOINT ["/fleetsy"]