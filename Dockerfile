FROM golang:1.15.2-alpine3.12 as builder

WORKDIR /src

COPY . .

RUN apk add --no-cache make curl git build-base && \
    go version && \
    cd plugin && \
    go build -x -buildmode=plugin -o rappi-middleware.so && \
    pwd && \
    ls -la

FROM devopsfaith/krakend:1.2.0

WORKDIR /etc/krakend

RUN mkdir /etc/krakend/plugin

COPY ./krakend/krakend.json ./krakend.json
COPY --from=builder /src/plugin/rappi-middleware.so /rappi-middleware.so 

ENTRYPOINT [ "/usr/bin/krakend" ]
CMD [ "run", "-c", "/etc/krakend/krakend.json" ]

EXPOSE 8000 8090