# Use the official Go image as the base image
FROM golang:1.21-alpine

RUN apk add --update nodejs npm 

# Set the working directory inside the container
WORKDIR /app

# Copy the source code into the container
COPY . .

RUN echo Installing Node dependencies
RUN npm i

# Build the Go application
RUN go build -o viscall .

# Command to run the application
CMD ["./viscall"]

