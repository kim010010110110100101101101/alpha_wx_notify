# alpha_wx_notify
币安alpha互动微信通知

# 功能
爬取空投日历，把今日的空投信息通过调用方糖的接口推送到微信。
只要今日空投有任何变化，就会发通知给你：
![image](https://github.com/user-attachments/assets/a4c55f0d-633b-438a-a709-c25a82599e36)
![image](https://github.com/user-attachments/assets/6641ab15-eeb6-4f79-a7eb-1f6e713bc60c)


# 使用方法
在config.json中填入你的sendKey，有多个就填入多个。（sendKey获取方法：用微信打开https://sct.ftqq.com/sendkey）

填完之后，直接启动
linux: nohup ./alpha_wx_notify &

window：双击打开

config.json的配置：
{
    "sendkeys": [""], #sendkey
    "interval": 5, # 间隔多少分钟检测一次
    "fiterTge": true # 是否过滤tge活动
}


# 编译
go build

# GitHub Actions配置
如果你使用GitHub Actions自动运行此程序，需要在仓库的Secrets中添加以下环境变量：

1. `CF_COOKIE` - CloudFlare验证Cookie，用于绕过网站的反爬虫保护
2. `USER_AGENT` - 浏览器用户代理字符串

获取这些值的方法：
1. 在浏览器中打开开发者工具（F12）
2. 访问 https://alpha123.uk/zh/index.html
3. 在网络标签页中找到任意请求
4. 在请求头中复制Cookie和User-Agent的值
5. 在GitHub仓库的Settings -> Secrets -> Actions中添加这些值

注意：CloudFlare Cookie可能会定期过期，如果GitHub Actions开始报403错误，请更新Cookie值。
