data "containerregistry_image" "ubuntu" {
  name = "ubuntu:latest"
}

output "ubuntu_latest_digest" {
  value = data.containerregistry_image.ubuntu.digest
}
