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

# 3. Generate Secret Key
if [ ! -f .env ]; then
    echo -e "\n${BLUE}Generating secure SEARXNG_SECRET...${NC}"
    SECRET_KEY=$(openssl rand -hex 32)
    echo "SEARXNG_SECRET=${SECRET_KEY}" > .env
    echo -e "${GREEN}Created .env file with a new secret key.${NC}"
else
    echo -e "\n${GREEN}Found existing .env file. Using existing secrets.${NC}"
fi

# 4. Setup SearXNG Configuration
echo -e "\n${BLUE}Setting up SearXNG configurations...${NC}"
mkdir -p searxng
cp searxng-settings.yml searxng/settings.yml
# Fix permissions for Docker volume
chmod -R 777 searxng

# 5. Start SearXNG Backend
echo -e "\n${BLUE}Starting SearXNG backend via Docker Compose...${NC}"
docker compose -f docker-compose.searxng.yml --env-file .env up -d

# 6. Build the MCP Server using Docker
echo -e "\n${BLUE}Building the SearchInlet MCP Server...${NC}"
mkdir -p bin

# We use the official golang image to compile the binary so the host doesn't need Go installed
# Building with CGO_ENABLED=1 requires a Debian-based image for glibc compatibility
docker run --rm -v "$(pwd):/app" -w /app golang:1.24-bookworm sh -c "
    apt-get update && apt-get install -y git gcc libc6-dev && \
    git config --global --add safe.directory /app && \
    go mod download && \
    CGO_ENABLED=1 GOOS=linux go build -buildvcs=false -o bin/mcp-server-linux ./cmd/mcp-server && \
    chmod 777 bin/mcp-server-linux
" && rm -rf /var/lib/apt/lists/*

echo -e "${GREEN}Build complete! Binary located at: bin/mcp-server-linux${NC}"

# 7. Verify Deployment
echo -e "\n${BLUE}Verifying SearXNG availability...${NC}"
echo "Waiting for SearXNG to boot (10 seconds)..."
sleep 10

if curl -s -G --data-urlencode 'q=test' --data-urlencode 'format=json' http://localhost:8080/search | grep -q "test"; then
    echo -e "${GREEN}SearXNG is up and responding to JSON API requests!${NC}"
else
    echo -e "${RED}Warning: SearXNG might not be ready yet. Check logs with: docker compose -f docker-compose.searxng.yml logs searxng${NC}"
fi

INSTALL_DIR=$(pwd)

echo -e "\n${GREEN}🎉 Installation Complete!${NC}"
echo -e "You can start the MCP server locally with: \n  cd $INSTALL_DIR && SEARXNG_URL=http://localhost:8080/search ./bin/mcp-server-linux"
echo -e "\nOr connect your AI Agent via SSH:"
echo "{
  \"mcpServers\": {
    \"searchinlet\": {
      \"command\": \"ssh\",
      \"args\": [
        \"user@your-server-ip\",
        \"SEARXNG_URL=http://localhost:8080/search $INSTALL_DIR/bin/mcp-server-linux\"
      ]
    }
  }
}"