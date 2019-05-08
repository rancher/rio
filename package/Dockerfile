FROM alpine
RUN apk -U --no-cache add ca-certificates git
ENV RIO_IN_CLUSTER=true
COPY rio-controller /usr/bin
CMD ["rio-controller"]
