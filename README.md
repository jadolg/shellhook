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
docker run -v $PWD/config.yaml:/config.yaml -p 9081:9081 --name shellhook -d ghcr.io/jadolg/shellhook:ubuntu-0.10.3
```


## Configure
We use a yaml configuration file, and it's read by default from `./config.yaml` (/etc/shellhook/config.yaml if you are using the **.deb** installation)

See <https://github.com/jadolg/shellhook/blob/main/config.yaml> for a full example

## Calling the service

```bash
curl -i -H 'Authorization: KXjk9waX9fqRLQ4t8sQf5IK94e2u1CXr8X4MscDc' https://myserver.example.com/hook?script=5e5adb92-0d04-11ee-97cf-4b6c30e50f6a
```
