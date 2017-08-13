# `docker exec` web console

This image is originally inspirated by [this gist](https://gist.github.com/Humerus/0268c62f359f7ee1ee2d).

The container use `socat` to bound Docker socket to a TCP port, so that I could avoid to change the `http.Post` part in the Go server to talk directly to a UNIX socket.

You can launch the container in this way:

```
docker run \
	--name docker-exec-web-console \
	-p 9999:8888 \
	-v /var/run/docker.sock:/var/run/docker.sock \
	bitbull/docker-exec-web-console
```

Then you can reach the console at the url `http://localhost:9999`

It's possible to pass a context path to which the container will responds, using `CONTEXT_PATH` environment variable:

```
docker run \
	--name docker-exec-web-console \
	-p 9999:8888 \
	-e "CONTEXT_PATH=/webconsole"
	-v /var/run/docker.sock:/var/run/docker.sock \
	bitbull/docker-exec-web-console
```

With the above example, the console will be reachable at the url `http://localhost:9999/webconsole`

You can select the container to exec into passing its id directly via `cid` querystring parameter ( eg. `http://localhost:9999?cid=<container id>` ) or in the prompt that will show at page load.

You can pass the command to execute passing it via `cmd` querystring parameter ( eg. `http://localhost:9999?cid=<container id>&cmd=/bin/sh` ), otherwise it default to `/bin/bash`.

## Build the image

The image is based on `alpine:latest`, so you need to compile the server for Linux with all libs statically linked:

```
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .
```

then you can build as usual:

```
docker build -t bitbull/docker-exec-web-console .
```