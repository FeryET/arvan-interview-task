terraform {
  required_providers {
    libvirt = {
      source  = "dmacvicar/libvirt"
      version = "0.8.1"
    }
  }
}

locals {
  vm_names = ["vm1", "vm2", "vm3"]
  # Default libvirt network gateway
  gateway_address = "192.168.122.1"
  ip_address = {
    "vm1" : "192.168.122.100",
    "vm2" : "192.168.122.101",
    "vm3" : "192.168.122.102"
  }
  memory = {
    "vm1" : "4000",
    "vm2" : "3000",
    "vm3" : "3000"
  }

  cpus = {
    "vm1": 2,
    "vm2": 2,
    "vm3": 2,
  }
}

# Instance the provider
provider "libvirt" {
  uri = "qemu:///system"
}

resource "libvirt_pool" "debian" {
  name = "debian"
  type = "dir"
  target {
    path = abspath("${path.module}/.volume")
  }
}

resource "libvirt_volume" "debian-qcow2" {
  name   = "debian-base.qcow2"
  pool   = libvirt_pool.debian.name
  source = "file:///var/tmp/debian.qcow2"
  format = "qcow2"
}


# Create cloud-init disk for each VM
resource "libvirt_cloudinit_disk" "init" {
  for_each       = toset(local.vm_names)
  name           = "${each.key}_init.iso"
  user_data      = templatefile("${path.module}/files/cloud_init.yml.tpl", { "hostname" : each.key })
  network_config = templatefile("${path.module}/files/network_config.yml.tpl", { "ip_address" : local.ip_address[each.key], "gateway_address" : local.gateway_address })
  pool           = libvirt_pool.debian.name
}

# Create individual Debian volumes based on the base image
resource "libvirt_volume" "debian" {
  for_each = toset(local.vm_names)

  name           = "${each.key}.qcow2"
  base_volume_id = libvirt_volume.debian-qcow2.id
  pool           = libvirt_pool.debian.name
}

resource "libvirt_domain" "domain-debian" {
  for_each = toset(local.vm_names)

  name   = each.key
  memory = local.memory[each.key]
  vcpu   = local.cpus[each.key]

  cloudinit = libvirt_cloudinit_disk.init[each.key].id

  network_interface {
    network_name = "default"
    addresses    = [local.ip_address[each.key]]
  }

  console {
    type        = "pty"
    target_port = "0"
    target_type = "serial"
  }

  console {
    type        = "pty"
    target_type = "virtio"
    target_port = "1"
  }

  disk {
    volume_id = libvirt_volume.debian[each.key].id
  }
}
