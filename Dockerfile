FROM golang

ENV SLACK_BOT_TOKEN $SLACK_BOT_TOKEN
ENV RECYCLE_LIST $RECYCLE_LIST

COPY . /build
WORKDIR /build
RUN go build -o /usr/bin/bidonbot
WORKDIR /
CMD bidonbot
