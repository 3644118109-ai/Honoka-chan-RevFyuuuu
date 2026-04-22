# SIF私服公网化注意事项
**注意**<br><hr>
本项目的唯一目的是技术研究与学习。

This project is solely for educational and research purposes.

不提供任何版权内容
本项目不包含任何游戏客户端、资源包、图片、音乐或商标。用户需自行准备已购买的游戏文件。

不鼓励盗版
本项目旨在帮助已合法拥有《LoveLive! 学园偶像祭》的用户，在官方服务器关闭后，仍能研究其技术实现。请勿将本项目用于任何商业用途。

版权归属
所有游戏相关资源（角色、音乐、插画、代码等）的版权归 Bushiroad / KLab / 盛趣游戏 及其相关权利人所有。

无任何担保
本项目按“现状”提供，不提供任何明示或暗示的担保。使用本项目所产生的任何后果由用户自行承担。

侵权联系
如您认为本项目的任何部分侵犯了您的合法权益，请联系 [此邮箱](3644118109@qq.com)，我们将在核实后尽快处理。

- 开始之前你需要先去观看阅读,该项目基于Github上的 [honoka-chan](https://github.com/YumeMichi/honoka-chan) 项目进行开发调整，编译和环境部署可参考 [此视频](https://www.bilibili.com/video/BV1Fk4y1S7HA/?share_source=copy_web&vd_source=a261a81c7cc6e30b70444c80a4015330) ，并使用了Codex工具进行辅助。<hr>

- [SIF私服公网化注意事项](#sif私服公网化注意事项)
	- [一.服务端配置](#一服务端配置)
		- [OSS域名加速](#oss域名加速)
		- [服务端错误提示配置和日志生成](#服务端错误提示配置和日志生成)
		- [端口号配置](#端口号配置)
		- [安全项设置](#安全项设置)
	- [二.客户端注意事项](#二客户端注意事项)
		- [1.资源包修改](#1资源包修改)
		- [2.你可能需要的](#可能你需要的)
	- [最后](#最后)
 <hr>

## 一.服务端配置
 ### OSS域名加速 
 在原Git上有提到，我们可以使用OSS服务来将游戏本体和额外资源分开放置，这样可以很大程度缓解服务器和带宽压力，但此处注意，在`sif_cdn_server`中，务必不要使用https，应优先考虑`http://SIF.oss-cn-hangzhou.aliyuncs.com`这样的域名方式，且注意，**域名加速**或者其他类似服务有可能导致资源下载服务不稳定，造成例如通信中断等提示。<br>
 ###  服务端错误提示配置和日志生成
 在服务端的主文件夹中的子文件夹`你的主文件夹\handler`中的`Download.go`中你可以通过自己修改和写入其他代码段，例如
 ```go
 if wd, err := os.Getwd(); err == nil {
		_ = os.WriteFile(filepath.Join(wd, "update_resp.json"), resp, 0644)
		if pretty, err := json.MarshalIndent(updateResp, "", "  "); err == nil {
			_ = os.WriteFile(filepath.Join(wd, "update_resp_debug.json"), pretty, 0644)
		}
		log.Printf("[DL] update_resp write dir=%s", wd)
	}
 ```
 来在服务端生成一段日志并通过
 ```go
		for _, pkg := range pkgInfo {
			url := fmt.Sprintf("%s/%s/archives/%d_%d_%d.zip", SifCdnServer, downloadReq.TargetOs, pkgType, pkg.Id, pkg.Order)
			size := pkgSizeOrRemote(pkg.Size, url)
			if size == 0 {
				log.Printf("[DL] missing url=%s", url)
				continue
			}
			pkgList = append(pkgList, model.UpdateRes{
				Size:    size,
				URL:     url,
				Version: ver,
			})
        }
 ```
类似这种写法来定义你需要看到的DL状态的内容 
 ### 端口号配置 
 我们可以通过`config.json`中的`listen_addr`项来自定义你想要开放的端口号，同时尽可能的避免被公网扫描器扫描，但同时，你需要在资源包的各种***serverinfo***项中指定你的端口，例如：``http://192.168.0.100:7789/static``（注意操作中修改为自己的IP）的格式进行保存。
 ### 安全项设置
 对于使用80端口的服务端我们提供了一个``antiscanner.go``作为中间件，它将在***middleware***文件夹中被加载，它可以被用于防扫描器扫描或骚扰，并很大程度上防止扫描器在你的服务端里面刷屏,并
  - 直接拦截这些方法：OPTIONS、PROPFIND、PRI、TRACE、CONNECT
  - 直接拦截常见扫描路径前缀：/.git、/phpmyadmin、/nacos、/hudson、/v1/models、/v2/keys、/mcp、/sse、/+CSCOE+ 等
  - 直接拦截常见探测文件：/robots.txt、/favicon.ico、/nice ports,/Trinity.txt.bak
  - 返回 404（减少暴露特征）
 同时有限流中间件``ratelimit.go``，但我们建议直接使用服务端**Config.json**文件中的访问数限制，因为它不仅限于扫描器
<hr>

## 二.客户端注意事项
 ### 1.资源包修改
 在原视频中有提到只需要提取``99_0_115``文件中的***serverinfo***然后进行对应IP地址修改再重打包回源压缩包这步，但在实际操作中，此步骤还需要修改其他url处将IP回填为自己服务端主机的IP，这步是为了最大的排除可能的兼容性和连接失败问题，具体需要像如下引用中修改
 ```json
 {
  {"name": "server_information",
  "domain": "http://你的IP",
  "maintenance_uri": "http://你的IP/resources/maintenace/maintenance.php",
  "update_uri": "http://你的IP/resources/maintenace/update.php",
  "login_news_uri": "http://你的IP/webview.php/announce/index?0=",
  "locked_user_uri": "http://你的IP/webview.php/static/index?id=13",
  "server_version": "你设置的版本号 原为97.1.1",
  "end_point": "/main.php",
  "consumer_key": "lovelive_test",
  "application_id": "626776655",
  "max_connection": 10,"region": "392",
  "date": "1665597600","application_key": "3d55330eb08835df468ab56a261a8cb6",
  "api_uri": {"/battle/startWait": "http://你的IP/main.php/battle/startWait",
  "/battle/endWait": "http://你的IP/main.php/battle/endWait",
  "/duty/startWait": "http://你的IP/main.php/duty/startWait",
  "/duty/endWait": "http://你的IP/main.php/duty/endWait",
  "/duty/privateStartWait": "http://你的IP/main.php/duty/privateStartWait"}}
 }
 ```
 也就是将原指向盛趣的旧IP地址定向到你自己的服务器，但需要注意，根据实际解包发现，不仅在`115`包体内发现了相关的`Serverinfo`段，
 如下包体也存在此段落
 - 99_0_10
 - 99_0_103
 - 99_0_108
 - 99_0_113
 - 99_0_15
 - 99_0_19
 - 99_0_23
 - 99_0_28
 - 99_0_33
 - 99_0_38
 - 99_0_43
 - 99_0_48
 - 99_0_5
 - 99_0_53
 - 99_0_58
 - 99_0_63
 - 99_0_68
 - 99_0_73
 - 99_0_78
 - 99_0_82
 - 99_0_87
 - 99_0_92
 - 99_0_97
 为了稳定性和不确定的可能性，建议同时修改并检查以上所有包体以内的IP
 ### 可能你需要的
 对于一些对服务端配置尚不明确或者不熟练的伙伴，在此我们也提供了一个释放80端口的一个小工具，因为部分环境下，系统会默认占用掉你的80端口，因此你只需要复制以下代码进你的文本文档并将其改为```.bat```文件，并`以管理员身份运行`此脚本，即可释放80端口
 ```powershell	 
@echo off
title Free Port 80
color 0A
echo =============================================
echo       Free Port 80 - System Process Fix
echo =============================================
echo.

net session >nul 2>&1
if %errorLevel% neq 0 (
    echo [ERROR] Please run as administrator!
    pause
    exit /b 1
)

echo [1/3] Checking port 80...
echo.

netstat -ano | findstr :80 | findstr LISTENING

for /f "tokens=5" %%a in ('netstat -ano ^| findstr :80 ^| findstr LISTENING') do (
    set PID=%%a
    goto :found
)

if not defined PID (
    echo Port 80 is not occupied.
    pause
    exit /b 0
)

:found
echo Found process with PID: %PID%

if "%PID%"=="4" (
    echo.
    echo [2/3] System process detected. Stopping HTTP service...
    echo.
    net stop http /y
    timeout /t 2 /nobreak >nul
) else (
    echo.
    echo [2/3] Killing process %PID%...
    taskkill /f /pid %PID%
)

echo.
echo [3/3] Verifying port 80...
echo.
timeout /t 2 /nobreak >nul
netstat -ano | findstr :80 | findstr LISTENING

if %errorLevel% equ 0 (
    echo.
    echo [WARNING] Port 80 still in use.
    echo Try: sc config http start= disabled
    echo Then restart your computer.
) else (
    echo.
    echo [SUCCESS] Port 80 is now free!
)

echo.
pause

```
# 最后
***Thank you for reading to this point***