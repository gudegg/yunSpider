百度云爬虫
-----
[![Build Status](https://travis-ci.org/gudegg/yunSpider.svg?branch=user_conf)](https://travis-ci.org/gudegg/yunSpider)
> 网站:[http://www.fastsoso.cn/](http://www.fastsoso.cn/)

#### 安装使用
1. [安装go与设置gopath](https://github.com/astaxie/build-web-application-with-golang/blob/master/zh/01.2.md)
2. clone项目到gopath目录
3. 安装依赖
  * go get github.com/go-sql-driver/mysql
  * go get github.com/siddontang/go/log
  * go get github.com/garyburd/redigo/redis
  * go get github.com/Unknwon/goconfig
4. 修改config.ini配置文件中你自己的MySQL和Redis配置(可以不配Redis),Redis不配或者错误会默认将已经爬取的uk保存到hasIndexKeys.txt文件中,这样做防止重复爬同一位用户多次
5. 创建database名为baidu,然后执行baidu.sql
6. go run spider.go (当然也可以编译后运行,此时config.ini记得也复制到编译后的文件目录下,然后运行)

    
#### 直接下载编译好的使用
1. [下载](https://github.com/gudegg/yunSpider/releases)适合自己系统的版本
2. 修改config.ini配置文件，并且放到程序运行的目录下
3. linux: chmod +x spider,然后直接./spider运行或者nohup ./spider 1>log.out 2>err.out &以后台方式运行;win:直接点击spider.exe运行

#### 爬到数据使用
- 只需替换链接中uk,shareid或者album_id即可访问
    * 分享专辑链接类型：https://pan.baidu.com/wap/album/info?uk=1112219283&album_id=129732510768224935
    * 普通文件或者文件夹类型：https://pan.baidu.com/wap/link?uk=1112219283&shareid=2109459878
    

- 获取用户订阅:
 http://yun.baidu.com/pcloud/friend/getfollowlist?query_uk=%s&limit=24&start=%s&bdstoken=e6f1efec456b92778e70c55ba5d81c3d&channel=chunlei&clienttype=0&web=1&logid=MTQ3NDA3NDg5NzU4NDAuMzQxNDQyMDY2MjA5NDA4NjU=  
    (query_uk limit start是必须参数)
- 获取用户粉丝: 
http://pan.baidu.com/pcloud/friend/getfanslist?query_uk=%s&limit=24&start=%s&bdstoken=null&channel=chunlei&clienttype=0&web=1&logid=MTQ3NDAzNjQwNzg3OTAuNzM1MzMxMDUyMDczMjYxNA==
    (query_uk limit start是必须参数)
- 获取用户分享: 
http://pan.baidu.com/pcloud/feed/getsharelist?t=1474202771918&category=0&auth_type=1&request_location=share_home&start=0&limit=60&query_uk=224534490&channel=chunlei&clienttype=0&web=1&logid=MTQ3NDIwMjc3MTkxOTAuMzA1NjAzMzQ4MTM1MDc0MTc=&bdstoken=e6f1efec456b92778e70c55ba5d81c3d
    (query_uk limit start auth_type是必须参数)

 
> 上面3个连接请求必须带上Referer`("Referer", "https://yun.baidu.com/share/home?uk=23432432#category/type=0")`,否则请求不到json数据,
获取用户订阅和获取用户粉丝每次请求一次休眠2s的话可以无限制请求,对ip没要求,获取用户分享超坑,一个ip只能请求10次,并且休眠也没用.
因为没有那么多ip,我就去研究手机版的用户分享,手机版获取用户分享可以一次性连续请求60次,60次后必须休眠35s左右在继续请求就可以,不会像pc版那样必须换ip,
但是手机版只能请求网页源码,然后用正则进行匹配.

- 手机版分享:
 http://pan.baidu.com/wap/share/home?uk=2889076181&start=%s&adapt=pc&fr=ftw (uk:**每个百度网盘用户的唯一标示**,start:用户可能有上百上千个分享,必须分页,start默认从0开始,手机版默认分页是20个每页)
 

 
 > [java版本](http://zhangguodong.me/2016/09/18/%E7%88%AC%E5%8F%96%E7%99%BE%E5%BA%A6%E7%BD%91%E7%9B%98%E7%94%A8%E6%88%B7%E5%88%86%E4%BA%AB/)
 
#### License

yunSpide source code is licensed under the Apache Licence, Version 2.0 (http://www.apache.org/licenses/LICENSE-2.0.html).
