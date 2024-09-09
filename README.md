# shellhook

Run scripts remotely with a simple HTTP call

## Install

Install from my private repository

```bash
wget -O - https://deb.akiel.dev/gpg.pub.key | gpg --dearmor | sudo tee /etc/apt/trusted.gpg.d/akiel.gpg
sudo apt-add-repository deb "https://deb.akiel.dev/ all main"
sudo apt install shellhook
```

or

Download the latest version `.deb` package from GitHub releases and install using `dpkg`

or

Download the binary from releases and make your own setup

## Docker

You can also use shellhook from Docker. We ship images based in alpine, debian, and ubuntu.
See https://github.com/jadolg/shellhook/pkgs/container/shellhook

```bash
docker run -v $PWD/config.yaml:/config.yaml -p 9081:9081 --name shellhook -d ghcr.io/jadolg/shellhook:alpine-0.10.3
```


## Configure
We use a yaml configuration file and it's read by default from `./config.yaml` (/etc/shellhook/config.yaml if you are using the **.deb** installation)

```yaml
default_token: KXjk9waX9fqRLQ4t8sQf5IK94e2u1CXr8X4MscDc # Token used for all scripts that don't specify one

environment:
  - key: TITLE
    value: Mr.

scripts:
  - id: 5e5adb92-0d04-11ee-97cf-4b6c30e50f6a # ID of the script (a UUID)
    path: ./scripts/success.sh # Path to the script
    user: akiel # If specified, the script is run using this user
  - id: c7c664c0-0d0e-11ee-a3c9-17023c4d78f3
    path: ./scripts/failure.sh
    token: YT9U08gqQ8yxa0Sk3PnDk6jpWu31bCyqa5SRQVFV8 # If specified, this token is used for authorization instead of the default one
    concurrent: true # Set this to true if your script can run concurrently (default: false)
  - id: 47878e38-a700-11ee-bc6d-f3d25921fcde
    inline: | # Use an inline script instead of a path to a script
      echo "Hello, world!"
  - id: 34ca006a-ece6-11ee-a395-17c174ecf4c7
    inline: |
      echo "Hello, $TITLE $NAME!"
    environment: # Environment variables to pass to the script
      - key: NAME
        value: Frodo
```

## Calling the service

```bash
curl -i -H 'Authorization: KXjk9waX9fqRLQ4t8sQf5IK94e2u1CXr8X4MscDc' https://myserver.example.com/hook?script=5e5adb92-0d04-11ee-97cf-4b6c30e50f6a
```
