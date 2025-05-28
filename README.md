# AWS Bedrock Claude Proxy

[![Go Report Card](https://goreportcard.com/badge/github.com/mmhk/bedrock-claude-proxy)](https://goreportcard.com/report/github.com/mmhk/bedrock-claude-proxy)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Docker Pulls](https://img.shields.io/docker/pulls/mmhk/bedrock-claude-proxy)](https://hub.docker.com/r/mmhk/bedrock-claude-proxy)
[![GitHub issues](https://img.shields.io/github/issues/mmhk/bedrock-claude-proxy)](https://github.com/mmhk/bedrock-claude-proxy/issues)

Welcome to the `AWS Bedrock Claude Proxy` project! This project provides a seamless proxy service that translates AWS Bedrock API calls into the format used by the official Anthropic API, making it easier for clients that support the official API to integrate with AWS Bedrock.

## Introduction

`AWS Bedrock Claude Proxy` is designed to act as an intermediary between AWS Bedrock and clients that are built to interact with the official Anthropic API. By using this proxy, developers can leverage the robust infrastructure of AWS Bedrock while maintaining compatibility with existing Anthropic-based applications.

## Features

- **Seamless API Translation**: Converts AWS Bedrock API calls to Anthropic API format and vice versa.
- **Ease of Integration**: Minimal changes required for clients already using the official Anthropic API.
- **Scalability**: Built to handle high volumes of requests efficiently.
- **Security**: Ensures secure communication between clients and AWS Bedrock.
- **Flexible Configuration**: Supports various environment variables for customization.
- **Docker Support**: Easy deployment using Docker and Docker Compose.

## Getting Started

### Prerequisites

Before you begin, ensure you have met the following requirements:

- You have an AWS account with access to AWS Bedrock.
- You have Go installed on your local machine (version 1.20 or higher).
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
    AWS_BEDROCK_ACCESS_KEY=your_access_key
    AWS_BEDROCK_SECRET_KEY=your_secret_key
    AWS_BEDROCK_REGION=your_region
    WEB_ROOT=/path/to/web/root
    HTTP_LISTEN=0.0.0.0:3000
    API_KEY=your_api_key
    AWS_BEDROCK_MODEL_MAPPINGS="claude-instant-1.2=anthropic.claude-instant-v1,claude-2.0=anthropic.claude-v2,claude-2.1=anthropic.claude-v2:1,claude-3-sonnet-20240229=anthropic.claude-3-sonnet-20240229-v1:0,claude-3-opus-20240229=anthropic.claude-3-opus-20240229-v1:0,claude-3-haiku-20240307=anthropic.claude-3-haiku-20240307-v1:0"
    AWS_BEDROCK_ANTHROPIC_VERSION_MAPPINGS=2023-06-01=bedrock-2023-05-31
    AWS_BEDROCK_ANTHROPIC_DEFAULT_MODEL=anthropic.claude-v2
    AWS_BEDROCK_ANTHROPIC_DEFAULT_VERSION=bedrock-2023-05-31
    AWS_BEDROCK_ENABLE_OUTPUT_REASON=false
    AWS_BEDROCK_REASON_BUDGET_TOKENS=2048
    AWS_BEDROCK_ENABLE_COMPUTER_USE=false
    AWS_BEDROCK_DEBUG=false
    LOG_LEVEL=INFO

    # Cache config
    CACHE_DB_PATH=/path/to/cache.db
    CACHE_BUCKET_NAME=claude-proxy-cache
    CACHE_DEFAULT_EXPIRY_HOURS=1

    # Zoho Auth Config
    ZOHO_ALLOW_DOMAINS=domain.com
    ZOHO_CLIENT_ID=your_zoho_client_id
    ZOHO_CLIENT_SECRET=your_zoho_client_secret
    ZOHO_REDIRECT_URI=https://your-domain/oauth/callback
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

### Running with Docker Compose

1. **Build and run the containers:**

    ```bash
    docker-compose up -d
    ```

2. **Make API requests to the proxy:**

   Point your Anthropic API client to the proxy server. For example, if the proxy is running on `http://localhost:3000`, configure your client to use this base URL.

### Environment Variables

- `AWS_BEDROCK_ACCESS_KEY`: Your AWS Bedrock access key.
- `AWS_BEDROCK_SECRET_KEY`: Your AWS Bedrock secret access key.
- `AWS_BEDROCK_REGION`: Your AWS Bedrock region.
- `WEB_ROOT`: The root directory for web assets.
- `HTTP_LISTEN`: The address and port on which the server listens (e.g., `0.0.0.0:3000`).
- `API_KEY`: The API key for accessing the proxy.
- `AWS_BEDROCK_MODEL_MAPPINGS`: Mappings of model IDs to their respective Anthropic model versions.
- `AWS_BEDROCK_ANTHROPIC_VERSION_MAPPINGS`: Mappings of Bedrock versions to Anthropic versions.
- `AWS_BEDROCK_ANTHROPIC_DEFAULT_MODEL`: The default Anthropic model to use.
- `AWS_BEDROCK_ANTHROPIC_DEFAULT_VERSION`: The default Anthropic version to use.
- `AWS_BEDROCK_ENABLE_OUTPUT_REASON`: Enable output reason.
- `AWS_BEDROCK_REASON_BUDGET_TOKENS`: Budget tokens for output reason.
- `AWS_BEDROCK_ENABLE_COMPUTER_USE`: Enable computer use.
- `AWS_BEDROCK_DEBUG`: Enable debug mode.
- `LOG_LEVEL`: The logging level (e.g., `INFO`, `DEBUG`, `ERROR`).

### Cache Configuration
- `CACHE_DB_PATH`: Path to the cache database file
- `CACHE_BUCKET_NAME`: Name of the cache bucket (default: claude-proxy-cache)
- `CACHE_DEFAULT_EXPIRY_HOURS`: Default cache expiry time in hours (default: 1)

### Zoho Authentication
The proxy supports Zoho authentication for enhanced security. To enable Zoho authentication:

1. Configure your Zoho application settings in the Zoho Developer Console
2. Set the following environment variables:
   - `ZOHO_ALLOW_DOMAINS`: Comma-separated list of allowed email domains (e.g., mixmedia.com)
   - `ZOHO_CLIENT_ID`: Your Zoho application client ID
   - `ZOHO_CLIENT_SECRET`: Your Zoho application client secret
   - `ZOHO_REDIRECT_URI`: The OAuth redirect URI for your application

## Contributing

We welcome contributions! Please read our [Contributing Guide](CONTRIBUTING.md) to learn how you can help.

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.

---

Thank you for using `AWS Bedrock Claude Proxy`. We hope it makes your integration process smoother and more efficient!

---

# AWS Bedrock Claude Proxy

[![Go Report Card](https://goreportcard.com/badge/github.com/mmhk/bedrock-claude-proxy)](https://goreportcard.com/report/github.com/mmhk/bedrock-claude-proxy)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Docker Pulls](https://img.shields.io/docker/pulls/mmhk/bedrock-claude-proxy)](https://hub.docker.com/r/mmhk/bedrock-claude-proxy)
[![GitHub issues](https://img.shields.io/github/issues/mmhk/bedrock-claude-proxy)](https://github.com/mmhk/bedrock-claude-proxy/issues)

欢迎使用 `AWS Bedrock Claude Proxy`项目！本项目旨在提供一个无缝的代理服务，将 AWS Bedrock API 调用转换为官方 Anthropic API 使用的格式，使支持官方 API 的客户端更容易与 AWS Bedrock 集成。

## 简介

`AWS Bedrock Claude Proxy`被设计为 AWS Bedrock 和为与官方 Anthropic API 交互而构建的客户端之间的中介。通过使用此代理，开发人员可以利用 AWS Bedrock 的强大基础设施，同时保持与现有 Anthropic 应用程序的兼容性。

## 特性

- **无缝 API 转换**：将 AWS Bedrock API 调用转换为 Anthropic API 格式，反之亦然。
- **易于集成**：已使用官方 Anthropic API 的客户端只需进行最小的更改。
- **可扩展性**：构建为高效处理大量请求。
- **安全性**：确保客户端和 AWS Bedrock 之间的安全通信。
- **灵活配置**：支持各种环境变量进行自定义。
- **Docker 支持**：使用 Docker 和 Docker Compose 轻松部署。

## 入门指南

### 前提条件

在开始之前，请确保您满足以下要求：

- 您拥有可访问 AWS Bedrock 的 AWS 账户。
- 您的本地机器上安装了 Go（版本 1.20 或更高）。
- 您的本地机器上安装了 Docker（可选，但推荐）。
- 您对 REST API 有基本的了解。

### 安装

1. **克隆仓库：**

    ```bash
    git clone https://github.com/MMHK/bedrock-claude-proxy.git
    cd bedrock-claude-proxy
    ```

2. **安装依赖：**

    ```bash
    go mod tidy
    ```

### 配置

1. **在根目录创建一个 `.env` 文件并添加您的 AWS 凭证：**

    ```env
    AWS_BEDROCK_ACCESS_KEY=your_access_key
    AWS_BEDROCK_SECRET_KEY=your_secret_key
    AWS_BEDROCK_REGION=your_region
    WEB_ROOT=/path/to/web/root
    HTTP_LISTEN=0.0.0.0:3000
    API_KEY=your_api_key
    AWS_BEDROCK_MODEL_MAPPINGS="claude-instant-1.2=anthropic.claude-instant-v1,claude-2.0=anthropic.claude-v2,claude-2.1=anthropic.claude-v2:1,claude-3-sonnet-20240229=anthropic.claude-3-sonnet-20240229-v1:0,claude-3-opus-20240229=anthropic.claude-3-opus-20240229-v1:0,claude-3-haiku-20240307=anthropic.claude-3-haiku-20240307-v1:0"
    AWS_BEDROCK_ANTHROPIC_VERSION_MAPPINGS=2023-06-01=bedrock-2023-05-31
    AWS_BEDROCK_ANTHROPIC_DEFAULT_MODEL=anthropic.claude-v2
    AWS_BEDROCK_ANTHROPIC_DEFAULT_VERSION=bedrock-2023-05-31
    AWS_BEDROCK_ENABLE_OUTPUT_REASON=false
    AWS_BEDROCK_REASON_BUDGET_TOKENS=2048
    AWS_BEDROCK_ENABLE_COMPUTER_USE=false
    AWS_BEDROCK_DEBUG=false
    LOG_LEVEL=INFO

    # 快取配置
    CACHE_DB_PATH=/path/to/cache.db
    CACHE_BUCKET_NAME=claude-proxy-cache
    CACHE_DEFAULT_EXPIRY_HOURS=1

    # Zoho 認證配置
    ZOHO_ALLOW_DOMAINS=domain.com
    ZOHO_CLIENT_ID=your_zoho_client_id
    ZOHO_CLIENT_SECRET=your_zoho_client_secret
    ZOHO_REDIRECT_URI=https://your-domain/oauth/callback
    ```

## 使用方法

1. **构建项目：**

    ```bash
    go build -o bedrock-claude-proxy
    ```

2. **启动代理服务器：**

    ```bash
    ./bedrock-claude-proxy
    ```

3. **向代理发送 API 请求：**

   将您的 Anthropic API 客户端指向代理服务器。例如，如果代理运行在 `http://localhost:3000`，请将您的客户端配置为使用此基本 URL。

### 使用 Docker 运行

1. **构建 Docker 镜像：**

    ```bash
    docker build -t bedrock-claude-proxy .
    ```

2. **运行 Docker 容器：**

    ```bash
    docker run -d -p 3000:3000 --env-file .env bedrock-claude-proxy
    ```

3. **向代理发送 API 请求：**

   将您的 Anthropic API 客户端指向代理服务器。例如，如果代理运行在 `http://localhost:3000`，请将您的客户端配置为使用此基本 URL。

### 使用 Docker Compose 运行

1. **构建并运行容器：**

    ```bash
    docker-compose up -d
    ```

2. **向代理发送 API 请求：**

   将您的 Anthropic API 客户端指向代理服务器。例如，如果代理运行在 `http://localhost:3000`，请将您的客户端配置为使用此基本 URL。

### 环境变量

- `AWS_BEDROCK_ACCESS_KEY`：您的 AWS Bedrock 访问密钥。
- `AWS_BEDROCK_SECRET_KEY`：您的 AWS Bedrock 秘密访问密钥。
- `AWS_BEDROCK_REGION`：您的 AWS Bedrock 区域。
- `WEB_ROOT`：Web 资源的根目录。
- `HTTP_LISTEN`：服务器监听的地址和端口（例如，`0.0.0.0:3000`）。
- `API_KEY`：访问代理的 API 密钥。
- `AWS_BEDROCK_MODEL_MAPPINGS`：模型 ID 到其相应 Anthropic 模型版本的映射。
- `AWS_BEDROCK_ANTHROPIC_VERSION_MAPPINGS`：Bedrock 版本到 Anthropic 版本的映射。
- `AWS_BEDROCK_ANTHROPIC_DEFAULT_MODEL`：要使用的默认 Anthropic 模型。
- `AWS_BEDROCK_ANTHROPIC_DEFAULT_VERSION`：要使用的默认 Anthropic 版本。
- `AWS_BEDROCK_ENABLE_OUTPUT_REASON`：启用输出原因。
- `AWS_BEDROCK_REASON_BUDGET_TOKENS`：输出原因的预算令牌。
- `AWS_BEDROCK_ENABLE_COMPUTER_USE`：启用计算机使用。
- `AWS_BEDROCK_DEBUG`：启用调试模式。
- `LOG_LEVEL`：日志级别（例如，`INFO`、`DEBUG`、`ERROR`）。

### 快取配置
- `CACHE_DB_PATH`：快取資料庫檔案路徑
- `CACHE_BUCKET_NAME`：快取儲存桶名稱（預設：claude-proxy-cache）
- `CACHE_DEFAULT_EXPIRY_HOURS`：預設快取過期時間（小時，預設：1）

### Zoho 認證
代理支援 Zoho 認證以提升安全性。要啟用 Zoho 認證，請：

1. 在 Zoho 開發者控制台中配置您的應用程式設定
2. 設定以下環境變數：
   - `ZOHO_ALLOW_DOMAINS`：允許的電子郵件網域列表（以逗號分隔，如：domain.com）
   - `ZOHO_CLIENT_ID`：您的 Zoho 應用程式客戶端 ID
   - `ZOHO_CLIENT_SECRET`：您的 Zoho 應用程式客戶端密鑰
   - `ZOHO_REDIRECT_URI`：應用程式的 OAuth 重導向 URI

## 许可证

本项目采用 Apache 2.0 许可证 - 有关详细信息，请参阅 [LICENSE](LICENSE) 文件。

---

感谢您使用 `AWS Bedrock Claude 代理`。我们希望它能使您的集成过程更加顺畅和高效！
