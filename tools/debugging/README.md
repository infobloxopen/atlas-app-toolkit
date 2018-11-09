# Debugging protoc plugins

Sometimes we need debug our protoc plugins with standart IDE debug tool.
We have two ways of how we can do this:
 - we can mock input content and start debug session
 - use [delve](https://github.com/derekparker/delve) remote debug server
 
This package provides script and instruction how we can use second way.

## Step 1
First of all you need build yor plugin and patch [gentool](https://github.com/infobloxopen/atlas-gentool) docker image. Aee example below:

```dockerfile
     FROM golang:1.10.0 AS builder
     
     LABEL stage=server-intermediate
     
     WORKDIR /go/src/github.com/infobloxopen/protoc-gen-atlas-validate
     COPY . .
     RUN CGO_ENABLED=0 GOOS=linux go build -o /out/usr/bin/protoc-gen-atlas-validate main.go
     
     FROM infoblox/atlas-gentool:latest AS runner
     
     COPY --from=builder /out/usr/bin/protoc-gen-atlas-validate /usr/bin/protoc-gen-atlas-validate
     
     WORKDIR /go/src
```

**NOTE:** Sometimes plugins executes too fast and we are late attach debugger to plugin, our advice is to add `time.Sleep(1 * time.Second)` at the beginning in your plugin main function.

Example shows how you can prepare docker file to patch [protoc-gen-atlas-validate](https://github.com/infobloxopen/protoc-gen-atlas-validate) plugin in latest gentool images. To build this example you need use command presented below:

```bash 
docker build -f Dockerfile -t infoblox/atlas-gentool:test-plugin $GOPATH/src/github.com/infobloxopen/protoc-gen-atlas-validate
```
**NOTE:** Be careful with this command as it uses some docker tricks. It uses `-f` to specif Docker file and passes project path as build scope to Docker daemon

If you need to debug another plugin you need change Docker file to build another plugin, and change scope dir in bash command.

## Step 2

This package contains [find-process](find-process) script which can be used to attach to a running process
and start debug session. It uses some tricks with `ps` linux command to find process id(pid) and attach debug server to this process. 
If you need to debug application in docker you can use this script but you must start it from `root` user (as by default docker starts all process from root).

As an argument you must provide a name of go bin file. See example below:
``` bash
./find-process protoc-gen-swagger
```
This will start a daemon and when application with given name starts deamon will transfer control to `delve` server.

## Step 3

You need setup Go remote debugger in your IDE. Example explain how setup it in Goland IDE:
 - Open run section and chose `Edit configurations`
 - In upper left corner press `+` button and find `Go Remote` in list
 - Keep configuration as is


Use `infoblox/atlas-gentool:test-plugin` created previously to generate the code. You should hit a breakpoint in your IDE.
