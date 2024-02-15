<div align="center">

<h1>EMOZI</h1>
参考古埃及圣书体设计的一种基于颜文字的汉字抽象转写法<br>🐑🚬🧗👤🕸️😐🧗✍️👈🌞😨🏠🌹🧦😨👥🌹🔐😨💦⬅️☀️😨🏡💦💡🍉🌱🍵💡🧗🪓🍆👔😨🐶<br><br>

<img src="https://counter.seku.su/cmoe?name=emozi&theme=r34" /><br>

</div>

## 命令行工具
位于`cmd`文件夹。
```bash
用法: [-h|f|nr] [-db 字.db] [-d 🌹⁪😺‎🐴‫👩] [-e 好] 形声字选择1 形声字选择2 ...
  -a string
        添加一个汉字到附加库
  -d string
        解码颜文字为汉字序列
  -db string
        符合规范的查询数据库位置, 不存在则会自动释放到该路径. (default "字.db")
  -e string
        编码汉字序列为颜文字
  -f    强制解码并非由本程序生成的颜文字序列
  -h    显示帮助信息
  -i    指定汉字-a和带声调的拼音-p以计算其全局唯一ID
  -nr
        不随机选取所有读音相近的颜文字
  -p string
        带声调的拼音
  -r string
        指定欲编辑的部首
  -re string
        指定部首对应的颜文字
```
下面是一些用例。
### 编码
> 注意: 可以指定`-nr`参数从而使编解码结果唯一。
```bash
go run cmd/main.go -e 好
编码结果: 🌹⁪😺‎🐴‫👩
```
### 解码
```bash
go run cmd/main.go -d 🌹⁪😺‎🐴‫👩
解码结果: 好
```
### 添加一个字到附加库
```bash
go run cmd/main.go -e 的
编码结果: 的‬🈳⁠🈳⁦🈳
go run cmd/main.go -a 的 -p de -r 日 -re 🌞
已添加汉字: 的 读音: t, ɤ, 轻声 部首: 日 ID: 130309308023300
已添加部首: 日 颜文字: 🌞
go run cmd/main.go -e 的                   
编码结果: 🔪⁮😋⁦😯‏🌞
```
### 指定多音字
```bash
go run cmd/main.go -e 你好，世界！看看多音字：行。
编码结果: 🥛​👔⁮🐴‮👤🐒🐱🐎👩，🌴☀️😨🌍➖✌️😨👨‍🌾！👖🔐🍉👁️😭🔐🍉👁️🕳️🌀🍉🪩🧂🎵🍉🎵⬅️☀️😨🚼：[🌟🦅🦎⛕|🌷👍🦎⛕]。
可选形声: [2]
在参数中指定形声字编号(从0开始)以生成不带中括号的编码结果
go run cmd/main.go -e 你好，世界！看看多音字：行。 1
编码结果: 🥛⁦👔⁢🐴⁭👤🐵🐈‍⬛🐎👩，🪸☀️😨🌍🐓✌️😨👨‍🌾！👖🔐🍉👁️😭🔐🍉👁️🪔🌀🍉🪩💊🎵🍉🎵👈🌅😨🚼：🐅👍🧗‍♂️⛕。
go run cmd/main.go -d 🥛⁦👔⁢🐴⁭👤🐵🐈‍⬛🐎👩，🪸☀️😨🌍🐓✌️😨👨‍🌾！👖🔐🍉👁️😭🔐🍉👁️🪔🌀🍉🪩💊🎵🍉🎵👈🌅😨🚼：🐅👍🧗‍♂️⛕。
解码结果: [你|儗]好，世[界|畍]！看看多音字：行。
```

## 实用工具
### 拼音识别拆分
将带声调的拼音拆分为以国际音标表示的声母韵母。
```go
s, y, t, err := emozi.SplitPinyin("jiǒng")
if err != nil {
    panic(err)
}
fmt.Println(s, y, tone) // tɕ i̯ʊŋ 上声
```
