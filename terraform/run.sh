#!/bin/bash

set -euo pipefail

echo "Checking if the debian image exists..."
DEBIAN_CLOUD_IMG_FILE="/var/tmp/debian.qcow2.org"
DEBIAN_CLOUD_IMG_URL="https://cloud.debian.org/images/cloud/bookworm/20241125-1942/debian-12-generic-amd64-20241125-1942.qcow2"
DEBIAN_CLOUD_IMG_SHA512="8811dc4dee6f9638616d7c3f534d36139eb8716839807b4730259f8f02c16d2f5098d37009ebd547405d9050a4cf29e849cb61eb3425656cda6a12911a10ed24"
if [ ! -f $DEBIAN_CLOUD_IMG_FILE ] || [ "$(sha512sum $DEBIAN_CLOUD_IMG_FILE | awk '{print $1}')" != "$DEBIAN_CLOUD_IMG_SHA512" ]; then
    echo "Downloading debian cloud image to $DEBIAN_CLOUD_IMG_FILE"
    curl -fSL "$DEBIAN_CLOUD_IMG_URL" -o $DEBIAN_CLOUD_IMG_FILE
    cp $DEBIAN_CLOUD_IMG_FILE /var/tmp/debian.qcow2
    qemu-img resize /var/tmp/debian.qcow2 +10G
fi

echo "Running terraform..."
echo "Checking if volume directory exists"
if [ ! -d .volume ]; then
    mkdir .volume
fi

terraform init && terraform apply
