FROM quay.io/hypershiftqe/builder:latest as build

RUN yum -y install upx
WORKDIR /go

RUN git clone https://github.com/openshift/hypershift.git && \
    cd hypershift/ && make hypershift && \
    cd bin/ && upx hypershift

FROM alpine as runner
COPY --from=build  /go/hypershift/bin/ /
CMD ["./hypershift"]
