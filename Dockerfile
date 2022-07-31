FROM golang:1.17.10

ENV APP_HOME /go/src/githubapp
RUN mkdir -p "$APP_HOME"
WORKDIR "$APP_HOME"

#RUN go install github.com/google/go-github/github@master
#RUN go install golang.org/x/oauth2@latest
#RUN go install github.com/go-kit/kit/endpoint@latest
#RUN go install github.com/go-kit/kit/transport/http@latest
#RUN go install github.com/gorilla/mux@latest
#RUN go install github.com/pkg/errors@latest

#ENV GO111MODULE=auto
#ENV GOFLAGS=-mod=vendor

#COPY go.mod ./
#COPY go.sum ./
#RUN go mod download

#COPY *.go ./

ADD Application ./
WORKDIR ./Application/cmd

RUN go build -o /githubapp
#EXPOSE 8010
CMD ["/githubapp"]

# next steps
# docker build --tag docker-github .
# docker image ls
# docker run -d --publish 10000:10000 docker-github
# curl http://localhost:10000/