# relayer_exporter

Prometheus exporter for ibc clients and wallet balance.
Returns metrics about clients expiration date.
Returns wallet balance for configured addresses.

## Configuration

Exporter needs config file in yaml format like following

```yaml
rpc:
  - chainName: archway
    chainId: archway-1
    url: https://rpc.mainnet.archway.io:443
  - chainName: agoric
    chainId: agoric-3
    url: https://main.rpc.agoric.net:443
  - chainName: archwaytestnet
    chainId: constantine-3
    url: https://rpc.constantine.archway.tech:443
    timeout: 2s

github:
  org: archway-network
  repo: networks
  dir: _IBC
  testnetsDir: testnets/_IBC

accounts:
  - address: archway1l2al7y78500h5akvgt8exwnkpmf2zmk8ky9ht3
    chainName: archwaytestnet
    denom: aconst
```

During startup it fetches IBC paths from github based on provided config.
If env var GITHUB_TOKEN is provided it will be used to make authenticated requests to GitHub API.
Using provided RPC endpoints it gets clients expiration dates for fetched paths.
Each RCP endpoint can have a different timeout specified.
If env var GLOBAL_RPC_TIMEOUT (default 5s) is provided, it specifies the timeout for endpoints
without having it defined.

For provided accounts it fetches wallet balances using endpoints defined in rpc list.

## Metrics

```
# HELP cosmos_ibc_client_expiry Returns light client expiry in unixtime.
# TYPE cosmos_ibc_client_expiry gauge
cosmos_ibc_client_expiry{client_id="07-tendermint-0",discord_ids="400514913505640451",dst_chain_id="cosmoshub-4",dst_chain_name="cosmoshub",src_chain_id="archway-1",src_chain_name="archway",status="success"} 1.706270594e+09
cosmos_ibc_client_expiry{client_id="07-tendermint-1152",discord_ids="400514913505640451",dst_chain_id="archway-1",dst_chain_name="archway",src_chain_id="cosmoshub-4",src_chain_name="cosmoshub",status="success"} 1.706270401e+09
# HELP cosmos_ibc_stuck_packets Returns stuck packets for a channel.
# TYPE cosmos_ibc_stuck_packets gauge
cosmos_ibc_stuck_packets{discord_ids="400514913505640451",dst_chain_id="archway-1",dst_chain_name="archway",dst_channel_id="channel-0",src_chain_id="cosmoshub-4",src_chain_name="cosmoshub",src_channel_id="channel-623",status="success"} 0
cosmos_ibc_stuck_packets{discord_ids="400514913505640451",dst_chain_id="cosmoshub-4",dst_chain_name="cosmoshub",dst_channel_id="channel-623",src_chain_id="archway-1",src_chain_name="archway",src_channel_id="channel-0",status="success"} 0
# HELP cosmos_wallet_balance Returns wallet balance for an address on a chain
# TYPE cosmos_wallet_balance gauge
cosmos_wallet_balance{account="archway1l2al7y78500h5akvgt8exwnkpmf2zmk8ky9ht3",chain_id="constantine-3",denom="aconst",status="success"} 4.64e+18
```
