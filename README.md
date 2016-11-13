## Distributed tracing demo in Golang
Demo code for my presentation on distributed tracing - get a grasp on production.


### Dependencies
The project needs open-zipkin and mongodb to be up and running. You can launch them locally via docker:
```docker run -d -p 9410:9410 -p 9411:9411 openzipkin/zipkin```
```docker run -p 27017:27017 -d mongo```

### Install
```
go install
```

### Launch
```
cd $GOPATH/bin
```

```
./go-distributed-tracing-demo
```

Note: You can also override default parameters, please refer to help for more details i.e.
```
./go-distributed-tracing-demo --help
```
Insert image here
