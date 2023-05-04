# redis-operator

_因为当前的工作主要就是负责 redis-opertaor，所以自己想写一个 operator 可以管理 redis 单机，哨兵，集群三种模式作为个人练习。此代码有参考 https://github.com/ucloud/redis-cluster-operator_

## 设计主旨
简单明了，一个 crd.yaml 就能拉起一个 redis 单机/集群/哨兵集群。因此没有用单独的 secret.yaml 存放密码，也没用单独的 configmap.yaml 存放配置，都是直接在 crd.yaml 里面配置。

## 单机模式
设计的十分简单，sts 启动一个 redis，configMap 保存配置，pvc 持久化数据，svc 用于访问。
```
$ k apply -f config/samples/redis_v1_redisstandalone.yaml
redisstandalone.redis.my.domain/redisstandalone-sample created

$ k get pod
NAME                                           READY   STATUS    RESTARTS   AGE
redisstandalone-sample-0                       1/1     Running   0          9m20s

$ k get svc
NAME                                   TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                   AGE
redisstandalone-sample                 ClusterIP   None             <none>        6379/TCP                  9m30s
redisstandalone-sample-nodeport        NodePort    10.107.123.14    <none>        6379:31676/TCP            9m30s

$ k get cm
NAME                               DATA   AGE
redisstandalone-sample-configmap   1      9m40s

$ redis-cli -h xxx -p 31676
10.20.9.60:31676> keys *
(error) NOAUTH Authentication required.
10.20.9.60:31676> 
```