# Docker Hub Setup and Push Guide

This guide will help you set up and push your Docker images to Docker Hub.

## 🚀 Quick Start

### 1. Docker Hub Account Setup

1. **Create Docker Hub Account**
   - Go to [Docker Hub](https://hub.docker.com/)
   - Sign up for a free account
   - Verify your email address

2. **Create Repository**
   - Click "Create Repository"
   - Repository name: `go-restfull`
   - Description: "Go RESTful API with Hexagonal Architecture"
   - Set visibility (Public/Private)
   - Click "Create"

### 2. Local Setup

#### Login to Docker Hub
```bash
docker login
# Enter your Docker Hub username and password
```

#### Build Images
```bash
# Build all images for Docker Hub
make docker-hub-build

# Or use the script
./scripts/docker-hub-setup.sh build
```

#### Push Images
```bash
# Push all images
make docker-hub-push

# Or use the script
./scripts/docker-hub-setup.sh build-and-push
```

## 📋 Available Commands

### Makefile Commands
```bash
# Build all Docker Hub images
make docker-hub-build

# Push all images to Docker Hub
make docker-hub-push

# Build and push all images
make docker-hub-build-and-push

# Push specific images
make docker-hub-push-latest
make docker-hub-push-prod
make docker-hub-push-staging
make docker-hub-push-dev
```

### Script Commands
```bash
# Check prerequisites
./scripts/docker-hub-setup.sh check

# Build all images
./scripts/docker-hub-setup.sh build

# Push all images
./scripts/docker-hub-setup.sh push

# Build and push all images
./scripts/docker-hub-setup.sh build-and-push

# Push specific images
./scripts/docker-hub-setup.sh push-latest
./scripts/docker-hub-setup.sh push-prod
./scripts/docker-hub-setup.sh push-staging
./scripts/docker-hub-setup.sh push-dev
```

## 🏷️ Image Tags

The following images will be pushed to Docker Hub:

- `turahe/go-restfull:latest` - Production optimized (latest)
- `turahe/go-restfull:prod` - Production environment
- `turahe/go-restfull:staging` - Staging environment
- `turahe/go-restfull:dev` - Development environment

## 🔧 Configuration

### Docker Hub Repository
- **Username**: `turahe`
- **Repository**: `go-restfull`
- **Full Image Name**: `turahe/go-restfull`

### Update Configuration
If you need to change the Docker Hub username or repository name:

1. **Update Makefile**
   ```makefile
   # Change these lines in Makefile
   docker build --target production -t YOUR_USERNAME/YOUR_REPO:latest ..
   ```

2. **Update Script**
   ```bash
   # Change these lines in scripts/docker-hub-setup.sh
   DOCKER_USERNAME="YOUR_USERNAME"
   IMAGE_NAME="YOUR_REPO"
   ```

## 🤖 Automated Deployment

### GitHub Actions

The repository includes a GitHub Actions workflow that automatically builds and pushes images to Docker Hub.

#### Setup GitHub Secrets

1. Go to your GitHub repository
2. Navigate to Settings → Secrets and variables → Actions
3. Add the following secrets:
   - `DOCKERHUB_USERNAME`: Your Docker Hub username
   - `DOCKERHUB_TOKEN`: Your Docker Hub access token

#### Create Docker Hub Access Token

1. Go to Docker Hub → Account Settings → Security
2. Click "New Access Token"
3. Give it a name (e.g., "GitHub Actions")
4. Copy the token and add it to GitHub secrets

#### Trigger Deployment

The workflow will automatically run on:
- Push to `main` or `master` branch
- Push of version tags (e.g., `v1.0.0`)
- Pull requests (build only, no push)

## 📊 Image Information

### Image Sizes
- **Production**: ~50MB
- **Staging**: ~50MB
- **Development**: ~500MB

### Base Images
- **Production/Staging**: `alpine:latest`
- **Development**: `golang:1.24.5-alpine`

### Multi-Architecture Support
The GitHub Actions workflow builds for:
- `linux/amd64`
- `linux/arm64`

## 🔍 Verification

### Check Local Images
```bash
# List all built images
docker images | grep turahe/go-restfull
```

### Test Images Locally
```bash
# Test production image
docker run --rm -p 8000:8000 turahe/go-restfull:latest

# Test with external config
docker run --rm -p 8000:8000 \
  -v $(pwd)/custom.yml:/configs/user.yml \
  -e CONFIG_PATH=/configs/user.yml \
  turahe/go-restfull:latest
```

### Check Docker Hub
After pushing, verify your images are available at:
- https://hub.docker.com/r/turahe/go-restfull

## 📝 README for Docker Hub

The `README.dockerhub.md` file contains the documentation that will be displayed on Docker Hub. This includes:

- Quick start instructions
- Configuration examples
- Usage examples
- Troubleshooting guide
- Feature list

## 🔄 Update Process

### Manual Update
1. Make your code changes
2. Build images: `make docker-hub-build`
3. Push images: `make docker-hub-push`

### Automated Update
1. Push changes to `main` branch
2. GitHub Actions will automatically build and push
3. Images will be available on Docker Hub

### Version Tags
```bash
# Create a version tag
git tag v1.0.0
git push origin v1.0.0

# This will trigger GitHub Actions to build and push with version tags
```

## 🛠️ Troubleshooting

### Common Issues

#### Docker Login Failed
```bash
# Clear Docker credentials
docker logout

# Login again
docker login
```

#### Permission Denied
```bash
# Make script executable
chmod +x scripts/docker-hub-setup.sh
```

#### Build Failed
```bash
# Check Docker is running
docker info

# Clean up Docker cache
docker system prune -a
```

#### Push Failed
```bash
# Check Docker Hub login
docker info | grep Username

# Verify repository exists
# Go to https://hub.docker.com/r/turahe/go-restfull
```

### Logs and Debugging
```bash
# Check build logs
docker build --target production -t turahe/go-restfull:latest . 2>&1 | tee build.log

# Check push logs
docker push turahe/go-restfull:latest 2>&1 | tee push.log
```

## 📚 Additional Resources

- [Docker Hub Documentation](https://docs.docker.com/docker-hub/)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker Build Documentation](https://docs.docker.com/engine/reference/commandline/build/)

## 🎯 Next Steps

1. **Set up Docker Hub account and repository**
2. **Configure GitHub secrets for automated deployment**
3. **Test the build and push process**
4. **Update the README.dockerhub.md if needed**
5. **Push your first images to Docker Hub**

---

**Happy Dockerizing! 🐳** 