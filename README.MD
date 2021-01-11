# 功能简介
[青云](https://docs.qingcloud.com/product/api/) 简化版本 cli

## 当前支持的API操作
- [DescribeInstances](https://docs.qingcloud.com/product/api/action/instance/describe_instances.html)
- [RunInstances](https://docs.qingcloud.com/product/api/action/instance/run_instances.html)
- [TerminateInstances](https://docs.qingcloud.com/product/api/action/instance/terminate_instances.html)

## 自动补全
支持 bash,zsh,fish,powershell 4种终端
- 子命令自动补全
- 参数静态补全
- 部分参数值可选合法参数补全
- 部分参数值动态补全, TerminateInstances 的--instances参数支持动态补全

# 安装
## Linux 平台
系统先安装 completion-bash 包，安装方法自行Google。
如果已经安装，可略过。 Go版本1.14以上。

```bash
mkdir $HOME/src
cd $HOME/src
git clone https://github.com/hex2tan/qingcloud-cli.git
cd qingcloud-cli
go install
```

*注意：如果系统变量PATH不包含GOPATH，请自行添加。*

## 开启自动补全
执行以下命令后，再重新开启一个新的bash则可。
```bash
qingcloud-cli completion bash > /etc/bash_completion.d/qingcloud-cli
```

## 配置文件
将 $HOME/src/qingcloud-cli/example.yaml 里的 access_key_id及secret_access_key替换成自己的id和key。
```
access_key_id : 'QYACCESSKEYIDEXAMPLE'
secret_access_key : 'SECRETACCESSKEY'
```
[如何获取青云access_key以access_id。](https://docs.qingcloud.com/product/api/common/signature.html#api-%E5%AF%86%E9%92%A5%E7%AD%BE%E5%90%8D)

请放在$HOME目录下，文件名为.qingcloud.yaml
```bash
cd $HOME/src/qingcloud-cli/
cat example.yaml > $HOME/.qingcloud.yaml
```

# 设计相关
- 基于[cobra](https://github.com/spf13/cobra) 库进行开发
- 命令参数的解析与构造使用golang的反射机制实现