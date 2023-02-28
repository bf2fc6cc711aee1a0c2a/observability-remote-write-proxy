# **Observability Remote Write Proxy**

### A proxy to accept remote write requests and forward them to Observatorium. Before forwarding the request, the proxy authenticates with the provided oidc credentials and adds the obtained token to the request. The proxy is intended to be used by customers who want to send data to Observatorium but don't want to store Observatorium credentials on their clusters.

## **How to set up the Proxy locally**

#### Flags that are needed to proxy to another prometheus instance:
--proxy.forwardUrl=<Prometheus remote write URL>

#### Flags that are needed for full functionality:
--proxy.forwardUrl=<Prometheus remote write URL> --oidc.enabled --proxy.forwardUrl=<Observatorium Forward URL> --oidc.issuerUrl=<OIDC Issuer URL> --oidc.clientId=<OIDC Client ID> --oidc.clientSecret=<OIDC Client Secret> --oidc.audience=<OIDC Audience> --token.verification.enabled --token.verification.url=<Token Verification URL>

#### Steps:
1. Run the script to start the Prometheus containers (environment.sh)
2. Start the proxy with the flags 
3. Start generating the time series with the [Prometheus-toolbox](https://github.com/pb82/prometheus-toolbox)

## Prerequisites

1. [Golang](https://go.dev/dl/)
2. [Docker](https://docs.docker.com/get-docker/)
3. [Prometheus-toolbox](https://github.com/pb82/prometheus-toolbox)
