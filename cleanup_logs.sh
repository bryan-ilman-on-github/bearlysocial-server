#!/bin/bash

# Script to delete benchmark logs, CSVs, and PNGs matching specific naming patterns.

# Delete log files.
echo "Deleting log files (bench-*.log)..."
rm -v bench-*.log

# Delete CSV files.
echo "Deleting CSV files (response_times-*.csv)..."
rm -v response_times-*.csv

# Delete PNG files.
echo "Deleting PNG files (response_time_distribution-*.png)..."
rm -v response_time_distribution-*.png

echo "Cleanup complete."
