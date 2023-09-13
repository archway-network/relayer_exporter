# relayer_exporter
Prometheus exporter for ibc clients and wallet balance.
Returns metrics about clients expiration date.
Returns wallet balance for configured addresses.

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

accounts:
  - address: archway1l2al7y78500h5akvgt8exwnkpmf2zmk8ky9ht3
    chainId: constantine-3
    denom: aconst
```

During startup it fetches IBC paths from github based on provided config.
If env var GITHUB_TOKEN is provided it will be used to make authenticated requests to GitHub API.
Using provided RPC endpoints it gets clients expiration dates for fetched paths.

For provided accounts it fetches wallet balances using endpoints defined in rpc list.

## Metrics
```
# HELP cosmos_ibc_client_expiry Returns light client expiry in unixtime.
# TYPE cosmos_ibc_client_expiry gauge
cosmos_ibc_client_expiry{client_id="07-tendermint-23",host_chain_id="archway-1",target_chain_id="agoric-3"} 1.695283384e+09
cosmos_ibc_client_expiry{client_id="07-tendermint-75",host_chain_id="agoric-3",target_chain_id="archway-1"} 1.69528327e+09
# HELP cosmos_wallet_balance Returns wallet balance for an address on a chain
# TYPE cosmos_wallet_balance gauge
cosmos_wallet_balance{account="archway1l2al7y78500h5akvgt8exwnkpmf2zmk8ky9ht3",chain_id="constantine-3",denom="aconst",status="success"} 4.64e+18
```
