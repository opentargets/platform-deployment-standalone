terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 6.47.0"
    }
  }
}

variable "OT_SNAPSHOT_CH" { type = string }
variable "OT_SNAPSHOT_OS" { type = string }
variable "OT_DOMAIN_NAME" { type = string }
variable "OT_SUBDOMAIN_NAME" { type = string }
variable "OT_DAYS_TO_LIVE" { type = number }
variable "OT_GCP_SECRET_OPENAI_TOKEN" { type = string }
variable "OT_GCP_PROJECT" { type = string }
variable "OT_GCP_REGION" { type = string }
variable "OT_GCP_ZONE" { type = string }
variable "OT_GCP_CLOUD_DNS_ZONE" { type = string }
variable "OT_GCP_NETWORK" { type = string }
variable "OT_GCP_SA" { type = string }

// FIREWALL RULES
resource "google_compute_firewall" "devinstance_allow" {
  count         = var.OT_GCP_NETWORK == "default" ? 1 : 0
  name          = "devinstance-allow-${var.OT_SUBDOMAIN_NAME}"
  project       = var.OT_GCP_PROJECT
  network       = "default"
  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["devinstance"]
  allow {
    protocol = "tcp"
    ports    = ["80", "8081", "8082"]
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

// COMPUTE INSTANCE
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
    access_config {
      // ephemeral public ip
    }
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
  }
  tags = ["devinstance"]
  metadata = {
    compose-file          = file("compose.yaml"),
    dockerfile-opensearch = file("Dockerfile-opensearch"),
    config                = file("config"),
  }
  metadata_startup_script = <<-EOF
    #!/bin/bash

    # mount data disks
    mkdir -p /platform/clickhouse
    mkdir -p /platform/opensearch
    mount /dev/disk/by-id/google-datavolume-ch /platform/clickhouse
    mount /dev/disk/by-id/google-datavolume-os /platform/opensearch
    chown -R 1000:1000 /platform/clickhouse
    chown -R 1000:1000 /platform/opensearch

    # install docker and at
    apt-get purge -y man-db
    apt-get update
    apt-get install -y ca-certificates curl at
    install -m 0755 -d /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/debian/gpg -o /etc/apt/keyrings/docker.asc
    chmod a+r /etc/apt/keyrings/docker.asc
    echo \
    "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/debian \
    $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
    tee /etc/apt/sources.list.d/docker.list > /dev/null
    apt-get update
    apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
    systemctl start docker

    # copy files from metadata server
    curl -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/attributes/compose-file > /platform/compose.yaml
    curl -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/attributes/dockerfile-opensearch > /platform/Dockerfile-opensearch
    curl -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/attributes/config > /platform/config

    # set auto-delete for disks manually because terraform does not let us
    gcloud config set project ${var.OT_GCP_PROJECT}
    gcloud compute instances set-disk-auto-delete devinstance-${var.OT_SUBDOMAIN_NAME} --auto-delete --disk=devinstance-datavolume-os-${var.OT_SUBDOMAIN_NAME} --zone=${var.OT_GCP_ZONE}
    gcloud compute instances set-disk-auto-delete devinstance-${var.OT_SUBDOMAIN_NAME} --auto-delete --disk=devinstance-datavolume-ch-${var.OT_SUBDOMAIN_NAME} --zone=${var.OT_GCP_ZONE}

    # prepare secrets
    gcloud secrets versions access latest --secret="${var.OT_GCP_SECRET_OPENAI_TOKEN}" > /platform/openai_token
    chmod 600 /platform/openai_token

    # schedule cleanup script
    cat > /platform/cleanup.sh <<-CLEANUP_EOF
      if [ "${var.OT_GCP_NETWORK}" == "default" ]; then
        gcloud compute firewall-rules delete "devinstance-allow-${var.OT_SUBDOMAIN_NAME}" --project="${var.OT_GCP_PROJECT}" --quiet
      fi
      gcloud dns record-sets delete "${var.OT_SUBDOMAIN_NAME}.${var.OT_DOMAIN_NAME}." --type=A --zone="${var.OT_GCP_CLOUD_DNS_ZONE}" --project="${var.OT_GCP_PROJECT}" --quiet
      gcloud compute instances delete "devinstance-${var.OT_SUBDOMAIN_NAME}" --zone="${var.OT_GCP_ZONE}" --project="${var.OT_GCP_PROJECT}" --quiet
    CLEANUP_EOF
    chmod +x /platform/cleanup.sh
    if [ ${var.OT_DAYS_TO_LIVE} -ne 0 ]; then
      echo "/usr/bin/bash /platform/cleanup.sh" | at now + ${var.OT_DAYS_TO_LIVE} days
    fi

    # run platform
    export OT_API_HOSTNAME="${var.OT_SUBDOMAIN_NAME}.${var.OT_DOMAIN_NAME}:8081"
    export OT_API_OPENAI_HOSTNAME="${var.OT_SUBDOMAIN_NAME}.${var.OT_DOMAIN_NAME}:8082"
    export OT_DEPLOYMENT_FOLDER="/platform"
    export OT_WEBAPP_PORT=80
    cd /platform
    docker compose -f compose.yaml up -d
  EOF
}

// DNS RECORD SET
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
  value = "http://${var.OT_SUBDOMAIN_NAME}.${var.OT_DOMAIN_NAME}"
}
