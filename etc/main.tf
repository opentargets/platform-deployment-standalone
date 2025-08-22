terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 6.47.0"
    }
  }
  backend "gcs" {}
}

variable "OT_SNAPSHOT_CH" { type = string }
variable "OT_SNAPSHOT_OS" { type = string }
variable "OT_DOMAIN_NAME" { type = string }
variable "OT_SUBDOMAIN_NAME" { type = string }
variable "OT_GCP_PROJECT" { type = string }
variable "OT_GCP_ZONE" { type = string }
variable "OT_GCP_CLOUD_DNS_ZONE" { type = string }
variable "OT_GCP_NETWORK" { type = string }
variable "OT_GCP_SA" { type = string }

data "external" "whoami" {
  program = ["sh", "-c", "echo '{\"username\":\"'$(whoami)'\"}'"]
}

locals {
  user = data.external.whoami.result.username
}

// FIREWALL RULES
resource "google_compute_firewall" "devinstance_allow" {
  name          = "devinstance-allow-${var.OT_SUBDOMAIN_NAME}"
  project       = var.OT_GCP_PROJECT
  network       = var.OT_GCP_NETWORK
  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["devinstance"]
  priority      = 65530
  allow {
    protocol = "tcp"
    ports    = ["80", "443"]
  }
}

// DISKS
resource "google_compute_disk" "clickhouse_data_volume" {
  name     = "devinstance-datavolume-ch-${var.OT_SUBDOMAIN_NAME}"
  project  = var.OT_GCP_PROJECT
  zone     = var.OT_GCP_ZONE
  type     = "pd-balanced"
  snapshot = "projects/${var.OT_GCP_PROJECT}/global/snapshots/${var.OT_SNAPSHOT_CH}"
}

resource "google_compute_disk" "opensearch_data_volume" {
  name     = "devinstance-datavolume-os-${var.OT_SUBDOMAIN_NAME}"
  project  = var.OT_GCP_PROJECT
  zone     = var.OT_GCP_ZONE
  type     = "pd-balanced"
  snapshot = "projects/${var.OT_GCP_PROJECT}/global/snapshots/${var.OT_SNAPSHOT_OS}"
}

// COMPUTE INSTANCES
resource "google_compute_instance" "dev_vm" {
  name         = "devinstance-${var.OT_SUBDOMAIN_NAME}"
  project      = var.OT_GCP_PROJECT
  zone         = var.OT_GCP_ZONE
  machine_type = "n1-standard-4"
  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-12"
      type  = "pd-ssd"
      size  = "10"
    }
  }
  attached_disk {
    source      = google_compute_disk.opensearch_data_volume.id
    mode        = "READ_WRITE"
    device_name = "datavolume-os"
  }
  attached_disk {
    source      = google_compute_disk.clickhouse_data_volume.id
    mode        = "READ_WRITE"
    device_name = "datavolume-ch"
  }
  network_interface {
    network = var.OT_GCP_NETWORK
    access_config {}
  }
  service_account {
    email  = var.OT_GCP_SA
    scopes = ["cloud-platform"]
  }
  labels = {
    "team"        = "opentargets"
    "product"     = "platform"
    "tool"        = "standalone"
    "environment" = "development"
    "created_by"  = "terraform"
    "author"      = local.user
  }
  tags = ["devinstance"]
  metadata = {
    compose-file          = file("compose.yaml"),
    dockerfile-opensearch = file("Dockerfile-opensearch"),
    config                = file("config"),
    nginx-conf = templatefile("nginx.conf.tftpl", {
      OT_DOMAIN_NAME    = var.OT_DOMAIN_NAME
      OT_SUBDOMAIN_NAME = var.OT_SUBDOMAIN_NAME,
    }),
    cleanup = templatefile("cleanup.sh.tftpl", {
      OT_DOMAIN_NAME        = var.OT_DOMAIN_NAME,
      OT_SUBDOMAIN_NAME     = var.OT_SUBDOMAIN_NAME,
      OT_GCP_PROJECT        = var.OT_GCP_PROJECT,
      OT_GCP_ZONE           = var.OT_GCP_ZONE,
      OT_GCP_NETWORK        = var.OT_GCP_NETWORK,
      OT_GCP_CLOUD_DNS_ZONE = var.OT_GCP_CLOUD_DNS_ZONE,
    }),
    config-watcher-script  = file("config-watcher.sh"),
    config-watcher-service = file("config-watcher.service"),
  }
  metadata_startup_script = file("google-startup-script.sh")

  lifecycle {
    ignore_changes = [
      labels["author"]
    ]
  }

  depends_on = [google_compute_disk.clickhouse_data_volume, google_compute_disk.opensearch_data_volume]
}

// DNS RECORD SETS
resource "google_dns_record_set" "main" {
  name         = "${var.OT_SUBDOMAIN_NAME}.${var.OT_DOMAIN_NAME}."
  project      = var.OT_GCP_PROJECT
  managed_zone = var.OT_GCP_CLOUD_DNS_ZONE
  type         = "A"
  ttl          = 300
  rrdatas      = [google_compute_instance.dev_vm.network_interface[0].access_config[0].nat_ip]
}

// OUTPUTS
output "instance_url" {
  value = "https://${var.OT_SUBDOMAIN_NAME}.${var.OT_DOMAIN_NAME}"
}
