# Traefik Authentik Forward Plugin

<div align="center">

  <img src="./.assets/icon.svg" width="128" alt="Traefik Authentik Forward Plugin">

</div>

<div align="center">

**Made with ❤️ for the Traefik community**
<br>
⭐ [Star repo](https://github.com/xabinapal/traefik-authentik-forward-plugin/stargazers)
🐛 [Report bug](https://github.com/xabinapal/traefik-authentik-forward-plugin/issues)
💡 [Request feature](https://github.com/xabinapal/traefik-authentik-forward-plugin/issues)

</div>

<div align="center">

**If this plugin helps you, consider supporting me**
<br>
[![Ko-fi](https://img.shields.io/badge/Ko--fi-Support%20me-ff5e5b?logo=ko-fi&logoColor=white)](https://ko-fi.com/xabinapal)

</div>

## What is this?

This plugin provides forward authentication specifically designed for [Authentik](https://goauthentik.io/), an open-source identity provider. It works as a Traefik middleware and integrates directly with Authentik's proxy outposts to provide authentication and authorization for your services.

**Key features:**

- **Built for Authentik**: Works out of the box with any Authentik outpost.
- **Blocks internal paths**: Blocks user access to Authentik's internal auth routes.
- **Improves security**: Prevents external tampering by making sure only Authentik can set authentication data.
- **Flexible behavior**: Supports per-path control over responses (block, redirect to login, or allow access).
- **Authentication cache**: Can cache authentication responses from Authentik to reduce overloading the server with multiple requests.

### How is authentication handled?

Every request received by Traefik is checked against Authentik. If a request is authenticated, the proxy forwards user details to the upstream service by attaching a set of headers. These headers are prefixed with `X-Authentik-`, such as `X-Authentik-Username` or `X-Authentik-Email`.

To prevent header spoofing, any `X-Authentik-*` headers sent by either the upstream or the downstream are removed. Only Authentik is allowed to set these values. These headers are never included in the response to the user.

User sessions are managed using a cookie named like `authentik_proxy_<ID>`. This cookie is set by Authentik during the login flow and is used to identify the session in later requests. The upstream service must not try to modify this cookie. Any changes made to it will be ignored and overwritten by the plugin.

### Why not just use `forwardAuth`?

Traefik's built-in `forwardAuth` middleware is flexible, but that also means more setup and more room for mistakes. This plugin is made specifically for Authentik, so there is no need to write custom routes, tweak headers, or fight with cookie issues.

The official Authentik documentation suggests using `forwardAuth`, but it doesn't mention key caveats related to header spoofing, cookie security, or infrastructure interference. In fact, the example setup shown in the docs can lead to security issues if used as-is.

This plugin handles those problems for you. It removes untrusted headers, protects cookies, and applies consistent logic to all requests. It also adds functionality that would otherwise require multiple middlewares and complex routing logic.

You can even define different auth behaviors for APIs and websites with just a few config lines. It's a faster, safer, and more reliable way to integrate Authentik with Traefik.

### Why does it use the `nginx` endpoint?

Authentik provides multiple outpost endpoints. This plugin uses the `/outpost.goauthentik.io/auth/nginx` one because it gives more control. Unlike the Traefik endpoint, which always returns `302` redirects, this one returns `401` responses, which lets you decide what to do: deny the request, redirect to login or even skip auth entirely for some paths. This puts the decision of how to handle unauthorized requests closer to your app logic.

It also avoids problems caused by proxies, load balancers, or CDNs that often mess with `X-Forwarded-*` HTTP headers required by the Traefik endpoint. Instead, the nginx endpoint uses an additional `X-Original-Uri` header, which stays intact across hops. This makes your auth setup more reliable and predictable.

> ⚠️ **Important**
>
> Authentik still relies on the `X-Forwarded-Host` header, so make sure it isn't modified by proxies in the request chain. That's still much simpler than managing all the headers required by other approaches.

### When should I enable caching?

Caching feature is designed to prevent overloading the Authentik server when handling multiple simultaneous requests for the same session. As an example, consider a website with a protected API: when a browser loads the site, it makes numerous API calls almost simultaneously. Since these requests happen within seconds of each other, it's inefficient to check authentication status with Authentik for every single request.

When enabling caching, it's crucial to set a low `cacheDuration` value, typically 30 seconds or 1 minute at most. This short duration reduces the risk of stale authentication data while still providing the performance benefits. The cache automatically handles security concerns by invalidating itself whenever any request is made to `/outpost.goauthentik.io/*` paths. This means when a user logs out via `/outpost.goauthentik.io/sign_out`, the cache is immediately cleared, preventing access to protected resources with outdated authentication data.

The middleware will add an `X-Authentik-Traefik-Cached` header to upstream requests, containing a boolean value that indicates whether the authentication status and user data were retrieved from a fresh Authentik query or from the cache.

> ⚠️ **Use with caution**
>
> While caching provides performance benefits, it also introduces security considerations that must be carefully evaluated. Use this feature only when you understand the trade-off between performance and security.

## Installation

Add the plugin to your Traefik configuration using the experimental plugins feature:

**File**

```yaml
experimental:
  plugins:
    authentik-forward:
      moduleName: "github.com/xabinapal/traefik-authentik-forward-plugin"
      version: "v1.0.0"
```

**CLI**

```sh
--experimental.plugins.authentik-forward.modulename=github.com/xabinapal/traefik-authentik-forward-plugin
--experimental.plugins.authentik-forward.version=v1.0.0
```

## Configuration

### Authentik settings

- `address`: `string`, **required** \
  Base URL of your Authentik server (e.g., `https://auth.example.com`).

- `cacheDuration`: `string`, optional, default `0s` \
  Caches Authentik responses for the same session to reduce load on Authentik when a user makes multiple requests in a short time.

- `unauthorizedStatusCode`: `uint`, optional, default `401` \
  HTTP status code to return when denying access for request paths matched by `unauthorizedPaths`.

- `redirectStatusCode`: `uint`, optional, default `302` \
  HTTP status code to return when redirecting to login for request paths matched by `redirectPaths`.

- `skippedPaths`: `[]string`, optional, default `["^/.*$"]` \
  List of regex patterns. If the request path matches one of them, the plugin won't ask Authentik for authorization. This list has priority over other both `unauthorizedPaths` and `redirectPaths`.

- `unauthorizedPaths`: `[]string`, optional, default `["^/.*$"]` \
  List of regex patterns. If the request path matches one of them, the plugin denies access using `unauthorizedStatusCode`. This list has priority over `redirectPaths`. Longest match wins.

- `redirectPaths`: `[]string`, optional, default `[]` \
  List of regex patterns. If the request path matches one of them, the plugin redirects to Authentik using `redirectStatusCode`. Longest match wins.

> 📝 **Path matching precedence**
>
> 1. The path is checked against `skippedPaths`. If any regex matches, the request is allowed and Authentik is not checked for authorization. `X-Authentik-*` headers won't be filled in the upstream request.
> 2. Both `unauthorizedPaths` and `redirectPaths` are checked. If no regex matches in either list, the request is allowed, but Authentik is checked, and user info will be sent upstream if authenticated.
> 3. If both lists contain matching regexes, the **longest matching pattern** (by string length) wins. If two matching regexes have the same length, the one from `unauthorizedPaths` takes precedence.

### HTTP Settings

- `timeout`: `string`, optional, default `0s` \
  Connection timeout duration for requests to Authentik (e.g., `"30s"`, `"1m"`). If not specified or equals to `0`, no timeout is applied.

- `tls.ca`: `string`, optional \
   Path to the CA certificate file for verifying the Authentik server certificate.

- `tls.cert`: `string`, optional \
  Path to the client certificate file for mutual TLS authentication to the Authentik server.

- `tls.key`: `string`, optional \
  Path to the client private key file for mutual TLS authentication to the Authentik server.

- `tls.minVersion`: `uint`, optional, default `12` \
  Minimum TLS version to use (`10` for TLS 1.0, `11` for TLS 1.1, `12` for TLS 1.2, `13` for TLS 1.3).

- `tls.maxVersion`: `uint`, optional, default `13` \
  Maximum TLS version to use (`10` for TLS 1.0, `11` for TLS 1.1, `12` for TLS 1.2, `13` for TLS 1.3).

- `tls.insecureSkipVerify`: `bool`, optional, default `false` \
  If set, skip TLS certificate verification, not recommended for production.

## Examples

### File YAML Provider

```yaml
http:
  middlewares:
    my-auth:
      plugin:
        authentik-forward:
          address: https://auth.example.com
          cacheDuration: "1m"

          unauthorizedStatusCode: 401
          redirectStatusCode: 302

          unauthorizedPaths:
            - "^/api/.*$"
            - "^/admin/.*$"
          redirectPaths:
            - "^/app/.*$"
            - "^/$"

          timeout: "30s"
          tls:
            ca: "/etc/ssl/certs/ca.pem"
            cert: "/etc/ssl/certs/client.pem"
            key: "/etc/ssl/private/client.key"
            minVersion: 12
            maxVersion: 13
            insecureSkipVerify: false

  routers:
    api:
      rule: Host(`api.example.com`)
      middlewares:
        - my-auth
      service: api-service
```

### Kubernetes CRD Provider

**Middleware**

```yaml
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: auth-middleware
spec:
  plugin:
    authentik-forward:
      address: https://auth.example.com
      cacheDuration: "1m"

      unauthorizedStatusCode: 401
      redirectStatusCode: 302

      unauthorizedPaths:
        - "^/api/.*$"
        - "^/admin/.*$"
      redirectPaths:
        - "^/app/.*$"
        - "^/$"

      timeout: "30s"
      tls:
        ca: "/etc/ssl/certs/ca.pem"
        cert: "/etc/ssl/certs/client.pem"
        key: "/etc/ssl/private/client.key"
        minVersion: 12
        maxVersion: 13
        insecureSkipVerify: false
```

**IngressRoute**

```yaml
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  name: api-route
spec:
  entryPoints:
    - websecure
  routes:
    - match: Host(`api.example.com`)
      kind: Rule
      middlewares:
        - name: auth-middleware
      services:
        - name: api-service
          port: 80
```

## License

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.
