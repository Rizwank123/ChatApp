# Use Official golang image as a base 
FROM golang:1.23 AS development
# Set working directory
WORKDIR  /usr/src/app

# Copy the current directory contents into the container at /app
COPY ../ .. 
# install dependencies
RUN go install ./...

# Run Application
CMD ["go", "run", "github.com/chatApp/cmd", "&"]