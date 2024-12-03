# Terraform VM Provisioning

## Setup the Environment

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

## Provisioning

In order to run the terraform project, use "./terraform/run.sh" entrypoint. This will initialize and apply the terraform configuration.

After the first initialization, you can directly use `terraform apply` to update your VMs configuration.
