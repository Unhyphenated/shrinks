# Load Testing

Simple load testing suite for the URL shortener.

## Quick Start

# 1. Create test data (100 URLs with realistic distribution)
./setup.sh 100 zipfian

# 2. Run read test (compares cold vs warm cache)
./test.sh read 1000 10s

# 3. Run write test
./test.sh write 100 10s

# 4. Run mixed workload (90% reads, 10% writes)
./test.sh mixed 1000 10s

# 5. Clean up
./cleanup.sh## Scripts

- `setup.sh <num_urls> <distribution>` - Create test data
  - Distribution: `zipfian` (realistic) or `uniform`
- `test.sh <mode> <rate> <duration>` - Run tests
  - Modes: `read`, `write`, `mixed`
- `cleanup.sh` - Remove artifacts

## Distribution Types

**Zipfian (Recommended):** 80% of requests hit 20% of URLs (realistic)
**Uniform:** All URLs equally likely (stress test)

## Results

See `performance_results.md` for baseline metrics.



