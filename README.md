# Arvan SRE Project

## Installing Prequisites

### Terraform

To install and configure terraform and libvirt, please run the following commands first:

1. Install the packages:

    ```sh
    sudo apt-get update && \
        sudo apt-get install -y \
                    bridge-utils \
                    qemu-kvm \
                    virtinst \
                    libvirt-daemon \
                    virt-manager \
                    terraform
    ```

2. Add your user to `libvirt` group:`sudo usermod -aG libvirt $USER`

3. Then go to "/etc/libvirt/qemu.conf" and change the following setting:

    ```conf
    # Uncomment and change this line temporarily, please revert this after you are done with this project.
    security_driver = "none"
    ```

4. Restart the libvirt daemon: `sudo systemctl daemon-reload && sudo systemctl restart libvirtd.service`.

5. In order to run the terraform project, use "./terraform/run.sh" entrypoint.

### Ansible

To install Ansible, first please install pyenv as described in its [installation guide](https://github.com/pyenv/pyenv?tab=readme-ov-file#installation). After installing pyenv, create a python virtualenv as described below:

```sh
# Create the virutalenv and activate it
pyenv install 3.11
pyenv virtualenv 3.11 ansible
pyenv activate ansible
# Install ansible in its venv
pip install ansible==11.0.0
```
