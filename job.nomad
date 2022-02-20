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
      port "ftpd" {
        to = 21
      }
      port "data21100" {
        static = 21100
      }
      port "data21101" {
        static = 21101
      }
      port "data21102" {
        static = 21102
      }
      port "data21103" {
        static = 21103
      }
      port "data21104" {
        static = 21104
      }
      port "data21105" {
        static = 21105
      }
      port "data21106" {
        static = 21106
      }
      port "data21107" {
        static = 21107
      }
      port "data21108" {
        static = 21108
      }
      port "data21109" {
        static = 21109
      }
      port "data21110" {
        static = 21110
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
      port = "8080"
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
      port = "3000"
      name = "filezilla"
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

    task "filezilla" {
      driver = "docker"
      config {
        image = "lscr.io/linuxserver/filezilla:latest"
      }
      resources {
        cpu    = 200 # MHz
        memory = 256 # MB
      }
  
    }

    task "vsftpd" {
      driver = "docker"
      config {
        image = "fauria/vsftpd"
        ports = ["ftpd","ftpc","data21100","data21101","data21102","data21103","data21104","data21105","data21106","data21107","data21108","data21109","data21110"]
      }
      resources {
        cpu    = 200 # MHz
        memory = 256 # MB
      }
      env {
        FTP_PASS="password"
        PASV_ADDRESS="${NOMAD_IP_ftpd}"
        PASV_MIN_PORT=21100
        PASV_MAX_PORT=21110
      }
    }
  }
}
