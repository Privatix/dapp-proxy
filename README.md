[![Go report](http://goreportcard.com/badge/github.com/Privatix/dapp-proxy)](https://goreportcard.com/report/github.com/Privatix/dapp-proxy)
[![Maintainability](https://api.codeclimate.com/v1/badges/9a17679c51f051697bb5/maintainability)](https://codeclimate.com/github/Privatix/dapp-proxy/maintainability)

# Proxy Service Plug-in

Proxy [service plug-in](https://github.com/Privatix/privatix/blob/master/doc/service_plug-in.md) 
allows Agents and Clients to buy and sell their internet traffic in form of Proxy service without 3rd party.

    This service plug-in is a PoC service plugin-in for Privatix core.

## Custom integration includes

-   Start and stop of session by Privatix Core

## Benefits from Privatix Core

-   Automatic billing
-   Automatic payment
-   Access control based on billing
-   Automatic credentials delivery
-   Automatic configuration delivery
-   Anytime increase of deposit
-   Privatix GUI for service control

## Service plug-in components:

-   Templates (offering and access)
-   Service adapter (with access to Proxy and Privatix core)
-   Proxy software (with management interface)

# Getting started

These instruction will help you build and configure `dapp-prox`y adapter.

## Building executable

`
./scripts/build.sh
`

## Configuration

Adapter runs either for client or agent. The decision is made upon examining options in configuration. If it's suitable for agent, the agent mode starts otherwise starts client mode.

### Agent

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

Example [agent configuration](/plugin/agent.config.json)

### Client

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

Example [client configuration](/plugin/client.config.json)

## Tests

`go test -v ./...`

# Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct, and the process for submitting pull requests to us.

## Versioning

We use [SemVer](http://semver.org/) for versioning. 
For the versions available, see the [tags on this repository](https://github.com/Privatix/dapp-proxy/tags).


## Authors

* [furkhat](https://github.com/furkhat)

See also the list of [contributors](https://github.com/Privatix/dapp-proxy/contributors)
who participated in this project.

# License

This project is licensed under the **GPL-3.0 License** - see the
[COPYING](COPYING) file for details.
