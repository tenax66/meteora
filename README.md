# meteora

Yet another social network

## generate keys
```bash
$ openssl genpkey -algorithm ED25519 -out key.pem
$ openssl pkey -in key.pem -pubout -out pubkey.pem
```

## run
### client
```bash
cd client
go run .
```

### server
```bash
cd server
go run .
```

#### using docker
```bash
cd server
docker build --tag meteora-server:latest .
docker compose up
```

## messages
Messages are the only object type on the meteora network. One example is given below:

```json
{
    "id":"d490c8bf7b6b96b77abec399ca5dfbbf218e11fc4fd1d28095849004372a481e",
    "content":{
        "created_at": 1699688626,
        "text": "Hello, WebSocket Server!",
    },
    "pubkey": "05a2f1ddf8f59c69da3bbb69f065a6f12f267a436d76235ca914c81e39ffa84b",
    "sig":"f1a7b8bd5b1195d292ab5639c124cc9d7219c338bc46dfaf7f297e1d90f275d344138c3fea7546d36434e8ec7abcfaf30c0e1bf9ac34dd83f3938c8198c2a40f"
}
```
