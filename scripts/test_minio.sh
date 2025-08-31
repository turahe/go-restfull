#!/bin/bash

# Test MinIO connection and configuration
# This script tests if MinIO is running and accessible

set -e

echo "🔍 Testing MinIO connection..."

# Check if MinIO is running
echo "1. Checking if MinIO service is running..."
if curl -s http://localhost:9000/minio/health/live > /dev/null; then
    echo "✅ MinIO service is running on port 9000"
else
    echo "❌ MinIO service is not running on port 9000"
    echo "   Make sure to start MinIO with: docker-compose up minio"
    exit 1
fi

# Test MinIO console
echo "2. Checking MinIO console..."
if curl -s http://localhost:8900 > /dev/null; then
    echo "✅ MinIO console is accessible on port 8900"
else
    echo "❌ MinIO console is not accessible on port 8900"
fi

# Test S3 API endpoint
echo "3. Testing S3 API endpoint..."
if curl -s -I http://localhost:9000 > /dev/null; then
    echo "✅ S3 API endpoint is responding"
else
    echo "❌ S3 API endpoint is not responding"
fi

echo ""
echo "📋 MinIO Configuration Summary:"
echo "   Endpoint: http://localhost:9000"
echo "   Console: http://localhost:8900"
echo "   Access Key: minioadmin"
echo "   Secret Key: minioadmin"
echo "   Bucket: uploads"
echo ""

echo "🌐 You can access MinIO console at: http://localhost:8900"
echo "   Login with: minioadmin / minioadmin"
echo ""

echo "✅ MinIO connection test completed successfully!"
echo ""
echo "💡 If you're still having issues with file uploads, check:"
echo "   1. The storage.yaml configuration matches these credentials"
echo "   2. The 'uploads' bucket exists in MinIO"
echo "   3. The application can reach localhost:9000"
