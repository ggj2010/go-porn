# 91pornTrack
    国外pornhub下载要收费，so只能对国内91porn的下手了
    
* 支持下载91热门视频    
* 支持下载91某个用户的所有公开视频
* 支持下载91某个网址的视频

## Dependency&Installation
   1、视频合成依赖ffmpeg，需要先安装ffmpeg
    
    brew install ffmpeg
    
   2、依赖chrome浏览器
   
     部分网页是内容是js动态渲染的，所以用了go的组件 chromedp
     
   3、翻墙 
   
     部分网络可能访问不了91pron,所以执行程序前最好翻墙下
     https://portal.shadowsocks.nz/aff.php?aff=20093
   4、登陆
   
      部分视频下载需要登陆状态，所以注册91pron账号，再登陆下。
      
   5、build
   
    go build track.go 
## Examples
   
### 所有参数
     ./track --help
### 下载热门视频
    ./track -t=1
    视频类型
    0:所有 
    1：当前最热 
    2：本月最热 
    3：10分钟以上 
    4：20分钟以上 
    5：本月收藏
    6：收藏最多 
    7：最近加精 
    8：高清 
    9：上月最热 
    10：本月讨论 
### 下载某个用户所有视频
    ./track -uid=c24dDoGZBAnwUtBbHweSJB8W6ACe8c7sJyQOJ9Af4DQ4sxul
### 下载某个网址视频
    ./track -vid=8ee92162ba6b47e1dfcf
### 修改文件保存目录
   默认文件保存目录 /data/91movie/
   
    ./track -p /data/92 -vid=8ee92162ba6b47e1dfcf  
    
   目前有一定概率文件下载失败，已经下载过的文件是不会再次下载的，需要手动删除后才会覆盖
## 免责声明
    本代码仅用于学习，下载后请勿用于商业用途!
