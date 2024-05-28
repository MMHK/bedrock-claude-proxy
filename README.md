# AWS Bedrock Claude Proxy

[![Go Report Card](https://goreportcard.com/badge/github.com/mmhk/bedrock-claude-proxy)](https://goreportcard.com/report/github.com/yourusername/aws-bedrock-claude-proxy)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Docker Pulls](https://img.shields.io/docker/pulls/mmhk/bedrock-claude-proxy)](https://hub.docker.com/r/yourusername/aws-bedrock-claude-proxy)
[![GitHub issues](https://img.shields.io/github/issues/mmhk/bedrock-claude-proxy)](https://github.com/yourusername/aws-bedrock-claude-proxy/issues)

Welcome to the `AWS Bedrock Claude Proxy` project! This project aims to provide a seamless proxy service that translates AWS Bedrock API calls into the format used by the official Anthropic API, making it easier for clients that support the official API to integrate with AWS Bedrock.

**Note: This project is currently under development and is not stable. It is not recommended for use in production environments.**


## Introduction

`AWS Bedrock Claude Proxy` is designed to act as an intermediary between AWS Bedrock and clients that are built to interact with the official Anthropic API. By using this proxy, developers can leverage the robust infrastructure of AWS Bedrock while maintaining compatibility with existing Anthropic-based applications.

## Features

- **Seamless API Translation**: Converts AWS Bedrock API calls to Anthropic API format and vice versa.
- **Ease of Integration**: Minimal changes required for clients already using the official Anthropic API.
- **Scalability**: Built to handle high volumes of requests efficiently.
- **Security**: Ensures secure communication between clients and AWS Bedrock.

## Getting Started

### Prerequisites

Before you begin, ensure you have met the following requirements:

- You have an AWS account with access to AWS Bedrock.
- You have Go installed on your local machine (version 1.19 or higher+).
- You have Docker installed on your local machine (optional, but recommended).
- You have a basic understanding of REST APIs.

### Installation

1. **Clone the repository:**

    ```bash
    git clone https://github.com/MMHK/bedrock-claude-proxy.git
    cd bedrock-claude-proxy
    ```

2. **Install dependencies:**

    ```bash
    go mod tidy
    ```

### Configuration

1. **Create a `.env` file in the root directory and add your AWS credentials:**

    ```env
    AWS_BEDROCK_ACCESS_KEY=your_access_key_id
    AWS_BEDROCK_SECRET_KEY=your_secret_access_key
    AWS_BEDROCK_REGION=your_aws_region
    ```

## Usage

1. **Build the project:**

    ```bash
    go build -o bedrock-claude-proxy
    ```

2. **Start the proxy server:**

    ```bash
    ./bedrock-claude-proxy
    ```

3. **Make API requests to the proxy:**

   Point your Anthropic API client to the proxy server. For example, if the proxy is running on `http://localhost:3000`, configure your client to use this base URL.


### Running with Docker

1. **Build the Docker image:**

    ```bash
    docker build -t bedrock-claude-proxy .
    ```

2. **Run the Docker container:**

    ```bash
    docker run -d -p 3000:3000 --env-file .env bedrock-claude-proxy
    ```

3. **Make API requests to the proxy:**

   Point your Anthropic API client to the proxy server. For example, if the proxy is running on `http://localhost:3000`, configure your client to use this base URL.


### Environment

- AWS_ACCESS_KEY_ID: Your AWS access key ID.
- AWS_SECRET_ACCESS_KEY: Your AWS secret access key.
- AWS_REGION: Your AWS region.
- MODEL_ID: The model ID to use.
- MAX_TOKENS: The maximum number of tokens to generate.
- TEMPERATURE: The temperature to use.

## Contributing

We welcome contributions! Please read our [Contributing Guide](CONTRIBUTING.md) to learn how you can help.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

Thank you for using `AWS Bedrock Claude Proxy`. We hope it makes your integration process smoother and more efficient!