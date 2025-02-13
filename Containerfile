FROM golang:alpine3.21 AS build

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /frnred

# Borrow a good package (without it Fiber w/ Prefork gets killed instantly by the kernel. Feel free to remove it if you don't need prefork)
RUN apk add --no-cache dumb-init

# Copy the binary to a scrath image (this decreases the image size from 0.5Gb to ~20Mb)
FROM scratch AS run
COPY --from=build /frnred /frnred

# dumb-init: remove this
COPY --from=build ["/usr/bin/dumb-init", "/usr/bin/dumb-init"] 

EXPOSE 8080

# dumb-init: remove this
ENTRYPOINT ["/usr/bin/dumb-init", "--"]

# Run
CMD ["/frnred"]
