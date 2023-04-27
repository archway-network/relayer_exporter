# relayer_exporter
Prometheus exporter for go relayer.
Returns metrics which are missing in [gorelayer](https://github.com/cosmos/relayer)

## Metrics
```
# HELP cosmos_relayer_client_expiry Returns light client expiry in unixtime.
# TYPE cosmos_relayer_client_expiry gauge
cosmos_relayer_client_expiry{chain_id="checkersa",path="demo"} 1.68412974e+09
cosmos_relayer_client_expiry{chain_id="checkersb",path="demo"} 1.68412974e+09
# HELP cosmos_relayer_configured_chain Returns configured chain.
# TYPE cosmos_relayer_configured_chain gauge
cosmos_relayer_configured_chain{chain_id="checkersa"} 1
cosmos_relayer_configured_chain{chain_id="checkersb"} 1
# HELP cosmos_relayer_up Was talking to relayer successful.
# TYPE cosmos_relayer_up gauge
cosmos_relayer_up 1
```
