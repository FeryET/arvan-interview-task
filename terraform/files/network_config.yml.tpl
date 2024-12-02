version: 2
ethernets:
  ens3:
    addresses:
      - ${ip_address}/24
    gateway4: ${gateway_address}
    nameservers:
      addresses:
        - 8.8.8.8
        - 8.8.4.4
