# xbench

xbench是xuperchain的压测工具，功能：

* 支持转账、存证和合约的压测
* 支持控制并发数、恒定RPS、阶梯发压、异步发压的能力
* 支持扩展，可以根据压测场景构造数据集

与老版本差异：

* 发压与统计能力分离，xbench专注于发压功能，统计功能使用grafana进行可视化展示
* 基于开源的grpc压测工具ghz进行二次开发，开发重点在根据测试场景构造测试数据集

## 快速开始

1.执行构建

```
make
```

2.准备bank账户

压测执行合约和转账会消耗token，需要确保bank账户(`data/bank`)中有充足的token：

* 创世块配置中为bank账户预分配一笔token，参考`data/genesis/xuper.json`；
* 区块链网络部署好后，为bank账户转入一笔token；

3.执行转账压测

```bash
bin/xbench --config=conf/transfer.yaml
```

> 注意：修改配置文件中压测节点的ip:port
> 更多案例参考：script/xbench.sh

4 执行存证压测
```bash
bin/cbench --config=conf/evidence.yaml
```

> 注意：修改xchain配置为**nofee**模式,
> 设置data/genesis/xuper.json中nofee
> 字段为true

5.压测结果

xbench 执行完后会打印基础指标：响应耗时分位值、tps均值和成功率，更详细的指标通过grafana展示。

#### 可视化展示

```
XuperChain（打点）=> prometheus（采样、存储）=> grafana（可视化）
                       ||
                       ||=> AlertManager（报警）
```

监控指标：

* 机器：cpu、mem、disk、net；
* 服务：接口的吞吐、耗时、错误率，go服务的协程数、文件描述符；
* 业务：XuperChain内部状态指标
    * 账本：上链交易量、账本高度；
    * 状态机：未确认交易量；
    * 网络：网络消息吞吐、耗时、字节量；

部署xchain参考https://github.com/xuperchain/xuperchain ，启动xchain时，配置开启监控：

* 设置conf/env.yaml文件中metricSwitch:true

部署prometheus和grafana服务，官网下载安装包，启动服务时使用下面的配置文件：

* prometheus配置：conf/metric/prometheus.yml
* grafana模板：conf/metric/grafana-xchain.json

如果需要查看机器指标：
* 在运行区块链网络节点的机器上部署prometheus的node_exporter服务
* grafana模板：conf/metric/grafana-node.json

#### 支持的压测用例

* transfer: 转账压测，通过调用sdk生成交易数据
* transaction: 转账压测，离线生成交易数据，没有进行 SelectUTXO
* evidence: 存证压测，离线生成存证数据，存证数据存放在desc字段
* counter: counter合约压测，调用sdk生成数据
* short_content: short_content存证合约压测，调用sdk生成数据
* file: 文件压测，读取文件中的数据进行压测，tx数据是json格式，要求每个并发一个独立的文件

#### 产出包介绍

```txt
bin
|-- generate                    --生成离线交易工具
`-- xbench                      --压测工具
conf
|-- metric                      --监控配置：直接导入prometheus+grafana启动服务
|   |-- grafana-node.json       --机器信息的grafana模板，需要先在机器上部署node_exporter服务
|   |-- grafana-xchain.json     --xchain服务的grafana模板
|   `-- prometheus.yml          --采集指标的prometheus配置
|-- contract
|   |-- counter.yaml            --counter合约压测配置
|   `-- short_content.yaml      --存证合约压测配置
|-- bench.yaml                  --配置指南详解
|-- evidence.yaml               --存证压测配置：desc字段
|-- file.yaml                   --文件压测配置：发压工具从文件读取交易
|-- sdk.yaml
|-- transaction.yaml            --转账交易压测配置：not SelectUTXO
`-- transfer.yaml               --转账交易压测配置
data
|-- account                     --压测账户：5000个账户
|   |-- address.dat             --账户地址
|   `-- mnemonic.dat            --账户助记词，通过助记词可以解析账户
|-- bank                        --银行账户AK
|   |-- address
|   |-- private.key
|   `-- public.key
|-- evidence                    --离线生成存证数据的默认产出路径
`-- transaction                 --离线生成转账数据的默认产出路径
pb
script
|-- build.sh                    --构建工具
`-- xbench.sh                   --xbench用例
```

## 进阶

发压过程：

```
                      grpc接口         测试用例
         ||go0 => DataProvider() => Generate(0) => proto.Message0 => Send =>||
         ||go1 => DataProvider() => Generate(1) => proto.Message1 => Send =>||
xbench =>||go2 => DataProvider() => Generate(2) => proto.Message2 => Send =>||=> reporter
         ||go3 => DataProvider() => Generate(3) => proto.Message3 => Send =>||
         ||go4 => DataProvider() => Generate(4) => proto.Message4 => Send =>||
```

新增grpc接口：实现Provider接口，为ghz发压工具提供压测数据
```
type Provider interface {
	DataProvider(*runner.CallData) ([]*dynamic.Message, error)
}
```

新增测试用例：实现Generator接口，提供生成压测数据功能
```
// 生成压测用例接口
type Generator interface {
	// 业务初始化
	Init() error
	// 根据并发ID获取构造的交易数据
	Generate(id int) (proto.Message, error)
}
```

新增合约：实现Contract接口，构造合约数据
```
type Contract interface {
	Deploy(from *account.Account, name string, code []byte, args map[string]string, opts ...xuper.RequestOption) (*xuper.Transaction, error)
	Invoke(from *account.Account, name, method string, args map[string]string, opts ...xuper.RequestOption) (*xuper.Transaction, error)
	Query(from *account.Account, name, method string, args map[string]string, opts ...xuper.RequestOption) (*xuper.Transaction, error)
}
```

## FAQ

问题1：构建时遇到missing go.sum错误：

* 解决：`go mod download all`

问题2：文件发压报错`read tx file error`：

* 原因：文件中的数据总量 <= 压测配置中的数据总量，每个并发总独立发送，有的先发送完。
* 解决：文件中的数据总量 > 1.1 * 压测配置中的数据总量

## 文档

[ghz Guide](https://ghz.sh/docs/intro)
