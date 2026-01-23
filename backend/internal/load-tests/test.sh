#!/bin/bash
# test.sh - Unified load testing script

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

MODE=${1:-read}  # read, write, or mixed
RATE=${2:-1000}
DURATION=${3:-10s}

echo -e "${GREEN}=== Load Test: $(echo $MODE | tr '[:lower:]' '[:upper:]') ===${NC}"
echo "Rate: ${RATE} req/s"
echo "Duration: ${DURATION}"
echo ""

# Check services
if ! docker ps | grep -q shrinks_backend; then
    echo -e "${RED}Error: Backend not running${NC}"
    exit 1
fi

case $MODE in
    read)
        if [ ! -f "targets.txt" ]; then
            echo -e "${RED}Error: targets.txt not found. Run ./setup.sh first${NC}"
            exit 1
        fi
        
        echo -e "${YELLOW}Test 1: Cold (no cache)${NC}"
        docker exec shrinks_redis redis-cli FLUSHALL > /dev/null
        docker exec shrinks_redis redis-cli CONFIG RESETSTAT > /dev/null
        vegeta attack -targets=targets.txt -rate=${RATE} -duration=${DURATION} -redirects=-1 -output=results_cold.bin
        
        echo ""
        echo -e "${YELLOW}Test 2: Warm (with cache)${NC}"
        docker exec shrinks_redis redis-cli CONFIG RESETSTAT > /dev/null
        vegeta attack -targets=targets.txt -rate=${RATE} -duration=${DURATION} -redirects=-1 -output=results_warm.bin
        
        echo ""
        echo -e "${GREEN}=== COLD (No Cache) ===${NC}"
        vegeta report results_cold.bin
        
        echo ""
        echo -e "${GREEN}=== WARM (With Cache) ===${NC}"
        vegeta report results_warm.bin
        
        HITS=$(docker exec shrinks_redis redis-cli INFO stats | grep "keyspace_hits:" | cut -d: -f2 | tr -d '\r')
        MISSES=$(docker exec shrinks_redis redis-cli INFO stats | grep "keyspace_misses:" | cut -d: -f2 | tr -d '\r')
        echo ""
        echo "Cache: ${HITS} hits, ${MISSES} misses"
        ;;
        
    write)
        NUM_WRITES=$((RATE * ${DURATION%s}))
        echo "Generating ${NUM_WRITES} unique URLs..."
        
        > write_targets.txt
        for i in $(seq 1 $NUM_WRITES); do
            TS=$(date +%s%N)
            echo "{\"url\": \"https://example.com/write-${TS}-${i}\"}" > "body_${i}.json"
            echo "POST http://localhost:8080/shorten" >> write_targets.txt
            echo "Content-Type: application/json" >> write_targets.txt
            echo "@body_${i}.json" >> write_targets.txt
            echo "" >> write_targets.txt
            
            if [ $((i % 100)) -eq 0 ]; then
                echo "  Generated $i/$NUM_WRITES..."
            fi
        done
        
        echo ""
        echo -e "${YELLOW}Running write test...${NC}"
        vegeta attack -targets=write_targets.txt -rate=${RATE} -duration=${DURATION} -redirects=-1 -output=results_write.bin
        
        echo ""
        vegeta report results_write.bin
        
        rm -f body_*.json write_targets.txt
        ;;
        
    mixed)
        if [ ! -f "targets.txt" ]; then
            echo -e "${RED}Error: targets.txt not found. Run ./setup.sh first${NC}"
            exit 1
        fi
        
        READ_RATE=$((RATE * 9 / 10))
        WRITE_RATE=$((RATE / 10))
        
        echo "Read rate: ${READ_RATE} req/s (90%)"
        echo "Write rate: ${WRITE_RATE} req/s (10%)"
        echo ""
        
        # Generate write targets
        NUM_WRITES=$((WRITE_RATE * ${DURATION%s}))
        > write_targets.txt
        for i in $(seq 1 $NUM_WRITES); do
            TS=$(date +%s%N)
            echo "{\"url\": \"https://example.com/mixed-${TS}-${i}\"}" > "body_${i}.json"
            echo "POST http://localhost:8080/shorten" >> write_targets.txt
            echo "Content-Type: application/json" >> write_targets.txt
            echo "@body_${i}.json" >> write_targets.txt
            echo "" >> write_targets.txt
        done
        
        echo -e "${YELLOW}Running mixed workload...${NC}"
        docker exec shrinks_redis redis-cli CONFIG RESETSTAT > /dev/null
        
        vegeta attack -targets=targets.txt -rate=${READ_RATE} -duration=${DURATION} -redirects=-1 -output=results_reads.bin &
        vegeta attack -targets=write_targets.txt -rate=${WRITE_RATE} -duration=${DURATION} -redirects=-1 -output=results_writes.bin &
        wait
        
        echo ""
        echo -e "${GREEN}=== READS ===${NC}"
        vegeta report results_reads.bin
        
        echo ""
        echo -e "${GREEN}=== WRITES ===${NC}"
        vegeta report results_writes.bin
        
        HITS=$(docker exec shrinks_redis redis-cli INFO stats | grep "keyspace_hits:" | cut -d: -f2 | tr -d '\r')
        MISSES=$(docker exec shrinks_redis redis-cli INFO stats | grep "keyspace_misses:" | cut -d: -f2 | tr -d '\r')
        echo ""
        echo "Cache: ${HITS} hits, ${MISSES} misses"
        
        rm -f body_*.json write_targets.txt
        ;;
        
    *)
        echo "Usage: ./test.sh <mode> <rate> <duration>"
        echo "Modes: read, write, mixed"
        echo "Example: ./test.sh read 1000 10s"
        exit 1
        ;;
esac

echo ""
echo -e "${GREEN}✓ Test complete!${NC}"