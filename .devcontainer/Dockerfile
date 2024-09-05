FROM mcr.microsoft.com/vscode/devcontainers/base:ubuntu

# Install Genesys Cloud CLI
RUN curl -s https://sdk-cdn.mypurecloud.com/external/go-cli/linux/dl/install.sh | sudo bash

# Install the Archy CLI
RUN cd ~ && \
    mkdir archy && \ 
    curl https://sdk-cdn.mypurecloud.com/archy/latest/archy-linux.zip > archy-linux.zip && \
    unzip archy-linux.zip -d ~/archy && \ 
    cd ~