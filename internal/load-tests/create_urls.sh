#!/bin/bash
# create_test_links.sh

URLS=(
  "https://en.wikipedia.org/wiki/Special:Random"
  "https://www.bbc.com/news/world"
  "https://www.nationalgeographic.com/science"
  "https://www.imdb.com/chart/top"
  "https://www.reddit.com/r/technology"
  "https://www.github.com/trending"
  "https://www.ted.com/talks"
  "https://www.nasa.gov/multimedia/imageoftheday"
  "https://www.goodreads.com/quotes"
  "https://www.archdaily.com"
)

> read_targets.txt  # Clear the file

for url in "${URLS[@]}"; do
  response=$(curl -s -X POST http://localhost:8080/shorten \
    -H "Content-Type: application/json" \
    -d "{\"url\": \"$url\"}")
  
  short_url=$(echo $response | jq -r '.short_url')
  echo "GET http://localhost:8080/$short_url" >> internal/load-tests/read_targets.txt
  echo "Created: $short_url -> $url"
done

echo "read_targets.txt created!"