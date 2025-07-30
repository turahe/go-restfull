#!/bin/bash

# Docker Hub Setup and Push Script
# This script helps you build and push Docker images to Docker Hub

set -e

# Configuration
DOCKER_USERNAME="turahe"
IMAGE_NAME="go-restfull"
FULL_IMAGE_NAME="${DOCKER_USERNAME}/${IMAGE_NAME}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if Docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker and try again."
        exit 1
    fi
    print_success "Docker is running"
}

# Function to check if logged into Docker Hub
check_docker_login() {
    if ! docker info | grep -q "Username"; then
        print_warning "Not logged into Docker Hub"
        print_status "Please run: docker login"
        exit 1
    fi
    print_success "Logged into Docker Hub"
}

# Function to build all images
build_images() {
    print_status "Building Docker images..."
    
    # Build production image
    print_status "Building production image..."
    docker build --target production -t "${FULL_IMAGE_NAME}:latest" .
    docker build --target production -t "${FULL_IMAGE_NAME}:prod" .
    
    # Build staging image
    print_status "Building staging image..."
    docker build --target staging -t "${FULL_IMAGE_NAME}:staging" .
    
    # Build development image
    print_status "Building development image..."
    docker build --target development -t "${FULL_IMAGE_NAME}:dev" .
    
    print_success "All images built successfully"
}

# Function to push all images
push_images() {
    print_status "Pushing Docker images to Docker Hub..."
    
    # Push all tags
    docker push "${FULL_IMAGE_NAME}:latest"
    docker push "${FULL_IMAGE_NAME}:prod"
    docker push "${FULL_IMAGE_NAME}:staging"
    docker push "${FULL_IMAGE_NAME}:dev"
    
    print_success "All images pushed successfully"
}

# Function to push specific image
push_specific_image() {
    local tag=$1
    print_status "Pushing ${FULL_IMAGE_NAME}:${tag}..."
    docker push "${FULL_IMAGE_NAME}:${tag}"
    print_success "Image ${FULL_IMAGE_NAME}:${tag} pushed successfully"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  build           Build all Docker images"
    echo "  push            Push all Docker images to Docker Hub"
    echo "  build-and-push  Build and push all Docker images"
    echo "  push-latest     Push only the latest image"
    echo "  push-prod       Push only the production image"
    echo "  push-staging    Push only the staging image"
    echo "  push-dev        Push only the development image"
    echo "  check           Check Docker and Docker Hub login status"
    echo "  help            Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 build-and-push"
    echo "  $0 push-latest"
    echo "  $0 check"
}

# Function to check prerequisites
check_prerequisites() {
    check_docker
    check_docker_login
}

# Main script logic
case "${1:-help}" in
    "build")
        check_docker
        build_images
        ;;
    "push")
        check_prerequisites
        push_images
        ;;
    "build-and-push")
        check_prerequisites
        build_images
        push_images
        ;;
    "push-latest")
        check_prerequisites
        push_specific_image "latest"
        ;;
    "push-prod")
        check_prerequisites
        push_specific_image "prod"
        ;;
    "push-staging")
        check_prerequisites
        push_specific_image "staging"
        ;;
    "push-dev")
        check_prerequisites
        push_specific_image "dev"
        ;;
    "check")
        check_prerequisites
        print_success "All checks passed"
        ;;
    "help"|*)
        show_usage
        ;;
esac 