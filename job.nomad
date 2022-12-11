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
        "traefik.http.routers.to-resticgui-http.rule=Host(`resticgui.mhx.lan`)",
        "traefik.http.routers.to-resticgui-http.middlewares=force-https@file",
        "traefik.http.routers.to-resticgui-https.entrypoints=websecure",
        "traefik.http.routers.to-resticgui-https.rule=Host(`resticgui.mhx.lan`)",
        "traefik.http.routers.to-resticgui-https.middlewares=sec-header@file",
        "traefik.http.routers.to-resticgui-https.middlewares=private-ip@file",
        "traefik.http.routers.to-resticgui-https.tls=true",
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
    }

  }
}
