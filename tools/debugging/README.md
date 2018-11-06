# Debugging protoc plugins

Sometimes we need debug our protoc plugins with standart IDE debug tool.
We have two ways how we can do it. First we can mock input content and start debug session, second use delve remote debug server.
This package provide script and some instruction how we can use second way.

* First of all you need build yor plugin and patch our gentool docker image see example below.
P.S. Sometimes plugins executed too fast and we late attach debugger to plugin, our advice add ```time.Sleep(1 * time.Second)``` 
at the beginning in your main function 

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
     
	This example shows how you can prepare docker file to patch protoc-gen-atlas-validate plugin in latest gentool images
	to build this example you need use command presented below. Be careful in this command we use some docker tricks,
	use -f to specif Docker file and pass project path as build scope to Docker daemon

	```bash 
	   docker build -f Dockerfile -t infoblox/atlas-gentool:test-plugin $GOPATH/src/github.com/infobloxopen/protoc-gen-atlas-validate
	```

	If you need debug another plugin you need change Docker file to build another plugin, and change scope dir in bash command

* Secondly package contains find-process script it can use for attach to running process,
and start debug session it uses some tricks with ```ps``` linux command to find process id(pid) and attach debug server to this process. 
If you need debug application in docker you can use this script but you must start it from root user(because as a default behavior docker start all process from root).
As first argument you must provide name go bin file see example below.
	``` bash
	./find-process protoc-gen-swagger
	```
	This command start daemon and when application with this name start it will transfer control to delve server.

* Thirdly you need setup Go remote debugger in your IDE(examples explain how setup it in Goland IDE)
     * Open run section and chose ```Edit configurations```
     * In upper left corner press ```+``` button and find ```Go Remote``` in list
     * Keep configuration as is
 
 
If you done all right you will get ```infoblox/atlas-gentool:test-plugin``` images with your plugin version
use it to generate code. Be careful script must running when you start protoc ```command``` 