FROM golang:1.18 as build

RUN apt update && apt-get install -y upx-ucl

RUN git clone https://github.com/openshift/hypershift.git && \
    cd hypershift/ && make hypershift && \
    cd bin/ && upx hypershift

FROM alpine as runner
COPY --from=build  /go/hypershift/bin/ /
CMD ["./hypershift"]
