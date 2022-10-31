FROM quay.io/openshifttest/hypershift-client:builder as build

WORKDIR /go

RUN git clone https://github.com/openshift/hypershift.git && \
    cd hypershift/ && make hypershift

FROM alpine as runner
COPY --from=build  /go/hypershift/bin/ /
CMD ["./hypershift"]
