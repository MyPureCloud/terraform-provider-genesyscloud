#!/bin/bash

# Check if environment variables are set
if [ -z "$GENESYSCLOUD_OAUTHCLIENT_ID" ]; then
    echo "Error: GENESYSCLOUD_OAUTHCLIENT_ID is not set"
    exit 1
fi

if [ -z "$GENESYSCLOUD_OAUTHCLIENT_SECRET" ]; then
    echo "Error: GENESYSCLOUD_OAUTHCLIENT_SECRET is not set"
    exit 1
fi

if [ -z "$GENESYSCLOUD_REGION" ]; then
    echo "Error: GENESYSCLOUD_REGION is not set"
    exit 1
fi

echo "Environment variables are set:"
echo "  GENESYSCLOUD_OAUTHCLIENT_ID: ${GENESYSCLOUD_OAUTHCLIENT_ID:0:10}..."
echo "  GENESYSCLOUD_REGION: $GENESYSCLOUD_REGION"
echo ""
