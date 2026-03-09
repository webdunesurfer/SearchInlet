#!/usr/bin/env bash

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}🌊 Welcome to the SearchInlet Installer${NC}"
echo "This script will deploy the entire SearchInlet stack via Docker."

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

# 3. Configuration
if [ ! -f .env ]; then
    echo -e "\n${BLUE}Configuring your SearchInlet instance...${NC}"
    
    # Domain setup
    read -p "Enter your domain name (e.g., search.example.com) or press Enter for localhost: " DOMAIN
    if [ -z "$DOMAIN" ]; then
        DOMAIN="localhost"
    fi

    read -p "Enter your admin email (for SSL certificates): " ADMIN_EMAIL
    if [ -z "$ADMIN_EMAIL" ]; then
        ADMIN_EMAIL="admin@example.com"
    fi

    echo -e "\n${BLUE}Generating secure credentials...${NC}"
    SEARXNG_SECRET=$(openssl rand -hex 32)
    ADMIN_PASSWORD=$(openssl rand -base64 12 | tr -dc 'a-zA-Z0-9' | head -c 12)
    
    echo "DOMAIN=${DOMAIN}" > .env
    echo "ADMIN_EMAIL=${ADMIN_EMAIL}" >> .env
    echo "SEARXNG_SECRET=${SEARXNG_SECRET}" >> .env
    echo "ADMIN_USER=admin" >> .env
    echo "ADMIN_PASSWORD=${ADMIN_PASSWORD}" >> .env
    echo -e "${GREEN}Created .env file with your configurations.${NC}"
else
    echo -e "\n${GREEN}Found existing .env file. Using existing configurations.${NC}"
    DOMAIN=$(grep DOMAIN .env | cut -d '=' -f2)
    ADMIN_PASSWORD=$(grep ADMIN_PASSWORD .env | cut -d '=' -f2)
fi

# 4. Setup SearXNG Configuration
echo -e "\n${BLUE}Setting up SearXNG configurations...${NC}"
mkdir -p searxng
cp searxng-settings.yml searxng/settings.yml
# Ensure SQLite DB file exists so Docker can mount it
touch searchinlet.db

# Use Docker to fix permissions
docker run --rm -v "$(pwd):/data" alpine chmod -R 777 /data/searxng /data/searchinlet.db

# 5. Start the Stack
echo -e "\n${BLUE}Starting all services via Docker Compose...${NC}"
docker compose -f docker-compose.searxng.yml up -d --build

# 6. Verify Deployment
echo -e "\n${BLUE}Verifying stack availability...${NC}"
echo "Waiting for services to boot (15 seconds)..."
sleep 15

# Verify SearXNG internally
if curl -s -G --data-urlencode 'q=test' --data-urlencode 'format=json' http://localhost:8888/search | grep -q "test"; then
    echo -e "${GREEN}Backend services are healthy!${NC}"
else
    echo -e "${RED}Warning: Services might not be ready yet. Check logs with: docker compose -f docker-compose.searxng.yml logs${NC}"
fi

PROTOCOL="https"
if [ "$DOMAIN" == "localhost" ]; then
    PROTOCOL="http"
fi

echo -e "\n${GREEN}🎉 Installation Complete!${NC}"
echo -e "${BLUE}--------------------------------------------------${NC}"
echo -e "  Admin Dashboard: ${GREEN}${PROTOCOL}://${DOMAIN}${NC}"
echo -e "  Username:        ${GREEN}admin${NC}"
echo -e "  Password:        ${GREEN}${ADMIN_PASSWORD}${NC}"
echo -e "${BLUE}--------------------------------------------------${NC}"

echo -e "\n${BLUE}🚀 How to connect your AI Agent (Cursor / Claude):${NC}"
echo -e "\n${GREEN}SSE (Modern/Recommended)${NC}"
echo -e "Use the URL: ${BLUE}${PROTOCOL}://${DOMAIN}/sse${NC}"
echo -e "Add Header:  ${BLUE}Authorization: Bearer sk-YOUR_TOKEN_FROM_DASHBOARD${NC}"
