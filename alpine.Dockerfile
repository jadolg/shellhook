FROM alpine:3.20
COPY shellhook /usr/bin/shellhook
CMD ["shellhook"]