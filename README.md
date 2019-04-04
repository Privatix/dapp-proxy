# Proxy Service Plug-in (based on v2ray)

## Getting started
These instruction will help you build and configure dapp-proxy adapter.

### Building executable

`go build -o <bin>/dapp_proxy <thisrepo>/adapter`

### Configuration
Adapter runs either for client or agent. The decision is made upon examining options in configuration. If it's suitable for agent, the agent mode starts otherwise starts client mode.

#### Agent
```
{
    FileLog         logger configuratoin
        Level           info|error|warning|debug
        StackLevel      info|error|warning|debug
        Prefix          added before each log
        UTC             true|false
        Filename        path to log file
        FileMode        mode to use when creating log file

    V2Ray           proxy configuration
        AlterID         alterId to use for vmess clients
        API             v2ray api endpoint
        InboundTag      tag of vmess inbound
        InboundPort     port of vmess inbound

    Sess            connect to session server configuration
        Endpoint        websocket endpoint
        Origin          
        Product         uuid
        Password        secret

    Monitor         that reports traffic usage to session server
        CountPeriod     period of usage reports in seconds
    }
}
```

Example [agent configuration](/adapter/agent.config.json)

#### Client
```
{
    FileLog         logger configuratoin
        Level           info|error|warning|debug
        StackLevel      info|error|warning|debug
        Prefix          added before each log
        UTC             true|false
        Filename        path to log file
        FileMode        mode to use when creating log file

    V2Ray           proxy configuration
        API             v2ray api endpoint
        InboundTag      tag of socks inbound

    Sess            connect to session server configuration
        Endpoint        websocket endpoint
        Origin          
        Product         uuid
        Password        secret

    Monitor         that reports traffic usage to session server
        CountPeriod     period of usage reports in seconds
    }
}
```

Example [client configuration](/adapter/client.config.json)

## Tests

`go test -v ./...`

## Authors

* [furkhat](https://github.com/furkhat)

## License