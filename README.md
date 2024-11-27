# veteran

`veteran`是一个基于`raft`协议实现的浮动`IP`控制器，支持以下功能：

* 在基于`raft`协议在集群环境执行选举`leader`，并设置浮动`IP`
* 支持通过`restful`动态添加/删除节点
* 支持持久化成员地址，主机/服务重启后`member`连接信息不会丢失

# 配置文件

`veteran` 运行时读取以下配置

```json
{
  // api 监听端口
  "listen": "0.0.0.0:27000",
  // 数据持久化目录
  "store": "/opt/veteran",
  // raft 日志配置
  "raft_log": {
    "enable": true,
    "level": "info"
  },
  // 初始化成员列表，列表包含当前节点
  "initial_cluster": {
    "274174d4-de2c-53b2-a366-27088884c56c": "172.28.117.42:27010"
  },
  // 浮动 IP 配置
  "floating": {
    "iface": "ens3",
    "type": "alias",
    "address": "172.28.117.100/24"
  }
}
```
