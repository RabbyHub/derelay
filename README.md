WalletConnect Relay Server V1
---

Derelay is a high performance Relay server implementation described in the [WalletConnect spec v1](https://docs.walletconnect.com/1.0/specs).

This README file will guide you through the process of compiling and installing the software from the source code.

## Run through Docker

We've provided an out-of-box `docker-compose.yaml` file for you, you could just run

```
docker-compose -f docker-compsoe.yaml up
```

to have your relay server running up.

## Building the source

Building requires both a Go (version 1.19 or later), you can install them using your favourite package manager. Once the dependencies are installed, run

```
go build
```

## Redis

You have to setup a redis server along side the relay server to have it run properly, redis is being used both to gain an unlimited horizontal scalability and to work as a cache for the pending messages.

## Running relay server

As mentioned above, you need to specify the redis server during running the relay server, besides that you can also specify a listening port for the relay server or leave it empty to have it listen on 8080.

```
./derelay -relay.addr :8080 -redis.server_addr 127.0.0.1:6379
```

## Extension

Extended upon the original spec, Derelay has some enhanced features for Dapp, including:

* Awareness of Wallet connection status: This feature allows Dapp to determine whether a wallet is currently connected or not, enabling Dapp to provide appropriate hint to the uesr or take specific actions.
* Wallet ack on received messages: This feature helps Dapp have a better knowledge on whether its message to Wallet has successfully been received, provides a option to Dapp to retry sending message on failure, ensuring a more reliable service to users enduring a poor network condition.
* Application layer ping/pong messages: Derelay introduces application layer ping/pong messages, which allow Dapp to measure the network latency between Dapp and relay server, ensuring a reliable and responsive communication channel.

Derelay introduced two new fields in the messages between Dapp and Wallet, the `role` and `phase`. To utilize the above features, your Dapp just needs set `"role":"dapp"` everytime it sends a message to Wallet, and set `"phase":"sessionRequest"` in the message attempting to connect the Wallet with a QRCode. If the Dapp conforms to this rule, then it will:

1. On session establishing, when a Wallet scanned the QRCode the Dapp shares, the Dapp will immediately receive such a notification:
   ```
   {
    "payload": "",
    "topic": "70a69a10-d3ca-43e8-a418-f6d6e6470969",
    "type": "pub",
    "phase": "sessionReceived",
   ```
2. For every message request the Dapp sends to the Wallet, it will receive an ACK message from the Wallet
   ```
   {
    "payload": "",
    "topic": "70a69a10-d3ca-43e8-a418-f6d6e6470969",
    "type": "ack",
   }
   ```
3. Everytime the Wallet disconnects or resumes connection, the Dapp will get such notification:
   ```
   {
    "payload": "",
    "topic": "70a69a10-d3ca-43e8-a418-f6d6e6470969",
    "type": "pub",
    "phase": "sessionSuspended",
   }
   ```
   or
   ```
   {
    "payload": "",
    "topic": "70a69a10-d3ca-43e8-a418-f6d6e6470969",
    "type": "pub",
    "phase": "sessionResumed",
   }
   ```

## Contributing

We welcome contributions from the community to help improve this project. To contribute, please follow these guidelines:

1. Fork the repository.
2. Create a new branch for your feature or bug fix.
3. Make your changes and commit them with descriptive messages.
4. Push your changes to your forked repository.
5. Submit a pull request, explaining the changes you have made.

## License

This relay server implementation is released under the [BSD 4-Clause license](https://spdx.org/licenses/BSD-4-Clause.html).