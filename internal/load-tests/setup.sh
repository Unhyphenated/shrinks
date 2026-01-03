#!/bin/bash
# setup.sh - Create test URLs with realistic distribution

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

NUM_URLS=${1:-100}
DISTRIBUTION=${2:-zipfian}  # zipfian or uniform

echo -e "${GREEN}=== Load Test Setup ===${NC}"
echo "Creating ${NUM_URLS} URLs with ${DISTRIBUTION} distribution"
echo ""

# Check if backend is running
if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "Error: Backend not running. Start with: docker-compose up -d"
    exit 1
fi

echo -e "${YELLOW}Creating URLs...${NC}"

# Create URLs and store short codes
> urls.txt
for i in $(seq 1 $NUM_URLS); do
    response=$(curl -s -X POST http://localhost:8080/shorten \
        -H "Content-Type: application/json" \
        -d "{\"url\": \"https://example.com/page${i}\"}")
    
    short_code=$(echo $response | jq -r '.short_url')
    echo "$short_code" >> urls.txt
    
    if [ $((i % 50)) -eq 0 ]; then
        echo "  Created $i/$NUM_URLS..."
    fi
done

echo -e "${GREEN}✓ Created ${NUM_URLS} URLs${NC}"
echo ""

# Generate targets file based on distribution
echo -e "${YELLOW}Generating targets file...${NC}"

if [ "$DISTRIBUTION" = "zipfian" ]; then
    echo "Using Zipfian distribution (80/20 rule)"
    > targets.txt
    
    # Calculate 20% of URLs (these will be "hot")
    HOT_COUNT=$((NUM_URLS / 5))
    
    # Generate 1000 requests with 80% hitting top 20% of URLs
    for i in $(seq 1 800); do
        # Pick random URL from top 20%
        LINE=$((RANDOM % HOT_COUNT + 1))
        URL=$(sed -n "${LINE}p" urls.txt)
        echo "GET http://localhost:8080/$URL" >> targets.txt
    done
    
    for i in $(seq 1 200); do
        # Pick random URL from all URLs
        LINE=$((RANDOM % NUM_URLS + 1))
        URL=$(sed -n "${LINE}p" urls.txt)
        echo "GET http://localhost:8080/$URL" >> targets.txt
    done
    
    echo "  Hot URLs (top 20%): ${HOT_COUNT}"
    echo "  Cold URLs (bottom 80%): $((NUM_URLS - HOT_COUNT))"
    
else
    echo "Using uniform distribution"
    > targets.txt
    while IFS= read -r url; do
        echo "GET http://localhost:8080/$url" >> targets.txt
    done < urls.txt
fi

TOTAL_TARGETS=$(wc -l < targets.txt | tr -d ' ')
echo -e "${GREEN}✓ Generated ${TOTAL_TARGETS} targets${NC}"
echo ""
echo "Ready to test! Run:"
echo "  ./test.sh read 1000 10s"
echo "  ./test.sh write 100 10s"
echo "  ./test.sh mixed 1000 10s"