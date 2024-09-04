FROM mcr.microsoft.com/vscode/devcontainers/base:ubuntu

# Install Terraform
ARG TERRAFORM_VERSION=1.9.5
RUN wget -O- https://apt.releases.hashicorp.com/gpg | sudo gpg --dearmor -o /usr/share/keyrings/hasicorp-archive-keyring.gpg && \
    deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main && \
    sudo tee /etc/apt/sources.list.d/hashicorp.list && \ 
    apt update && sudo apt install terraform

# Install Go
ARG GO_VERSION=1.23
RUN apt-get update && apt-get install -y build-essential  && \
    wget -q -O /tmp/go.tgz https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf /tmp/go.tgz