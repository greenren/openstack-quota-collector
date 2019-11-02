FROM golang:1.13.3-alpine3.10 as build

WORKDIR /src/openstack
ADD . /src/openstack

RUN go build -o /openstack-quota-collector

FROM alpine:3.10

COPY --from=build /openstack-quota-collector /

CMD /openstack-quota-collector