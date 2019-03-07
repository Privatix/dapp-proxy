{
  "log": {
    "loglevel": "warning"
  },
  "inbounds": [
    {
      "port": 1080,
      "listen": "127.0.0.1",
      "tag": "socks-inbound",
      "protocol": "socks",
      "settings": {
        "auth": "noauth",
        "udp": false,
        "ip": "127.0.0.1"
      },
      "sniffing": {
        "enable": true,
        "destOverride": [
          "http",
          "tls"
        ]
      }
    },
    {
      "listen": "127.0.0.1",
      "port": 10081,
      "protocol": "dokodemo-door",
      "settings": {
        "address": "127.0.0.1"
      },
      "tag": "api"
    }
  ],
  "outbounds": [
    {
      "protocol": "vmess",
      "settings": {
        "vnext": [
          {
            "address": "{{if .Address}}",
            "port": 10086,
            "users": []
          }
        ]
      }
    }
  ],
  "routing": {
    "domainStrategy": "IPOnDemand",
    "strategy": "rules",
    "rules": [
      {
        "inboundTag": ["api"],
        "outboundTag": "api",
        "type": "field"
      }
    ]
  },
  "dns": {
    "hosts": {
      "domain:v2ray.com": "www.vicemc.net",
      "domain:github.io": "pages.github.com",
      "domain:wikipedia.org": "www.wikimedia.org",
      "domain:shadowsocks.org": "electronicsrealm.com"
    },
    "servers": [
      "1.1.1.1",
      {
        "address": "114.114.114.114",
        "port": 53,
        "domains": [
          "geosite:cn"
        ]
      },
      "8.8.8.8",
      "localhost"
    ]
  },
  "policy": {
    "levels": {
      "0": {
        "connIdle": 300,
        "downlinkOnly": 30,
        "handshake": 4,
        "uplinkOnly": 5,
        "statsUserDownlink": true,
        "statsUserUplink": true
      }
    },
    "system": {
      "statsInboundUplink": true,
      "statsInboundDownlink": true
    }
  },
  "stats": {},
  "api": {
    "tag": "api",
    "services": [
      "HandlerService",
      "LoggerService",
      "StatsService"
    ]
  }
}