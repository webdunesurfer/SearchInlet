#!/usr/bin/env bash

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}🌊 Welcome to the SearchInlet Installer${NC}"
echo "This script will deploy SearXNG and build the MCP Server."

# 1. Check dependencies
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Error: Docker is not installed. Please install Docker first.${NC}"
    exit 1
fi

if ! docker compose version &> /dev/null; then
    echo -e "${RED}Error: Docker Compose is not available. Please install it.${NC}"
    exit 1
fi

if ! command -v git &> /dev/null; then
    echo -e "${RED}Error: Git is not installed. Please install Git first.${NC}"
    exit 1
fi

# 2. Clone repository if not running inside it
if [ ! -f "docker-compose.searxng.yml" ]; then
    if [ ! -d "SearchInlet" ]; then
        echo -e "\n${BLUE}Cloning SearchInlet repository...${NC}"
        git clone https://github.com/webdunesurfer/SearchInlet.git
    else
        echo -e "\n${BLUE}SearchInlet directory already exists. Pulling latest changes...${NC}"
        cd SearchInlet && git pull && cd ..
    fi
    cd SearchInlet
fi

# 3. Generate Secrets
if [ ! -f .env ]; then
    echo -e "\n${BLUE}Generating secure credentials...${NC}"
    SEARXNG_SECRET=$(openssl rand -hex 32)
    # Generate a random 12-character alphanumeric password
    ADMIN_PASSWORD=$(openssl rand -base64 12 | tr -dc 'a-zA-Z0-9' | head -c 12)
    
    echo "SEARXNG_SECRET=${SEARXNG_SECRET}" > .env
    echo "ADMIN_USER=admin" >> .env
    echo "ADMIN_PASSWORD=${ADMIN_PASSWORD}" >> .env
    echo -e "${GREEN}Created .env file with new secret key and admin credentials.${NC}"
else
    echo -e "\n${GREEN}Found existing .env file. Using existing secrets.${NC}"
    ADMIN_PASSWORD=$(grep ADMIN_PASSWORD .env | cut -d '=' -f2)
fi

# 4. Setup SearXNG Configuration
echo -e "\n${BLUE}Setting up SearXNG configurations...${NC}"
mkdir -p searxng
cp searxng-settings.yml searxng/settings.yml
# Use Docker to fix permissions as root since the folder might be owned by root from previous runs
docker run --rm -v "$(pwd):/data" alpine chmod -R 777 /data/searxng

# 5. Start SearXNG Backend
echo -e "\n${BLUE}Starting SearXNG backend via Docker Compose...${NC}"
docker compose -f docker-compose.searxng.yml --env-file .env up -d

# 6. Build the MCP Server using Docker
echo -e "\n${BLUE}Building the SearchInlet MCP Server...${NC}"
mkdir -p bin

# We use the official golang image to compile the binary so the host doesn't need Go installed
# Building with CGO_ENABLED=0 produces a static binary that works on any Linux system
docker run --rm -v "$(pwd):/app" -w /app golang:1.24-alpine sh -c "
    apk add --no-cache git && \
    git config --global --add safe.directory /app && \
    go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o bin/mcp-server-linux ./cmd/mcp-server && \
    chmod 777 bin/mcp-server-linux
"

echo -e "${GREEN}Build complete! Binary located at: bin/mcp-server-linux${NC}"

# 7. Verify Deployment
echo -e "\n${BLUE}Verifying SearXNG availability...${NC}"
echo "Waiting for SearXNG to boot (10 seconds)..."
sleep 10

if curl -s -G --data-urlencode 'q=test' --data-urlencode 'format=json' http://localhost:8088/search | grep -q "test"; then
    echo -e "${GREEN}SearXNG is up and responding to JSON API requests!${NC}"
else
    echo -e "${RED}Warning: SearXNG might not be ready yet. Check logs with: docker compose -f docker-compose.searxng.yml logs searxng${NC}"
fi

INSTALL_DIR=$(pwd)

echo -e "\n${GREEN}🎉 Installation Complete!${NC}"
echo -e "${BLUE}--------------------------------------------------${NC}"
echo -e "  Admin Dashboard: ${GREEN}https://searchinlet.com${NC}"
echo -e "  Username:        ${GREEN}admin${NC}"
echo -e "  Password:        ${GREEN}${ADMIN_PASSWORD}${NC}"
echo -e "${BLUE}--------------------------------------------------${NC}"

echo -e "\n${BLUE}🚀 How to connect your AI Agent (Cursor / Claude):${NC}"
echo -e "\n${GREEN}Option A: SSE (Recommended)${NC}"
echo -e "Use the URL: ${BLUE}https://searchinlet.com/sse${NC}"
echo -e "And add the header: ${BLUE}Authorization: Bearer sk-YOUR_TOKEN_FROM_DASHBOARD${NC}"

echo -e "\n${GREEN}Option B: SSH (Legacy)${NC}"
echo -e "Copy this into your MCP settings:"
echo "{
  \"mcpServers\": {
    \"searchinlet\": {
      \"command\": \"ssh\",
      \"args\": [
        \"user@your-server-ip\",
        \"SEARXNG_URL=http://localhost:8088/search $INSTALL_DIR/bin/mcp-server-linux\"
      ]
    }
  }
}"
