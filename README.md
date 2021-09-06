# Minecraft Bot for Discord

This is a small personal project to allow small-scale management of a Minecraft
server via Discord.

## Features

- **Whitelist Management:** Discord members can request to be added to the Minecraft
whitelist. Once a member with the Approvers role confirms, they will be added.
- **Server Restarts:** Discord members can request a server restart. The restart
blocks until the server is empty. A member with the Approvers role can override.

### Screenshots

#### Whitelist Command

![Whitelist Command](https://i.imgur.com/z5AC4Io.png)

#### Whitelist Interaction

![Whitelist Interaction](https://i.imgur.com/nJmqkYs.png)

#### Whitelist Result

![Whitelist Result](https://i.imgur.com/wSY3nB0.png)

#### Restart Command

![Restart Command](https://i.imgur.com/V1uBmXD.png)

#### Restart Interaction

![Restart Interaction](https://i.imgur.com/KBshxTn.png)

#### Restart Result

![Restart Result](https://i.imgur.com/OBY5zU6.png)

## Deployment

The bot can be deployed on Kubernetes.
Simply edit the manifests in [examples/](./examples) and then run:

```sh
kubectl apply --server-side -f examples/
```

It's recommended to build the bot yourself as the upstream image built on [quay.io/kwiesmueller/mc-bot](https://quay.io/repository/kwiesmueller/mc-bot) won't follow any stability guarantees and be updated on demand. Stay safe and make sure you know what you run.

## Contributing

Should you want to try the bot yourself or even contribute, please go ahead.
It won't receive any active support, but Github Issues are a good way to report issues.
This is not intended as a full server management solution running across multiple guilds
and servers. It's just a project from friends for friends.

If you want to do more with this, feel free to get in touch :-)

## Thanks to

- [bwmarrin/discordgo](https://github.com/bwmarrin/discordgo)
- [willroberts/minecraft-client](https://github.com/willroberts/minecraft-client)
- [itzg/docker-minecraft-server](https://github.com/itzg/docker-minecraft-server)
- [pl3xgaming/Purpur](https://github.com/pl3xgaming/Purpur)
