# 原神自动点击 - 使用 Golang 实现

## ✨ 功能：

1. 支持设置点击间隔时间和无动作超时自动关闭时间，单位为秒,不设置均为默认时间。
   
   示例1：命令行设置：
   ```
   click.exe 2 10
   表示点击间隔时间为2秒，超时自动关闭时间为10秒。
   ```
   示例2：双击后设置
   ![image](https://github.com/cggpro/GenshinClick/assets/120552503/17aabe2e-3b25-4122-a060-e8b75a4d674f)





2. 支持开启和关闭自动点击功能。

   - 按下 F9 键开启自动点击
   - 按下 F10 键关闭自动点击




3. 如果您想要自己修改请注意[robotgo](https://pkg.go.dev/github.com/go-vgo/robotgo@v0.110.1)安装GCC的要求
