variable "ci_registry_image" {
  type    = string
  default = "latest"
}

variable "ci_deploy_token" {
  type    = string
  default = "latest"
}

job "resticgui_job" {

  datacenters = ["dc1"]

  type = "service"

  constraint {
    attribute = "${attr.unique.hostname}"
    value     = "^nomad.*"
    operator  = "regexp"
  }

  constraint {
    attribute = "${attr.cpu.arch}"
    value     = "amd64"
  }

  group "resticgui" {
    count = 1
    network {
      mode = "bridge"
    }
    service {
      port = "3030"
      name = "resticgui"
      connect {
        sidecar_service {
           proxy {

           }
        }
      }
      tags = [
        "traefik.enable=true",
        "traefik.consulcatalog.connect=true",
        "traefik.http.routers.to-resticgui-http.entrypoints=web",
        "traefik.http.routers.to-resticgui-http.rule=Host(`resticgui.hanotaux.local`)",
        "traefik.http.routers.to-resticgui-http.middlewares=force-https@file",
        "traefik.http.routers.to-resticgui-http.middlewares=privateip@file",
        "traefik.http.routers.to-resticgui-https.entrypoints=websecure",
        "traefik.http.routers.to-resticgui-https.middlewares=privateip@file",
        "traefik.http.routers.to-resticgui-https.rule=Host(`resticgui.hanotaux.local`)",
        "traefik.http.routers.to-resticgui-https.tls=true",
      ]
    }
    volume "resticgui-db" {
      type      = "csi"
      read_only = false
      attachment_mode = "file-system"
      access_mode     = "single-node-writer"
      source    = "ceph-resticgui-db"
    }
    volume "resticgui-data" {
      type      = "csi"
      read_only = false
      attachment_mode = "file-system"
      access_mode     = "single-node-writer"
      source    = "ceph-resticgui-data"
    }
    volume "resticgui-backup" {
      type      = "csi"
      read_only = false
      attachment_mode = "file-system"
      access_mode     = "single-node-writer"
      source    = "ceph-resticgui-backup"
    }

    task "resticgui" {
      driver = "docker"
      config {
        image = "${var.ci_registry_image}"
        force_pull = "true"
        auth {
          username = "nomad"
          password = "${var.ci_deploy_token}"
        }
      }
      resources {
        cpu    = 200 # MHz
        memory = 256 # MB
      }
      // volume_mount {
      //   volume      = "resticgui-db"
      //   destination = "/opt/test"
      //   read_only   = false
      // }
      // volume_mount {
      //   volume      = "resticgui-data"
      //   destination = "/data"
      //   read_only   = false
      // }
      // volume_mount {
      //   volume      = "resticgui-backup"
      //   destination = "/backup"
      //   read_only   = false
      // }
    }

  }
}
