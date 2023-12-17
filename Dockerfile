############ BUILD ####################
FROM golang:1.19 as builder

ENV GO111MODULE=on

# Copy service
RUN mkdir -p $GOPATH/src/dev.hackerman.me/artheon/artheon-rpc
COPY main.go go.mod $GOPATH/src/dev.hackerman.me/artheon/artheon-rpc/
COPY models $GOPATH/src/dev.hackerman.me/artheon/artheon-rpc/models/
COPY web $GOPATH/src/dev.hackerman.me/artheon/artheon-rpc/web/
COPY public $GOPATH/src/dev.hackerman.me/artheon/artheon-rpc/public/

WORKDIR $GOPATH/src/dev.hackerman.me/artheon/artheon-rpc
RUN pwd && ls -lah

# Authorize SSH Host
RUN mkdir -p /root/.ssh && \
    chmod 0700 /root/.ssh && \
    ssh-keyscan gitlab.com > /root/.ssh/known_hosts

# Add SSH configuration for the root user for gitlab.com
RUN echo "Host gitlab.com\
        HostName gitlab.com\
        User root\
        IdentityFile ~/.ssh/id_rsa\
        ForwardAgent yes" > ~/.ssh/config

ENV PUBLIC_KEY="ssh-rsa == builder@veverse.com"
ENV PRIVATE_KEY="-----BEGIN OPENSSH PRIVATE KEY-----\n\
-----END OPENSSH PRIVATE KEY-----"

# Add the keys and set permissions
RUN echo $PRIVATE_KEY > /root/.ssh/id_rsa && \
    echo $PUBLIC_KEY > /root/.ssh/id_rsa.pub && \
    chmod 600 /root/.ssh/id_rsa && \
    chmod 600 /root/.ssh/id_rsa.pub

# Start ssh-agent
RUN eval `ssh-agent -s` && \
    ssh-add ~/.ssh/id_rsa

# Configure git
RUN git config --global url."git@gitlab.com:".insteadOf "https://gitlab.com/"
RUN git config --global url."git@dev.hackerman.me:".insteadOf "https://dev.hackerman.me/"

# Download required dependencies
ENV GOPRIVATE="dev.hackerman.me/artheon/*"
RUN go get -u -v; go mod tidy

# Build
RUN CGO_ENABLED=0 GO111MODULE=on go build -o /artheon-rpc

# Remove ssh keys
RUN rm -rf /root/.ssh/

############ RUN ####################
FROM alpine:3.8

COPY --from=builder /artheon-rpc /usr/local/bin/

RUN ls -lah /usr/local/bin/

#RUN mkdir -p /opt/artheon/
#COPY artheon-rpc.yaml /opt/artheon/

WORKDIR /tmp

RUN ls -lah /usr/local/bin/

ENTRYPOINT [ "/usr/local/bin/artheon-rpc" ]