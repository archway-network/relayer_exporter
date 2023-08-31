# relayer_exporter
Prometheus exporter for ibc clients.
Returns metrics about clients expiration date.

## Configuration
Exporter needs config file in yaml format like following

```yaml
rpc:
  - chainId: archway-1
    url: https://rpc.mainnet.archway.io:443
  - chainId: agoric-3
    url: https://main.rpc.agoric.net:443

github:
  org: archway-network
  repo: networks
  dir: _IBC
```

During startup it fetches IBC paths from github based on provided config.
Using provided RPC endpoints it gets clients expiration dates for fetched paths.

## Metrics
```
# HELP cosmos_ibc_client_expiry Returns light client expiry in unixtime.
# TYPE cosmos_ibc_client_expiry gauge
cosmos_ibc_client_expiry{chain_id="agoric-3",client_id="07-tendermint-75",path="archway-1<->agoric-3"} 1.694678473e+09
cosmos_ibc_client_expiry{chain_id="archway-1",client_id="07-tendermint-23",path="archway-1<->agoric-3"} 1.694678593e+09
```
