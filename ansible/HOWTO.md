# K8s Cluster Ansible

## Setup

To install Ansible, first please install pyenv as described in its [installation guide](https://github.com/pyenv/pyenv?tab=readme-ov-file#installation). After installing pyenv, create a python virtualenv as described below:

```sh
# Create the virutalenv and activate it
pyenv install 3.11
pyenv virtualenv 3.11 ansible
pyenv activate ansible
# Install ansible in its venv
pip install -r requirements.txt
```

After you've set up the ansible environment, run './setup.sh' to setup the ansible dependency chain.

## Execution

In order to execute the ansible playbook, execute "./run.sh". This will run the ansible playbook on your VMs and will install a K8s cluster using K3s.

## Post-Execution

After the execution is completed, a kubeconfig file will be created at path: `~/.kube/k3s.config`. This file has the configuration needed for you to use via kubectl on your host, to control the k3s cluster. You can run kubectl via the command below:

```sh
KUBECONFIG="~/.kube/k3s.config" kubectl get pods --all-namespaces
# or
export KUBECONFIG="~/.kube/k3s.config"
kubectl get pods --all-namespaces
```
