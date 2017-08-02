# Docker Regsiter

## 安装

```
go get git.wolaidai.com/DevOps/register
cd $GOPATH/src/git.wolaidai.com/DevOps/register

make build
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

#####  ######  ####  #  ####  ##### ###### #####
#    # #      #    # # #        #   #      #    #
#    # #####  #      #  ####    #   #####  #    #
#####  #      #  ### #      #   #   #      #####
#   #  #      #    # # #    #   #   #      #   #
#    # ######  ####  #  ####    #   ###### #    #

NAME:
    start - start a new consul-register

USAGE:
    start [command options] [arguments...]

OPTIONS:
   --concurrency value      concurrency number (default: 10)
   --docker.endpoint value  Docker Conn EndPoint (default: "unix:///var/run/docker.sock")
   --consul.host value      consul server host
   --consul.port value      consul server port (default: "8500")
   --marathon.host value    marathon server host (default: "192.168.20.2")
   --marathon.port value    marathon server port (default: "8080")
   --marathon.dyups value   marathon of orange api url (default: "/base-service/loadbalance/orange")
   --dyups.port value       orange server port (default: "18081")
   --dyups.url value        orange upstream url (default: "/upstream-admin")
   --check.second value     timer check upstream duration (default: 10)
   --check.enable           timer check upstream is enable
   --log.level value        Only log messages with the given severity or above. Valid levels: [debug, info, warn, error, fatal] (default: "info")
```



