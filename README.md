# meteora

Yet another social network

## generate keys
```bash
$ openssl genpkey -algorithm ED25519 -out key.pem
$ openssl pkey -in key.pem -pubout -out pubkey.pem
```
