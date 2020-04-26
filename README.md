## 波场钱包代理

1. 采用GPRC协议调用超级节点服务
2. 使用golang开发，目前满足trx、trc10和trc20转账以及交易记录扫描，归集，地址生成功能
3. 每一种合约开启json-rpc端口兼容现有清算系统
4. 钱包私钥全部加密保存，密钥保存在sqlite数据库，所以必须定期备份
5. 启动配置文件为当前目录下trx.toml文件，具体示例请查看文件参数

git config --global https.proxy http://127.0.0.1:10809
git submodule update --init --recursive


[*change log*](CHANGELOG.md)