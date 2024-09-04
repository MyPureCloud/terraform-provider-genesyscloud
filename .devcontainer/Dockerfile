FROM mcr.microsoft.com/vscode/devcontainers/base:ubuntu

# Install Terraform
ARG TERRAFORM_VERSION=1.9.5
RUN apt-get update && apt-get install -y wget unzip && \
    wget -q -O /tmp/terraform.zip https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
    unzip -d /usr/local/bin /tmp/terraform.zip

# Install Go
ARG GO_VERSION=1.23
RUN apt-get update && apt-get install -y build-essential  && \
    wget -q -O /tmp/go.tgz https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf /tmp/go.tgz