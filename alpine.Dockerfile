FROM alpine:3.21
COPY shellhook /usr/bin/shellhook
CMD ["shellhook"]