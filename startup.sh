#!/bin/bash

# Port 56
# .env volume
# keys.json volume
if docker run -d --name ai-backend -p 56:8080 \
  -v $(pwd)/.env:/app/.env \
  -v $(pwd)/keys.json:/app/keys.json \
  ghcr.io/arinji2/ai-backend:latest; then
  echo "Container 'ai-backend' started successfully."
else
  echo "Failed to start the container."
fi

