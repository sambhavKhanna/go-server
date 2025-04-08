FROM scratch

WORKDIR /usr/src/app

COPY ./go-server .


CMD ["./go-server"]