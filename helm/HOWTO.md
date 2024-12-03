# Helm Deployment

## Setup

In order to deploy the helm charts, you will need to install the following tools:

1. helm.
2. helmfile.

## Execution

After the tools are installed, you can execute the deployments using:

```sh
# at ./helm directory invoke:
helmfile -f helmfile.yml apply --skip-deps --debug
```

This will install the prometheus stack, and the postgres cluster needed.
