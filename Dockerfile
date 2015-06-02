FROM golang
ADD . /go/src/github.com/twanies/flow
RUN go install github.com/twanies/flow
CMD ["flow"]
EXPOSE 9999

