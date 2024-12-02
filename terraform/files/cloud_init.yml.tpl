#cloud-config

# Set the timezone and locale to your preferred settings
timezone: "Etc/UTC"
locale: "en_US.UTF-8"

hostname: "${hostname}"

# Create a user with sudo privileges and no password prompt for sudo
users:
  - name: vmuser
    groups: [sudo]
    shell: /bin/bash
    sudo: ["ALL=(ALL) NOPASSWD:ALL"]

chpasswd:
  expire: false
  users:
    - name: vmuser
      password: vmuser
      type: text

# Disable root SSH login and password authentication for SSH
ssh_pwauth: true
disable_root: false

# Configure unattended upgrades for security updates
package_update: false
package_upgrade: false
package_reboot_if_required: false
unattended_upgrades:
  enable: false
