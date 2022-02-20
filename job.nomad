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
    value     = "^[^reverse]"
    operator  = "regexp"
  }

  group "resticgui" {
    count = 1
    network {
      mode = "bridge"
      port "ftpc" {
        to = 20
      }
    }

    volume "ceph-resticgui" {
      type      = "csi"
      attachment_mode = "file-system"
      access_mode     = "single-node-writer"
      read_only = false
      source    = "ceph-resticgui"
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
        "traefik.http.routers.to-resticgui-https.middlewares=secHeader@file",
        "traefik.http.routers.to-resticgui-https.middlewares=privateip@file",
        "traefik.http.routers.to-resticgui-https.rule=Host(`resticgui.hanotaux.local`)",
        "traefik.http.routers.to-resticgui-https.tls=true",
      ]
    }

    service {
      port = "8080"
      name = "guacamole"
      connect {
        sidecar_service {
           proxy {

           }
        }
      }
      tags = [
        "traefik.enable=true",
        "traefik.consulcatalog.connect=true",
        "traefik.http.routers.to-filezilla-http.entrypoints=web",
        "traefik.http.routers.to-filezilla-http.rule=Host(`filezilla.hanotaux.local`)",
        "traefik.http.routers.to-filezilla-http.middlewares=force-https@file",
        "traefik.http.routers.to-filezilla-http.middlewares=privateip@file",
        "traefik.http.routers.to-filezilla-https.entrypoints=websecure",
        "traefik.http.routers.to-filezilla-https.middlewares=secHeader@file",
        "traefik.http.routers.to-filezilla-https.middlewares=privateip@file",
        "traefik.http.routers.to-filezilla-https.rule=Host(`filezilla.hanotaux.local`)",
        "traefik.http.routers.to-filezilla-https.tls=true",
      ]
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
      volume_mount {
        volume      = "ceph-resticgui"
        destination = "/srv"
        read_only   = true
      }
    }

    task "vnc" {
      driver = "docker"
      config {
        image = "yggnas/vncserver"
      }
      resources {
        cpu    = 200 # MHz
        memory = 512 # MB
      }
      volume_mount {
        volume      = "ceph-resticgui"
        destination = "/app"
        read_only   = true
      }
  
    }

    task "guac" {
      driver = "docker"
      config {
        image = "maxwaldorf/guacamole"
      }
      resources {
        cpu    = 200 # MHz
        memory = 512 # MB
      }
    }
  }
}
