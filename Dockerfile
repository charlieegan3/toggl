FROM golang:1.10 as build

WORKDIR /go/src/github.com/charlieegan3/toggl

COPY . .

RUN CGO_ENABLED=0 go build -o toggl main.go


FROM scratch
ADD ca-certificates.crt /etc/ssl/certs/
COPY --from=build /go/src/github.com/charlieegan3/toggl/toggl /
CMD ["/toggl"]
