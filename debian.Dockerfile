FROM debian:12
COPY shellhook /usr/bin/shellhook
CMD ["shellhook"]