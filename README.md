# relayer_exporter
Prometheus exporter for go relayer.
Returns metrics about client expiry which are missing in [gorelayer](https://github.com/cosmos/relayer)

## Metrics
```
# HELP cosmos_relayer_client_expiry Returns light client expiry in unixtime.
# TYPE cosmos_relayer_client_expiry gauge
cosmos_relayer_client_expiry{chain_id="axelar-testnet-lisbon-3",path="archwaytestnet-axelartestnet"} 1.6817361e+09
cosmos_relayer_client_expiry{chain_id="constantine-2",path="archwaytestnet-axelartestnet"} 1.68173604e+09
cosmos_relayer_client_expiry{chain_id="constantine-2",path="archwaytestnet-cosmoshubtestnet"} 1.68173634e+09
cosmos_relayer_client_expiry{chain_id="constantine-2",path="archwaytestnet-osmosistestnet"} 1.68173166e+09
cosmos_relayer_client_expiry{chain_id="osmo-test-4",path="archwaytestnet-osmosistestnet"} 1.68173166e+09
cosmos_relayer_client_expiry{chain_id="theta-testnet-001",path="archwaytestnet-cosmoshubtestnet"} 1.68173634e+09
# HELP cosmos_relayer_up Was talking to relayer successful.
# TYPE cosmos_relayer_up gauge
cosmos_relayer_up 1
```
