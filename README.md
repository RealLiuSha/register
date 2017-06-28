# Docker Regsiter

## Install

```
go get github.com/itchenyi/register
cd $GOPATH/src/github.com/itchenyi/register
go build
```

## 基于Docker事件关联Consul

1. `start`, `unpause` 关联注册逻辑

```
                            event
                              |
                              |
                          http check
                              |
                              |
               success <-------------> failed
                  |                      |
                  |                      |
               register              continue
                  |
            Sync Upstream
```

2. `pause`, `die` 关联反注册逻辑

```
                            event
                              |
                              |
                           success 
                              |        
                           deRegister
                              |
                        Sync Upstream
```


## 程序参数

```
  -check.upstream int
    	timer check upstream duration (default 10)
  -concurrency int
    	concurrency number (default 10)
  -consul.host string
    	consul server host
  -consul.port string
    	consul server port (default "8500")
  -docker.endpoint string
    	Docker Conn EndPoint (default "unix:///var/run/docker.sock")
  -dyups.port string
    	orange server port (default "18081")
  -dyups.url string
    	orange upstream url (default "/upstream-admin")
  -log.format value
    	Set the log target and format. Example: "logger:syslog?appname=bob&local=7" or "logger:stdout?json=true" (default "logger:stderr")
  -log.level value
    	Only log messages with the given severity or above. Valid levels: [debug, info, warn, error, fatal] (default "info")
  -marathon.dyups string
    	marathon of orange api url (default "/base-service/loadbalance/orange")
  -marathon.host string
    	marathon server host (default "192.168.20.2")
  -marathon.port string
    	marathon server port (default "8080")
```



