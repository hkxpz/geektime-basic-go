### 1. 实现思路

* 点赞接口推送kafka, 异步增加文章点赞
    * interactive service Like
      方法: https://github.com/hkxpz/geektime-basic-go/blob/master/webook/interactive/service/interactive.go
    * producer: https://github.com/hkxpz/geektime-basic-go/blob/master/webook/interactive/events/article/change_like.go
* consumer
  批量消费: https://github.com/hkxpz/geektime-basic-go/blob/master/webook/interactive/events/article/change_like.go

### 2. 点赞过程

* 批量插入数据库
* 操作缓存
    * 本地缓存 bizID, 如果本地缓存存在bizID, redis自增/减 1
        *
        BatchIncrLike/BatchDecrLike: https://github.com/hkxpz/geektime-basic-go/blob/master/webook/interactive/repository/interactive.go
    * 本地缓存没有 bizID, 操作数据库统计文章点赞数并缓存 redis
        *
        checkExists: https://github.com/hkxpz/geektime-basic-go/blob/master/webook/interactive/repository/interactive.go
        *
        BatchIncrLikeCntIfPresent/BatchDecrLikeCntIfPresent: https://github.com/hkxpz/geektime-basic-go/blob/master/webook/interactive/repository/cache/interactive.go
    * redis 存储 1000 个 bizID, 超出部分删除, 同步删除本地缓存的 bizID
        *
        BatchSetLikeCnt: https://github.com/hkxpz/geektime-basic-go/blob/master/webook/interactive/repository/cache/interactive.go

### 3. 压测

* 本机环境: 4c16
* 远端环境: 2c4
* 数据库, kafka 在远端