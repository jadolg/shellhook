FROM ubuntu:24.04
COPY shellhook /usr/bin/shellhook
CMD ["shellhook"]