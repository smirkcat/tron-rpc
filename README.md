## 波场钱包代理

1. 采用GPRC协议调用超级节点服务
2. 使用golang开发，目前满足trx、trc10和trc20转账以及交易记录扫描，归集，地址生成功能
3. 每一种合约开启json-rpc端口兼容比特币rpc接口
4. 钱包私钥全部加密保存，密钥保存在sqlite数据库，所以必须定期备份
5. 启动配置文件为当前目录下tron.toml文件，具体示例请查看文件参数

注：clone之后需更新子模块  
git submodule update --init --recursive

**本程序不能直接启动，为了安全，需要自己配置主钱包地址，加密私钥等相关参数**


### 合约发布网站

https://cn.developers.tron.network/docs/issuing-trc20-tokens-tutorial

[*change log*](CHANGELOG.md)