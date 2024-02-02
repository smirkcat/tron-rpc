## 波场钱包代理

1. 采用GPRC协议调用超级节点服务
2. 使用golang开发，目前满足trx、trc10和trc20转账以及交易记录扫描，归集，地址生成功能
3. 每一种合约开启json-rpc端口兼容比特币rpc接口
4. 钱包私钥全部加密保存，密钥保存在sqlite数据库，所以必须定期备份
5. 启动配置文件为当前目录下tron.toml文件，具体示例请查看文件参数

注：clone之后需更新子模块  
git submodule update --init --recursive  
**为了安全，需要自己配置主钱包地址，加密私钥等相关参数**

### 合约发布

文档，中文文档已经无法查看，需要管理员登录，请直接查看英文文档
https://developers.tron.network/docs/creating-and-compiling

[*change log*](CHANGELOG.md)


### 生成相关配置 

```shell
# 主私钥
go test -timeout 30s -run ^TestCreatAddress$ tron/trx -v
# === RUN   TestCreatAddress
#     client_test.go:xx: 9f974ecb-839c-483f-9f36-f0e730b51658
#     client_test.go:xx: TL4kyKaXJ9gThBhHtyMSN4ZMKSaD5cZUGL
#     client_test.go:xx: PxfYdiHPZ6SloTlLBTEKmCHbB/YDOtD44KWIm7eZELKhjOu335WTq/eethyym3ks2KW2ZUGLGJbmOOcOMBaBianlmpg2SUbhih9Yk1HKWgk=
# --- PASS: TestCreatAddress (0.00s)
# PASS
# ok      tron/trx        0.093s

# db 文件生成
go test -timeout 30s -run ^TestDB$ tron/trx -v
# === RUN   TestDB
# --- PASS: TestDB (0.06s)
# PASS
# ok      tron/trx        0.180s
```