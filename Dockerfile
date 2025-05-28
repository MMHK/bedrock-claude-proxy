FROM golang:1.20-alpine as builder

# Add Maintainer Info
LABEL maintainer="Sam Zhou <sam@mixmedia.com>"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go version \
 && export GO111MODULE=on \
 && export GOPROXY=https://goproxy.io,direct \
 && go mod vendor \
 && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bedrock-claude-proxy \
 && echo "{}" > config.json


######## Start a new stage from scratch #######
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata dumb-init gettext-envsubst \
    && update-ca-certificates \
    && envsubst --version \
    && rm -rf /var/cache/apk/*

WORKDIR /app


# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/bedrock-claude-proxy .
COPY --from=builder /app/webroot ./webroot
COPY --from=builder /app/config.json .

ENV HTTP_LISTEN=0.0.0.0:3000 \
 WEB_ROOT=/app/webroot \
 API_KEY= \
 AWS_BEDROCK_ACCESS_KEY= \
 AWS_BEDROCK_SECRET_KEY= \
 AWS_BEDROCK_REGION=us-west-2 \
 AWS_BEDROCK_MODEL_MAPPINGS="claude-3-5-sonnet-20240620=anthropic.claude-3-5-sonnet-20240620-v1:0,claude-3-5-sonnet-latest=anthropic.claude-3-5-sonnet-20241022-v2:0,claude-3-5-sonnet-20241022=anthropic.claude-3-5-sonnet-20241022-v2:0,claude-3-5-haiku-20241022=anthropic.claude-3-5-haiku-20241022-v1:0" \
 AWS_BEDROCK_ANTHROPIC_VERSION_MAPPINGS="2023-06-01=bedrock-2023-05-31" \
 AWS_BEDROCK_ANTHROPIC_DEFAULT_MODEL="anthropic.claude-3-5-haiku-20241022-v1:0" \
 AWS_BEDROCK_ANTHROPIC_DEFAULT_VERSION=bedrock-2023-05-31 \
 AWS_BEDROCK_ENABLE_OUTPUT_REASON=false \
 AWS_BEDROCK_REASON_BUDGET_TOKENS=1024 \
 AWS_BEDROCK_ENABLE_COMPUTER_USE=false \
 AWS_BEDROCK_DEBUG=false \
 CACHE_DB_PATH=/app/cache.db \
 CACHE_BUCKET_NAME=bedrock-claude-proxy-cache \
 CACHE_DEFAULT_EXPIRY_HOURS=720 \
 ZOHO_ALLOW_DOMAINS= \
 ZOHO_CLIENT_ID= \
 ZOHO_CLIENT_SECRET= \
 ZOHO_REDIRECT_URI= \
 ZOHO_ALLOW_DOMAINS= \
 LOG_LEVEL=INFO

EXPOSE 3000

ENTRYPOINT ["dumb-init", "--"]

CMD envsubst < /app/config.json > /app/temp.json \
 && /app/bedrock-claude-proxy -c /app/temp.json

