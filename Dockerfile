# Multi-stage build for ezhttp
# Stage 1: Build the application
FROM golang:1.24.4-alpine3.22 AS builder
ARG golang_version=1.24.4
ARG build_date
ARG build_git_hash

WORKDIR /usr/src/app
COPY . .

RUN go mod tidy && \
	go build -o "./ezhttp" \
	-ldflags="-s -w -X main.BuildDate=${build_date} -X main.BuildGoVersion=${golang_version} -X main.BuildGitHash=${build_git_hash}" \
	./main.go && \
	./ezhttp --version

# Stage 2: Runtime image
FROM alpine:3.22.0
ARG ARCH="none"

RUN apk add --no-cache curl && \
	addgroup -S appgroup && \
	adduser -S appuser -G appgroup --home "/usr/src/app" --no-create-home

WORKDIR /usr/src/app

# Copy binary from builder stage
COPY --from=builder /usr/src/app/ezhttp ./ezhttp
RUN chmod +x ./ezhttp

# Copy config and public files
COPY ./config.json .
COPY ./public ./public

USER appuser:appgroup

# Docker only
#HEALTHCHECK --interval=15s --retries=2 --start-period=5s --timeout=5s CMD curl --fail http://127.0.0.1:8080/health || exit 1

EXPOSE 8080
CMD ["/usr/src/app/ezhttp"]
