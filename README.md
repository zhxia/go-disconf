# go-disconf

 - 基于golang 与 redis开发
 - 通过redis的pubsub特性
 
# 具体工作流程
  - server 与 agent分别通过redis的消息机制进行通讯
  - server端下发配置，经过redis广播到所有的agent
  - agent端连接redis，接收redis的广播消息，并解析命令，执行响应的操作。
  - agent在命令执行完毕时，会在本次的配置下发动作结果中写入agent所在的机器ip，
  并每隔一段时间发送心跳信息

# 构建
 - 通过gb 进行编译： gb build
# 运行
 
 - 启动redis Server
 - 启动server端: 
 
     <code>./server -r 192.168.33.200:6379 -h 192.168.33.100:8487 </code>
 - 启动agent端：
 
     <code>./agent -r 192.168.33.200:6379 -d /data/config </code>
 
 
