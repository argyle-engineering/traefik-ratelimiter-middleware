This repository includes Traefik middleware plugin for calling a ratelimiter service.

[![Build Status](https://github.com/argyle-engineering/traefik-ratelimiter-middleware/actions/workflows/main.yml/badge.svg?branch=master)](https://github.com/argyle-engineering/traefik-ratelimiter-middleware/actions)

The existing plugins can be browsed into the [Plugin Catalog](https://plugins.traefik.io).

# Developing a Traefik plugin

[Traefik](https://traefik.io) plugins are developed using the [Go language](https://golang.org).

A [Traefik](https://traefik.io) middleware plugin is just a [Go package](https://golang.org/ref/spec#Packages) that provides an `http.Handler` to perform specific processing of requests and responses.

Rather than being pre-compiled and linked, however, plugins are executed on the fly by [Yaegi](https://github.com/traefik/yaegi), an embedded Go interpreter.

### Configuration

For each plugin, the Traefik static configuration must define the module name (as is usual for Go packages).

The following declaration (given here in YAML) defines a plugin:

```yaml
# Static configuration

experimental:
  plugins:
    example:
      moduleName: github.com/argyle-engineering/traefik-ratelimiter-middleware
      version: v0.0.1
```

Here is an example of a file provider dynamic configuration (given here in YAML), where the interesting part is the `http.middlewares` section:

```yaml
# Dynamic configuration

http:
  routers:
    my-router:
      rule: host(`demo.localhost`)
      service: service-foo
      entryPoints:
        - web
      middlewares:
        - my-plugin

  services:
   service-foo:
      loadBalancer:
        servers:
          - url: http://127.0.0.1:5000

  middlewares:
    my-plugin:
      plugin:
        example:
          url: http://ratelimiter.default.svc.cluster.local.:8000
          dryRun: true
```

## Logs

Currently, the only way to send logs to Traefik is to use `os.Stdout.WriteString("...")` or `os.Stderr.WriteString("...")`.

In the future, we will try to provide something better and based on levels.
