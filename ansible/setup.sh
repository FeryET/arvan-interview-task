#!/bin/bash

set -euo pipefail
cd "$(dirname "$0")"
eval "$(pyenv init -)"

pyenv activate ansible || exit 1

echo "Installing ansible role..."
ansible-galaxy role install -r "./ansible-requirements.yml"

echo "Downloading ansible collections"
ansible-galaxy collection download -r "./ansible-requirements.yml"

echo "Installing ansible collections"
cd "./collections" || exit 1
ansible-galaxy collection install -r "./requirements.yml"
