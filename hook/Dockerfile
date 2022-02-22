FROM golang as build
WORKDIR /go/src/hook
COPY . /go/src/hook/
WORKDIR /go/src/hook/cmd/hook
RUN go build

FROM docker:20.10.12-git as runner
COPY --from=build  /go/src/hook/cmd/hook/hook /run/
RUN wget -O /tmp/oc.tar.gz https://mirror.openshift.com/pub/openshift-v4/$(arch)/clients/ocp/stable/openshift-client-linux.tar.gz && \
    cd /usr/local/bin && tar -xvf /tmp/oc.tar.gz && rm -rf /tmp/oc.tar.gz && chmod +x /usr/local/bin/*
WORKDIR run
CMD ["./hook"]