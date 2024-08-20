# Use the official Golang image as the base image
FROM golang:1.22-alpine

# Set the working directory in the container
ENV APP_HOME myapp
RUN mkdir -p "$APP_HOME"
WORKDIR "$APP_HOME"

# Copy the go.mod and go.sum files into the container
COPY src/go.mod src/go.sum ./

# Download the dependencies
RUN go mod download

# Copy the rest of the application code into the container
COPY src/*.go ./

#Copy html file
COPY src/html.html ./

# Build the Go application
RUN go build -o main

# Expose the port that the application will listen on
EXPOSE 1323

# Run the Go application
CMD ["./main"]