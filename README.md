# censorship-proxy

Censorship-proxy is a proxy for [JSON-RPC API](https://ethereum.org/developers/docs/apis/json-rpc) that censors transactions issued by blacklisted addresses. This proxy was developed as part of a proof-of-concept for a censorship attack on [Optimism](https://www.optimism.io/) network.

## Installation

Clone the repository and build the proxy:

```bash
go build .
```

## Usage

Run the proxy with the following command:

```bash
./censorship-proxy <config address> <client address> <target address>
```

Note that addresses should be in the following format: `<address>:<port>`. For example:

```bash
127.0.0.1:1234
```

The configuration can be updated in real-time by sending a `POST` request to the proxy with the following JSON payload:

```json
{
  "censoredAddresses": ["0xabcd...", "0x1234...", ...]
}
```

The easiest way to send a `POST` request to the proxy is by using `curl`:

```bash
curl -X POST -H "Content-Type: application/json" -d '{"censoredAddresses": ["0xabcd...", "0x1234..."]}' http://<config address>
```
