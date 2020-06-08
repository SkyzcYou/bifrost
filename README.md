# 项目介绍
go-nginx-conf-parser 是基于golang语言开发的项目，它目前还处于测试阶段，用于对Nginx配置文件解析并提供配置文件展示和修改的接口，支持json、字符串格式与golang结构相互转换。该项目持续更新中。目前可用版本为[v0.0.3](https://github.com/ClessLi/go-nginx-conf-parser/tree/v0.0.3-alpha.2)。

# 项目特点
支持将配置文件、json数据、字符串与配置结构体相互转换

配置结构体支持增加、删除、查询（暂实现查询server上下文结构体）

提供配置文件展示和修改的接口

# 使用方法
## 下载地址
Windows:  [go-nginx-conf-parser.v0_0_3.win_x64](https://github.com/ClessLi/go-nginx-conf-parser/releases/download/v0.0.3-alpha.2/go-nginx-conf-parser.v0_0_3-alpha.2.win_x64.zip)

Linux: [go-nginx-conf-parser.v0_0_3.linux_x64](https://github.com/ClessLi/go-nginx-conf-parser/releases/download/v0.0.3-alpha.2/go-nginx-conf-parser.v0_0_3-alpha.2.linux_x64.zip)

## 应用配置
> configs/ng-conf-info.yml
```yaml
NGConfigs:
  -
    name: "nginx-conf-test"
    relativePath: "/ng_conf"
    port: 18080
    confPath: "/usr/local/openresty/nginx/conf/nginx.conf"
    nginxBin: "/usr/local/openresty/nginx/sbin/nginx"
#  -
#    name: "ng-conf-test2"
#    relativePath: "/ng_conf"
#    port: 28080
#    confPath: "/GO_Project/src/go-nginx-conf-parser/test/config_test/nginx.conf"
#    nginxBin: "xxxxxxxxxxxx/nginx"
DBConfig:
  DBName: "ng_conf_admin"
  host: "127.0.0.1"
  port: 3306
  protocol: "tcp"
  user: "ngadmin"
  password: "ngadmin"
logConfig:
  logDir: "./logs"
  level: 2
```

## 命令帮助
```
> ./go-nginx-conf-parser -h
Usage of ./go-nginx-conf-parser:
  -f conf
    	go-nginx-conf-parser ng-conf-info.y(a)ml path. (default "./configs/ng-conf-info.yml")
  -h help
    	this help
```

## 应用接口调用方式
### 访问认证接口
####1.登录认证接口
接口地址：
```
http://<Host>:<Port>/login/<username>/<password>
```
返回格式：
```
json
```
请求方式
```
http get
```
请求示例：
```
http://127.0.0.1:18080/login/ngadmin/ngadmin
```
请求参数：
无
返回参数说明；

| 名称 | 类型 | 说明 |
| :-: | :-: | :- |
| json返回示例 | - | - |

json返回示例：
```json
{
  "message": "null",
  "status": "success",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1OTE2MDcwMzcsImlhdCI6MTU5MTYwMzQzNywidXNlcl9pZCI6MSwicGFzc3dvcmQiOiJuZ2FkbWluIiwidXNlcm5hbWUiOiJuZ2FkbWluIiwiZnVsbF9uYW1lIjoibmdhZG1pbiIsInBlcm1pc3Npb25zIjpbXX0.l5qE1sMBD9VJHspzXlhHNmHhbZiF00YlCafUIsIEJpo"
}
```

####2.校验会话token接口
接口地址：
```
http://<Host>:<Port>/verify/<token>
```
返回格式：
```
json
```
请求方式
```
http get
```
请求示例：
```
http://127.0.0.1:18080/verify/eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1OTE2MDcwMzcsImlhdCI6MTU5MTYwMzQzNywidXNlcl9pZCI6MSwicGFzc3dvcmQiOiJuZ2FkbWluIiwidXNlcm5hbWUiOiJuZ2FkbWluIiwiZnVsbF9uYW1lIjoibmdhZG1pbiIsInBlcm1pc3Npb25zIjpbXX0.l5qE1sMBD9VJHspzXlhHNmHhbZiF00YlCafUIsIEJpo
```
请求参数：
无
返回参数说明；

| 名称 | 类型 | 说明 |
| :-: | :-: | :- |
| json返回示例 | - | - |

json返回示例：
```json
{
  "message": "Certified user \u003cngadmin\u003e",
  "status": "success"
}
```

####3.更新会话token接口
接口地址：
```
http://<Host>:<Port>/refresh/<token>
```
返回格式：
```
json
```
请求方式
```
http get
```
请求示例：
```
http://127.0.0.1:18080/refresh/eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1OTE2MDcwMzcsImlhdCI6MTU5MTYwMzQzNywidXNlcl9pZCI6MSwicGFzc3dvcmQiOiJuZ2FkbWluIiwidXNlcm5hbWUiOiJuZ2FkbWluIiwiZnVsbF9uYW1lIjoibmdhZG1pbiIsInBlcm1pc3Npb25zIjpbXX0.l5qE1sMBD9VJHspzXlhHNmHhbZiF00YlCafUIsIEJpo
```
请求参数：
无
返回参数说明；

| 名称 | 类型 | 说明 |
| :-: | :-: | :- |
| json返回示例 | - | - |

json返回示例：
```json
{
  "message": "null",
  "status": "success",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1OTE2MDc3MDEsImlhdCI6MTU5MTYwMzQzNywidXNlcl9pZCI6MSwicGFzc3dvcmQiOiJuZ2FkbWluIiwidXNlcm5hbWUiOiJuZ2FkbWluIiwiZnVsbF9uYW1lIjoibmdhZG1pbiIsInBlcm1pc3Npb25zIjpbXX0.fDoe4v37XyjmrK4wnfhOUnePwJLdszYXveOfoRXyUj8"
}
```

###Nginx配置接口
####1.查询nginx配置接口
接口地址：
```
http://<Host>:<Port>/<relativePath>/<token>
```
注：\<relativePath>为ng管理工具配置中NGConfigs列表各自元素的relativePath子参数值。
返回格式：
```
json
```
请求方式
```
http get
```
请求示例：
```
http://127.0.0.1:18080/ng_conf/eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1OTE2MDc3MDEsImlhdCI6MTU5MTYwMzQzNywidXNlcl9pZCI6MSwicGFzc3dvcmQiOiJuZ2FkbWluIiwidXNlcm5hbWUiOiJuZ2FkbWluIiwiZnVsbF9uYW1lIjoibmdhZG1pbiIsInBlcm1pc3Npb25zIjpbXX0.fDoe4v37XyjmrK4wnfhOUnePwJLdszYXveOfoRXyUj8?type=json
```
请求参数：

| 名称 | 必填 | 类型 | 说明 |
| :-: | :-: | :-: | :- |
| type | 否 | string | 默认为字符串格式，指定“string”为字符串格式；“json”为json格式|

返回参数说明；

| 名称 | 类型 | 说明 |
| :-: | :-: | :- |
| json返回示例 | - | - |

json返回示例：

`string格式返回示例`
```json
{
  "message": ["# user  nobody;\n", "worker_processes 1;\n", "# error_log  logs/error.log;\n"],
  "status": "success"
}
```

`json格式返回示例`
```json
{
  "message": {
    "config": {
      "value": "/usr/local/openresty/nginx/conf/nginx.conf",
      "param": [{
        "comments": "user  nobody;",
        "inline": true
      }, {
        "name": "worker_processes",
        "value": "1"
      }, {
        "comments": "error_log  logs/error.log;",
        "inline": false
      }]
    }
  },
  "status": "success"
}
```

####2.更新nginx配置接口
接口地址：
```
http://<Host>:<Port>/<relativePath>/<token>
```
注：\<relativePath>为ng管理工具配置中NGConfigs列表各自元素的relativePath子参数值。
返回格式：
```
json
```
请求方式
```
http post
```
请求示例：
```shell script
curl 'http://127.0.0.1:18080/ng_conf/eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1OTE2MDc3MDEsImlhdCI6MTU5MTYwMzQzNywidXNlcl9pZCI6MSwicGFzc3dvcmQiOiJuZ2FkbWluIiwidXNlcm5hbWUiOiJuZ2FkbWluIiwiZnVsbF9uYW1lIjoibmdhZG1pbiIsInBlcm1pc3Npb25zIjpbXX0.fDoe4v37XyjmrK4wnfhOUnePwJLdszYXveOfoRXyUj8'\
    -X POST\
    -F "data=@/tmp/ng_conf.json"
```
请求参数：
无

请求表单参数：

| 名称 | 必填 | 类型 | 说明 |
| :-: | :-: | :-: | :- |
| data | 必填 | file | 配置文件数据json文件 |

返回参数说明；

| 名称 | 类型 | 说明 |
| :-: | :-: | :- |
| json返回示例 | - | - |

json返回示例：
```json
{
  "message":"Nginx ng update.",
  "status":"success"
}
```