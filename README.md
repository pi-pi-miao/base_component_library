## log库
```go
日志库，基础组建使用的是zap
在zap基础之上做了优化
1、支持按照天进行分割为目录日志
2、支持每天按照日志大小分割为不同的文件
3、支持保存多少天的日志
4、日志支持异步和同步输出
5、方便后续改造

// 调用方式
// filePath:日志输出路径
// level：日志级别
// maxSize：每个文件大小，超过大小会分割M
// maxBackup： 保存多少天的日志，多了删除
// maxAge ： 目前没用到，和maxBackup设置一样即可
func init(){
    logger.InitLog(filePath, level, maxSize, maxBackup, maxAge)
}
  
```