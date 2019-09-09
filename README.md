### 连接协议
目前实现了HTTP协议，local和server直接连接

### 数据通信协议
包[head,body],head为4字节包含当前加密包长度，body为加密的byte字节，读取完成后进行解密

### 关于加密
默认为自己实现的OORR，主要是对byte进行或与运算，修改所有的byte达到加密的目的，速度很快，耗费资源少

### TODO
SOCKS5 协议
