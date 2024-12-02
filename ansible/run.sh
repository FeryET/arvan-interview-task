#!/usr/bin/env bash

set -euo pipefail
cd "$(dirname "$0")"

eval "$(pyenv init -)"

pyenv activate ansible || exit 1

export ANSIBLE_CONFIG="$PWD/ansible.cfg"
export ANSIBLE_HOST_KEY_CHECKING=False

echo "Running ansible playbooks..."

ansible-playbook -i "./inventory" "./playbooks/main.yml" --diff "${@:1}"
