
### **一、部署文件**
##### 1、克隆代码到本地
git clone http://git.0717996.com/Tomas/dezhoupoker.git

##### 2、进入RedBlack-War文件夹
cd dezhoupoker

##### 3、编译
go build -o dezhoupoker ./main.go  `要权限`

##### 4、后台运行
nohup ./dezhoupoker >load.log 2>&1 &  `要权限`

##### 5、查看是否运行成功
cat load.log

###### 如果看到日志文件输出以下数据代表成功启动~
2020/04/18 14:36:36 [release] Leaf 1.1.3 starting up
2020/04/18 14:36:36 [debug  ] Connect DataBase 数据库连接SUCCESS~
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           

### **二、項目所需套件支持**
###### 1、Go语言配置环境   `go version go1.13 linux/amd64`
###### 2、Mongo数据库     `MongoDB server version: 4.0.12`


### **三、配置文件位置及文件名稱**
##### 1、Mongo数据库名：dezhoupoker
##### 2、文件名称: `dezhoupoker/conf/server.json`
##### 3、日志文件：`load.log`  `路径为：编译好可执行文件同级`
##### 4、服务配置信息：
```
{
  "LogLevel": "debug",
  "LogPath": "",
  "WSAddr": "0.0.0.0",
  "Port": "1344",                                               服务器端口
  "HTTPPort": "3344",                                           运营后台数据对接端口
  "MaxConnNum": 20000,
    
  "MongoDBAddr": "0.0.0.0:27017",                               Mongo数据库连接地址
  "MongoDBAuth": "",                                            Mongo认证(可不填默认admin)
  "MongoDBUser": "",                                            Mongo连接用户名
  "MongoDBPwd": "",                                             Mongo连接密码

  "TokenServer": "http://172.16.100.2:9502/Token/getToken",     中心服Token
  "CenterServer": "172.16.100.2",                               中心服地址 
  "CenterServerPort": "9502",                                   中心服端口 
  "DevKey": ""                                                  devKey
  "DevName": "",                                                devName
  "GameID": ""                                                  gameID
  "CenterUrl": "ws://172.16.1.41:9502/"                         连接中心服URL地址
}
```

