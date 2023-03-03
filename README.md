**Observability Remote Write Proxy**
---

A proxy to accept remote write requests and forward them to Observatorium. Before forwarding the request, the proxy authenticates with the provided oidc credentials and adds the obtained token to the request. The proxy is intended to be used by customers who want to send data to Observatorium but don't want to store Observatorium credentials on their clusters.

### **How to set up the Proxy locally**

Flags:
```
  -oidc.enabled
    	enable oidc authentication
  -oidc.filename string
    	path to oidc configuration file
  -proxy.forwardUrl string
    	url to forward requests to
  -proxy.listen.port int
    	port on which the proxy listens for incoming requests (default 8080)
  -proxy.metrics.port int
    	port on which proxy metrics are exposed (default 9090)
  -token.verification.cacert.enabled
    	If enabled, the CA certificate provided in -token.verification.cacert.filepath will be appended to the trusted CAs store when performing token verification
  -token.verification.cacert.filename string
    	The CA certificate file path. See -token.verification.cacert.enabled for more information. If -token.verification.cacert.enabled it not set or set to false it is not used
  -token.verification.enabled
    	enable data plane token verification
  -token.verification.url string
    	url to validate data plane tokens
```

Flags that are needed to proxy to another prometheus instance:
```
--proxy.forwardUrl=<Prometheus remote write URL>
```

Flags that are needed for full functionality:
```
--proxy.forwardUrl=<Prometheus remote write URL> --oidc.enabled --proxy.forwardUrl=<Observatorium Forward URL> --oidc.issuerUrl=<OIDC Issuer URL> --oidc.clientId=<OIDC Client ID> --oidc.clientSecret=<OIDC Client Secret> --oidc.audience=<OIDC Audience> --token.verification.enabled --token.verification.url=<Token Verification URL>
```

#### Steps:
1. Run the script to start the Prometheus containers (environment.sh)
2. Start the proxy with the flags 
3. Start generating the time series with the [Prometheus-toolbox](https://github.com/pb82/prometheus-toolbox)

#### Proxy image: 
```
https://quay.io/repository/rhoas/observability-remote-write-proxy 
```

## Prerequisites
* [Golang](https://go.dev/dl/)
* [Docker](https://docs.docker.com/get-docker/)
* [Prometheus-toolbox](https://github.com/pb82/prometheus-toolbox)
