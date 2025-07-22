#!/bin/bash

# LocalStack Test Helper Script
# This script helps with manual testing of LocalStack functionality

set -e

CONTAINER_NAME="terraform-provider-genesyscloud-localstack"
BUCKET_NAME="testbucket"
ENDPOINT_URL="http://localhost:4566"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        log_error "Docker is not running. Please start Docker and try again."
        exit 1
    fi
    log_info "Docker is running"
}

# Check if AWS CLI is installed
check_aws_cli() {
    if ! command -v aws &> /dev/null; then
        log_error "AWS CLI is not installed. Please install AWS CLI and try again."
        exit 1
    fi
    log_info "AWS CLI is available"
}

# Start LocalStack
start_localstack() {
    log_info "Starting LocalStack..."
    
    # Remove existing container if it exists
    if docker ps -a --filter "name=$CONTAINER_NAME" --format "{{.ID}}" | grep -q .; then
        log_warn "Removing existing container..."
        docker rm -f $CONTAINER_NAME > /dev/null 2>&1 || true
    fi
    
    # Start new container
    docker run -d \
        --name $CONTAINER_NAME \
        -p 4566:4566 \
        -e SERVICES=s3 \
        -e DEBUG=1 \
        localstack/localstack:latest
    
    log_info "Waiting for LocalStack to be ready..."
    for i in {1..30}; do
        if curl -f http://localhost:4566/_localstack/health > /dev/null 2>&1; then
            log_info "LocalStack is ready!"
            return 0
        fi
        sleep 2
    done
    
    log_error "LocalStack failed to start within 60 seconds"
    return 1
}

# Stop LocalStack
stop_localstack() {
    log_info "Stopping LocalStack..."
    docker stop $CONTAINER_NAME > /dev/null 2>&1 || true
    docker rm $CONTAINER_NAME > /dev/null 2>&1 || true
    log_info "LocalStack stopped"
}

# Create S3 bucket
create_bucket() {
    log_info "Creating S3 bucket: $BUCKET_NAME"
    aws s3api create-bucket \
        --bucket $BUCKET_NAME \
        --region us-east-1 \
        --endpoint-url $ENDPOINT_URL 2>/dev/null || log_warn "Bucket may already exist"
}

# Upload test file
upload_file() {
    local file_path="$1"
    local object_key="$2"
    
    if [ ! -f "$file_path" ]; then
        log_error "File not found: $file_path"
        return 1
    fi
    
    log_info "Uploading $file_path to s3://$BUCKET_NAME/$object_key"
    aws s3 cp "$file_path" "s3://$BUCKET_NAME/$object_key" --endpoint-url $ENDPOINT_URL
}

# Clean up S3 bucket
cleanup_bucket() {
    log_info "Cleaning up S3 bucket: $BUCKET_NAME"
    aws s3 rm "s3://$BUCKET_NAME" --recursive --endpoint-url $ENDPOINT_URL 2>/dev/null || true
    aws s3api delete-bucket --bucket $BUCKET_NAME --endpoint-url $ENDPOINT_URL 2>/dev/null || true
}

# Show status
show_status() {
    log_info "LocalStack Status:"
    if docker ps --filter "name=$CONTAINER_NAME" --format "{{.Names}}: {{.Status}}" | grep -q .; then
        docker ps --filter "name=$CONTAINER_NAME" --format "{{.Names}}: {{.Status}}"
    else
        log_warn "LocalStack container is not running"
    fi
    
    log_info "S3 Buckets:"
    aws s3 ls --endpoint-url $ENDPOINT_URL 2>/dev/null || log_warn "No buckets found or LocalStack not ready"
}

# Main function
main() {
    case "${1:-help}" in
        "start")
            check_docker
            check_aws_cli
            start_localstack
            ;;
        "stop")
            stop_localstack
            ;;
        "status")
            show_status
            ;;
        "create-bucket")
            check_aws_cli
            create_bucket
            ;;
        "upload")
            if [ -z "$2" ] || [ -z "$3" ]; then
                log_error "Usage: $0 upload <file_path> <object_key>"
                exit 1
            fi
            check_aws_cli
            upload_file "$2" "$3"
            ;;
        "cleanup")
            check_aws_cli
            cleanup_bucket
            ;;
        "test")
            check_docker
            check_aws_cli
            start_localstack
            create_bucket
            log_info "LocalStack is ready for testing!"
            log_info "Endpoint: $ENDPOINT_URL"
            log_info "Bucket: $BUCKET_NAME"
            log_info "Run '$0 stop' to clean up"
            ;;
        "help"|*)
            echo "Usage: $0 {start|stop|status|create-bucket|upload|cleanup|test}"
            echo ""
            echo "Commands:"
            echo "  start         - Start LocalStack container"
            echo "  stop          - Stop and remove LocalStack container"
            echo "  status        - Show LocalStack and S3 status"
            echo "  create-bucket - Create S3 bucket"
            echo "  upload        - Upload file to S3 (usage: upload <file_path> <object_key>)"
            echo "  cleanup       - Clean up S3 bucket"
            echo "  test          - Start LocalStack and create bucket for testing"
            echo "  help          - Show this help message"
            ;;
    esac
}

main "$@" 