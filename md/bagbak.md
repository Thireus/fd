![](../gif/bagbak.webp)

#### bagbak(ipa脱壳) 使用方法:
````
bakbag appname -devi string
````

#### appname:
- ios is app icon label
- android is app icon label
- fd lsapp 获取
- 如果获取不到,fd lsps 获取

#### -devi:
- default: usb
- -devi usb(usb devi)
- -devi u(usb devi)
- -devi local(local devi)
- -devi localhost(local devi)
- -devi ip:port(remote device)
- -devi 1234(devi id)

#### -pid
- 要脱壳的 process id (appname or pid)
- pid模式应该支持macos
- pid模式应该支持系统库，比如webkit

pid 例子:
````
bagbak -pid 123 -devi string
````


