#!/bin/bash

if [ "$1" = "paths" ]
then
    cat <<EOF
 0: archwaytestnet-cosmoshubtestnet -> chns(✔) clnts(✔) conn(✔) (constantine-2<>theta-testnet-001)
 1: archwaytestnet-axelartestnet -> chns(✔) clnts(✔) conn(✔) (constantine-2<>axelar-testnet-lisbon-3)
 2: archwaytestnet-osmosistestnet -> chns(✔) clnts(✔) conn(✔) (constantine-2<>osmo-test-5)
EOF
fi

if [ "$3" = "archwaytestnet-cosmoshubtestnet" ]
then
    cat <<EOF
client 07-tendermint-103 (constantine-2) expires in 4h38m34s (26 Apr 23 11:36 UTC)
client 07-tendermint-1967 (theta-testnet-001) expires in 4h39m20s (26 Apr 23 11:37 UTC)
EOF
fi

if [ "$3" = "archwaytestnet-axelartestnet" ]
then
    cat <<EOF
client 07-tendermint-102 (constantine-2) expires in 3h7m41s (26 Apr 23 10:07 UTC)
client 07-tendermint-401 (axelar-testnet-lisbon-3) expires in 3h10m6s (26 Apr 23 10:09 UTC)
EOF
fi


if [ "$3" = "archwaytestnet-osmosistestnet" ]
then
    >&2 echo Error
    exit 1
fi
