#!/bin/bash
set -e

echo "START"

pnpm build

echo "uploading to s3"
aws s3 cp ./dist s3://bot-telega/payments/ --recursive --exclude "*.jpg" --exclude "*.png" --region il-central-1 --profile YOUR_PROFILE_HERE

echo "DONE"