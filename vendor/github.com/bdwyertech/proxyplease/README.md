# proxyplease

Ask nicely, and you might get proxied.

`proxyplease` is a Go package that attempts to establish a valid proxy connection based on available assumptions. It does  by using native and third-party libraries. `proxyplease` returns a DialContext which can be used in an http.Client Transport or other contexts. 

## Examples

You can assume the complete proxy configuration and authentication from system.

```golang
dialContext := proxyplease.NewDialContext(proxyplease.Proxy{})
```

Or maybe you want to specify a specific SOCKS5 proxy:

```golang
u, _ := url.Parse("socks5h://localhost:1080")
dialContext := proxyplease.NewDialContext(proxyplease.Proxy{URL: u})
```

Let's try a HTTP CONNECT Proxy. This proxy requires NTLM authentication. We don't know the user's credentials so we will assume the credentials from the current user session via SSPI (Windows only). 

```golang
u, _ := url.Parse("http://proxy.example.com:8888")
dialContext := proxyplease.NewDialContext(proxyplease.Proxy{URL: u})
```

Or maybe you want to use specific credentials for NTLM authentication. Oh, but you forgot proxy URL. If the system was configured with a proxy URL, you'll still get proxied:

```golang
dialContext := proxyplease.NewDialContext(proxyplease.Proxy{Username: "foo", Password: "bar", Domain: "EXAMPLE"})
```

Oh no! This user's proxy only supported Basic authentication. Don't worry. The above example still covers you if those credentials are valid. 

But what if you need a specific user-agent. Easy!

```golang
h := &http.Header{}
h.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:68.0) Gecko/20100101 Firefox/68.0")
dialContext := proxyplease.NewDialContext(proxyplease.Proxy{Headers: h})
```

What if the proxy depends on the target URL and you need to look up via PAC?

```golang
t, _ := url.Parse("https://www.google.com")
dialContext := proxyplease.NewDialContext(proxyplease.Proxy{TargetURL: t})
```

## Proxy Support

### SOCKS

| Protocol | URI          | No Auth | User / Pass | GSSAPI | DNS |
| -------- | ------------ | ------- | ----------- | ------ | --- |
| SOCKS4   | `socks4://`  | ✔️     | ❌          | ❌    | ❌ |
| SOCKS4a  | `socks4a://` | ✔️     | ❌          | ❌    | ✔️ |
| SOCKS5   | `socks5://`  | ✔️     | ✔️          | ❌    | ✔️ |
| SOCKS5h  | `socks5h://` | ✔️     | ✔️          | ❌    | ✔️ |

The `golang.org/x/net/proxy` will always do remote DNS for `socks5://`.

### HTTP CONNECT

| Protocol | URI        | No Auth | Basic | NTLM | Negotiate::Kerberos | Negotiate::NTLM | Kerberos | Digest |
| -------- | ---------  | ------- | ----- | ---- | ------------------- | --------------- | -------- | ------ |
| HTTP     | `http://`  | ✔️     | ✔️   | ✔️  | ✔️                 | ❌             | ❌      | ❌    |
| HTTPS    | `https://` | ✔️     | ✔️   | ✔️  | ✔️                 | ❌             | ❌      | ❌    |

If the proxy iniitally responds with a `407 Proxy Authentication Required`, the `Proxy-Authenticate` header(s) will be inspected for authentication schemes supported by the server. Each authentication scheme will be attempted in order of response until a `200 Connection Established`. If no credentials are supplied, `proxyplease` will attempt to transparently assume the current user's credentials from SSPI (SSPI is supported on Windows only and used for NTLM, Kerberos and Negotiate authentications schemes) or if they are hardcoded in environemt variables. Ex: `HTTP_PROXY=http://foo:bar@example.com:3128`. If `proxyplease` does not have enough information to attempt the authentication, the attempt will fail and another scheme will be attempted.

## Proxy Selection

The proxy URL can be specified by passing a URL type. Example:

```golang
u, _ := url.Parse("socks5://localhost:8888")
dialContext := proxyplease.NewDialContext(proxyplease.Proxy{URL: u})
```

If a proxy URL is not provided, `proxyplease` will attempt to infer the URL from the system utilizing [go-get-proxied](https://github.com/rapid7/go-get-proxied). If a proxy cannot be determined, it will be assumed the connection is direct.

The proxy will be selected by the following priority:

**Windows**
   1. `proxyplease.Proxy.URL`
   1. Environment Variable: `HTTPS_PROXY`, `HTTP_PROXY`, `FTP_PROXY`, or `ALL_PROXY`. `NO_PROXY` is respected.
   1. Internet Options: Automatically detect settings (`WPAD`)
   1. Internet Options: Use automatic configuration script (`PAC`)
   1. Internet Options: Manual proxy server
   1. WINHTTP: (`netsh winhttp`)

**Linux**
   1. `proxyplease.Proxy.URL`
   1.  Environment Variable: `HTTPS_PROXY`, `HTTP_PROXY`, `FTP_PROXY`, or `ALL_PROXY`. `NO_PROXY` is respected.

**MacOS**
   1. `proxyplease.Proxy.URL`
   1. Environment Variable: `HTTPS_PROXY`, `HTTP_PROXY`, `FTP_PROXY`, or `ALL_PROXY`. `NO_PROXY` is respected.
   1. Network Settings: `scutil`

## Known Issues

- The Negotiate authentication sequence is supposed to fallback to Negotiate::NTLM if Negotiate::Kerberbos fails. This is currently unsupported.
- Digest authentication is currently unsupported
- Pure Kerberos authentication is currently unsupported. (In most environments, Kerberos authentication is usually wrapped as Negotiate::Kerberos, which is supported)
- Negotiate::Kerberos is currently only supported on Windows
- No tests
- No keyring support (example: Windows Credential Manager might have stored credentials to a SOCKS proxy)

## References

The code for this project was heavily influenced by the following authors. Many thanks to them.

- https://github.com/rapid7/go-get-proxied
- https://github.com/Codehardt/go-ntlm-proxy-auth
- https://github.com/dpotapov/go-spnego
- https://github.com/G-Research/go-ntlm-auth
- https://github.com/alexbrainman/sspi
- https://github.com/git-lfs/git-lfs/tree/master/lfsapi
- https://github.com/ThomsonReutersEikon/go-ntlm
