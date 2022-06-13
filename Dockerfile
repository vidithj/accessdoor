# Used for build step
FROM golang:1.16.7-alpine3.14 as build

# Run updates on container
RUN apk update && apk add gcc libc-dev

# Create app directory and copy project to the app directory
RUN mkdir /app
ADD . /app

# Set the working directory to the location of the main file for the cloud version of the application
WORKDIR /app

# Ensure CGO is enabled
ENV CGO_ENABLED=1

# Build the final binary
RUN go build -mod=vendor -ldflags '-linkmode=external' -o dooraccess ./dooraccess/

# Build final image
FROM alpine:3.14

# Allow pipeline to pass in the git commit hash and build number and then make them
# available to the container via the environment
ARG git_commit
ARG build_number
ENV GIT_COMMIT=$git_commit
ENV BUILD_NUMBER=$build_number

# Run updates on container
RUN apk add --no-cache ca-certificates libc6-compat

# Copy the binary built from the build step to the new image
COPY --from=build /app/dooraccess /bin/dooraccess

# Make the binary executable
RUN chmod +x /bin/dooraccess

# Expose ports 8080 and 8082 for HTTP Server and Prometheus HTTP Server respectfully
EXPOSE 8080
EXPOSE 8082

# Launch the application binary
ENTRYPOINT ["/bin/dooraccess"]