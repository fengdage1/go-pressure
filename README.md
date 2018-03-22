# go-pressure
压力测试工具
build:
go build pressure.go

参数:
(必填)
-n 总量
-c 并发数
-u url
(选填)
-t 超时(秒，默认5秒)
-e 输出错误


example1: pressure -n 10000 -c 1000 -u http://www.google.com
example1: pressure -n 10000 -c 1000 -u http://www.google.com -t 10 -e


result:
---------------------------------
timeused 用时
complete 完成数量
failed 失败数量
successpercent 成功率
recved 收到字节总数
resp/sec 服务器每秒吞吐量
