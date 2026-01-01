#!/bin/bash
# cleanup.sh - Remove all test artifacts

echo "Cleaning up..."

rm -f results_*.bin
rm -f body_*.json
rm -f targets.txt
rm -f urls.txt
rm -f write_targets.txt
rm -f *.html

echo "✓ Cleanup complete!"
echo ""
echo "To regenerate test data:"
echo "  ./setup.sh 100 zipfian"