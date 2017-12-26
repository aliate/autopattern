## auto create kibana index patterns use logstash

### 编译alpine版本

> docker run --rm -v `pwd`:/go/src/github.com/aliate/autopattern -w /go/src/github.com/aliate/autopattern golang:1.8-alpine go build -v

### 环境变量

| 变量名               | 说明                 | 默认值    |
| ------------------- | ------------------- | -------- |
| AK_PORT             | 监听端口              | 12345    |
| ES_HOSTS            | ES主机列表            |          |
| ES_INDEX_KEEP_DAYS  | INDEX保留多少天       | 30       |
| ES_INDEX_CLEAR_TIME | 清理过期INDEX的时间    | 0(晚12点) |
| KIBANA_INDEX        | kibana使用的index名称 | .kibana |
