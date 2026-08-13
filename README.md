# Honeynet

Honeynet 是面向企业安全场景的蜜网与欺骗防御平台。先安装管理端（Server），管理端自动提供一个内置节点；需要扩展覆盖范围时，再从控制台向 Linux 或 Windows 主机安装节点端（Agent/Client），由节点端运行蜜罐（Pot）。

> [!IMPORTANT]
> 本项目按 [PolyForm Noncommercial License 1.0.0](LICENSE) 提供，仅授权非商业用途。未经版权方另行书面授权，不得用于商业目的。本项目属于“源码可用”软件，不是 OSI 定义的开源软件；仓库内第三方组件和资源继续适用各自的许可证与权利声明。

## 已打包版本

最新完整发行包为 [Honeynet Server v0.24.0（Linux AMD64）](releases/v0.24.0/README.md)，包含 Server、Agent、Web 控制台、规则与蜜罐资源。发行包通过 Git LFS 保存，下载后请按发布目录中的 SHA-256 校验值核验完整性。

Server、Agent 与蜜罐运行链不依赖 Docker。默认且受支持的交付方式是 Go 原生二进制、`honeypot-templates-server` 资源目录和操作系统服务（systemd / Windows Service）；Docker 只承载数据引擎，不运行任何蜜罐。

运行边界固定为两个进程：`honeynet-server` 只负责管理 API、Web 控制台、数据与控制通道，`honeynet-agent` 只负责节点能力、蜜罐监听、蜜饵和事件回传。内置节点同样由独立 Agent 二进制运行，不嵌入 Server。MySQL 保存业务数据，ClickHouse 专门保存攻击事件与安全分析数据；默认 ClickHouse Docker 端口只绑定本机。

## 当前能力

- Gin + GORM + MySQL 管理 API，JWT 登录与 admin/operator/viewer 权限
- MySQL + ClickHouse 双引擎，业务控制面与高吞吐安全事件分析面相互隔离
- React 18 + Arco Design 控制台，覆盖仪表盘、节点、蜜罐、事件、告警、情报、资产和系统管理
- 管理端内置节点，以及节点创建、30 分钟一次性注册令牌、心跳和在线状态
- 独立 TLS 1.3 Agent 网关，CA 指纹固定引导、节点客户端证书、自动续期与序列号吊销
- mTLS WSS 声明式控制通道、断网持久化队列、gzip 批量事件回传与幂等确认
- 36 种 Go 原生协议蜜罐，加上直接来自 `honeypot-templates-server` 的 67 种完整 Web 资源，共 103 种可真实监听的低/中交互蜜罐
- 111 种蜜罐服务目录与节点能力协商：Server 只允许创建目标 Agent 实际支持的服务，不再把目录收录等同于可运行
- 自定义 Web 模板运行时；文件/凭据蜜饵安全投放与访问监控、网络 Token 被动关联
- Linux 386/AMD64/ARM/ARM64/Loong64 与 Windows 386/AMD64 Agent 构建
- Agent 构建 SHA-256 + Ed25519 清单签名；Linux Agent 支持金丝雀/分批灰度发布、失败自动暂停与进程外健康检查自动回滚
- Linux Agent 被动全端口扫描感知：AF_PACKET 捕获 TCP SYN/UDP 探测，按来源与时间窗聚合告警
- 攻击事件实时推送、规则阈值/静默、IOC 自动沉淀、资产聚合和审计日志
- Agent 实时 YARA 兼容初筛、Server 统一规则下发与入库复核；只有中心复核命中才生成检测告警
- IPIP.net 免费版城市库离线定位，Server 统一将攻击源 IP 补全到国家/省/市并回填历史事件
- SMTP、Syslog、通用 Webhook、企业微信、钉钉、飞书真实告警外发，支持持久化重试和投递审计

Agent 的本地事件队列满载时会明确拒绝新写入并记录错误，不会静默删除尚未收到 Server ACK 的旧取证事件。启用 ClickHouse 后，Server 只有在安全事件已持久化且 MySQL 告警/IOC 等业务副作用完成后才确认事件；临时故障会由 Agent 保留并自动重传。Server 明确标记为永久无效的事件会先将完整原始事件、拒绝原因和时间追加并 `fsync` 到 Agent 状态目录下权限为 `0600` 的 `dead-letter-events.jsonl`，确认落盘后才移出主队列，便于人工修复和审计。

## 节点服务能力与协议蜜罐

Agent 在每次 `hello` 和心跳中上报 `pot.<service>` 能力，Server 规范化后持久化到节点记录。进入“蜜罐编排”时必须先选择节点，控制台会展示该节点实际可运行的服务数量，并禁用未支持的目录项；Server 在创建和启动接口再次强制校验，旧客户端尚未上报能力时会要求先升级并上线。

v0.9 首批新增四类协议蜜罐：

- FTP：模拟 vsFTPd 登录与常用命令，产生 `ftp.command`、`ftp.credential`。
- PostgreSQL：支持启动包、SSL 降级、明文认证和简单查询交互，产生 `postgresql.credential`、`postgresql.query`。
- SMTP：支持 EHLO、AUTH PLAIN/LOGIN、邮件信封与受限 DATA 捕获，产生 `smtp.credential`、`smtp.message`。
- DNS：UDP 查询解析与安全的保留地址响应，产生 `dns.query`。

v0.10 第二批新增六类协议蜜罐：

- MSSQL：支持 TDS Prelogin、Login7 凭据解码与 SQL Batch 捕获，产生 `mssql.credential`、`mssql.query`。
- MongoDB：支持 OP_QUERY/OP_MSG、Hello/BuildInfo 响应及 SCRAM 用户识别，产生 `mongodb.command`、`mongodb.authentication`。
- Elasticsearch：模拟集群信息与健康接口，捕获 HTTP Basic 凭据和请求正文，产生 `elasticsearch.request`、`elasticsearch.credential`。
- MQTT：支持 3.1.1/5.0 CONNECT、PUBLISH、SUBSCRIBE 和 PING，产生 `mqtt.connect`、`mqtt.credential`、`mqtt.publish`。
- Modbus TCP：解析 MBAP、读写线圈/寄存器功能码并安全响应，产生 `modbus.request`、`modbus.write`。
- RTSP 摄像头：支持 OPTIONS、DESCRIBE、SETUP、PLAY 与 Basic 认证，产生 `rtsp.request`、`rtsp.credential`。

v0.11 第三批新增六类协议蜜罐：

- SMB：支持 NetBIOS 帧、SMB1 探测、SMB2 Negotiate/Session Setup 和 NTLM Type 3 身份信息捕获，产生 `smb.negotiate`、`smb.authentication`。
- RDP：支持 TPKT/X.224 连接协商与 `mstshash` Cookie 用户识别，产生 `rdp.connection`、`rdp.username`。
- LDAP：支持 BER 编码的 LDAP v2/v3 Simple Bind、Search、Unbind，产生 `ldap.credential`、`ldap.search`。
- SNMP：支持 SNMP v1/v2c Get/GetNext/Set 请求解析与 GetResponse，产生 `snmp.request`、`snmp.community`。
- POP3：支持 USER/PASS、APOP、AUTH PLAIN 和邮箱命令模拟，产生 `pop3.command`、`pop3.credential`。
- IMAP：支持 LOGIN、AUTHENTICATE PLAIN、CAPABILITY、SELECT、FETCH 等交互，产生 `imap.command`、`imap.credential`。

v0.12 增加节点级蜜罐 TLS 能力：

- Agent 在状态目录生成并持久化共享 ECDSA P-256 自签名服务证书，私钥权限为 `0600`；证书有效期 30 天，剩余 7 天自动轮换，服务运行中无需重启。
- HTTPS、SMTPS、IMAPS、POP3S、LDAPS 复用现有协议交互和事件模型，通过 TLS 1.2/1.3 监听；证书管理器不可用时拒绝启动，禁止静默降级为明文。

v0.13 第五批新增六类协议蜜罐：

- TFTP：支持 UDP RRQ/WRQ、DATA/ACK 交互和文件名采集，产生 `tftp.request`、`tftp.data`。
- VNC：支持 RFB 3.8 协商和 VNC Authentication 挑战响应采集，产生 `vnc.connection`、`vnc.authentication`。
- Memcached：支持 ASCII `version/stats/get/set/add/replace/delete` 等命令和受限数据项模拟，产生 `memcached.command`、`memcached.item`。
- Oracle TNS：解析 CONNECT 包与连接描述符，提取 `SERVICE_NAME/USER/PROGRAM` 并返回 TNS REFUSE，产生 `oracle.connect`。
- ZooKeeper：支持 `ruok/stat/envi/conf/srvr/mntr` 四字命令和基础连接握手，产生 `zookeeper.command`、`zookeeper.connect`。
- Kafka：解析长度帧、API Key/Version、Correlation ID、Client ID，响应 ApiVersions/Metadata，产生 `kafka.request`。

v0.14 工控/IoT 协议阶段新增三类蜜罐：

- Siemens S7comm：支持 ISO-on-TCP TPKT、COTP Connection、Setup Communication、Read Var/Write Var，提取 TSAP、DB、存储区和地址，产生 `s7.connection`、`s7.request`、`s7.read`、`s7.write`。
- CoAP：支持 Confirmable/Non-confirmable GET/POST/PUT/DELETE、Token、URI Path/Query 和 Payload 解析，产生 `coap.request`、`coap.write`。
- BACnet/IP：支持 BVLC/NPDU、Who-Is/I-Am 设备发现以及 ReadProperty/WriteProperty 识别，产生 `bacnet.request`、`bacnet.discovery`、`bacnet.read_property`、`bacnet.write_property`。

## Web 蜜罐资源运行时

内置 Web 蜜罐不再使用项目生成的 `profile_pages_*` 页面。Agent 直接读取 `honeypot-templates-server/services/config.json`，以每条配置的 `root`、`index`、`https` 和原始文件作为唯一运行依据：

- 当前资源包配置了 68 条静态 Web 服务，其中 67 条目录和入口完整，可原生启动；`router-cmcc` 缺少资源目录，因此 Agent 不会上报该能力，也不会用其他页面冒充。
- 支持普通文件、`URL 查询参数 → 文件名中的全角 ？`、`METHOD__接口名/index.html` 三种资源寻址，并应用同名 `.res` 中的状态码和安全响应头。
- 正文始终使用资源包中的原始字节。已解压资源不会错误继承抓包时的 `Content-Encoding`/`Content-Length`，绝不执行包内脚本或下载文件。
- 每次访问产生带 `template_code` 和 `template_source=honeypot-templates-server` 的 `web.request`；表单、JSON、查询参数或 Basic Auth 中出现凭据时产生 `web.credential`。
- 资源包中的静态 Web 配置均启用 HTTPS，由 Agent 的节点蜜罐证书提供 TLS。实例监听端口仍以 Server 编排值为准，不强占配置文件中的历史端口。

旧的 62 种手写画像和相关探测响应已从 Agent 工厂及新安装服务目录移除。资源包没有提供的产品（例如此前手写的金蝶页面）不会继续宣称兼容；后续要增加此类产品，必须先以相同目录协议补充真实资源。

测试门禁会校验 111 条新安装服务目录、67 条可用 Web 资源和 103 项 Agent 运行能力，并原生逐个启动全部 67 条 Web 服务。Server/Agent 发布包和远程节点安装脚本会一并分发、校验并安装该资源目录。

服务目录仍用于规划后续协议覆盖；未出现在节点能力集中的服务不会被宣传为已实现，也不能下发运行。

## 原生安装 Server（推荐）

当前管理端安装包支持 Linux AMD64/ARM64 与 Windows x86/x64。Linux 要求 systemd；所有平台都需要一个可用的 MySQL 8 实例。先在 MySQL 中创建数据库和最小权限账号，例如：

```sql
CREATE DATABASE honeynet CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
CREATE USER 'honeynet'@'127.0.0.1' IDENTIFIED BY 'replace-with-a-strong-password';
GRANT ALL PRIVILEGES ON honeynet.* TO 'honeynet'@'127.0.0.1';
```

MySQL 始终是用户、节点、蜜罐、规则、告警处置和系统配置的业务数据库，不会被 ClickHouse 替换。Linux Server 安装包默认同时部署一个独立的 ClickHouse 安全数据分析引擎，要求主机已安装并启动 Docker Engine 与 Compose v2；如使用外部 ClickHouse 或暂不启用分析引擎，可给安装命令增加 `--skip-clickhouse`。

从源码构建管理端发布包：

```bash
make release-server VERSION=0.24.0 TARGET_OS=linux TARGET_ARCH=amd64
tar -xzf dist/honeynet-server-0.24.0-linux-amd64.tar.gz
cd dist/honeynet-server-0.24.0-linux-amd64
sudo ./scripts/install-server.sh \
  --mysql-dsn 'honeynet:replace-with-a-strong-password@tcp(127.0.0.1:3306)/honeynet?charset=utf8mb4&parseTime=True&loc=Local' \
  --public-url 'http://192.0.2.10:8080' \
  --agent-public-url 'https://192.0.2.10:8443'
```

默认会额外安装一个独立的本机 `honeynet-agent.service` 作为内置节点；若管理端主机只运行 Server，可在命令末尾增加 `--server-only`。这不会影响之后从控制台向其他主机安装节点。

需要最小化管理端包时可执行 `make release-server-only ...`。生成的 `honeynet-server-only-*` 只包含 Server 二进制、Web 控制台和 Server 系统服务，安装器会自动进入 Server-only 模式，不包含或启动 Agent。

安装程序会完成以下工作：

- 安装 Server、Agent、前端、各平台 Agent 下载文件及 Web 蜜罐资源到 `/opt/honeynet`
- 首次安装时生成 JWT 密钥、内置节点 Token 和随机管理员密码
- 写入 `/etc/honeynet/server.yaml`，现有配置在升级时会保留
- 在 `/var/lib/honeynet/pki` 生成并持久化内部节点 CA，私钥权限为 `0600`
- 注册并启动 `honeynet-server.service` 与同机内置节点 `honeynet-agent.service`
- 自动迁移数据库并创建五个默认蜜罐
- 默认启动固定兼容版本的单节点 ClickHouse，只将 HTTP `8123` 和 Native `9000` 绑定到 `127.0.0.1`
- 为 ClickHouse 创建持久卷、版本化 Schema 和无 DDL 权限的 Server 应用账户；凭据保存到 `/etc/honeynet/clickhouse.env`，分析配置保存到 `/etc/honeynet/analytics.yaml`
- 新节点 mTLS 证书默认有效期 400 天、到期前 30 天自动续期；已有证书按自身到期时间平滑轮换

安装完成后打开 `--public-url` 指定的地址。Console 的 `8080` 默认使用 HTTP；节点一键安装在 HTTP 和 HTTPS Console 下都可生成，下载后仍会核对 Agent、稳定守护器和模板包的 SHA-256。若需要加密管理登录与下载链路，可显式传入 `--console-tls`，或由可信反向代理完成公网 TLS。`--console-tls` 会把 `server.tls_enabled` 写为 `true`，控制台使用 Honeynet CA 签发的 Server 证书并强制 TLS 1.3，但不会要求浏览器提供节点客户端证书；Agent 的身份控制通道始终由独立的 `8443` TLS 1.3/mTLS 网关保护。首次访问私有 CA HTTPS 前需将 `/var/lib/honeynet/pki/ca.crt` 安全导入管理员浏览器或操作系统信任库，绝不能复制或暴露同目录中的任何私钥。

Console 也可直接使用域名或公网 IP 对应的外部可信证书。在 `server` 段成对配置 `tls_cert_file` 与 `tls_key_file`（或环境变量 `HONEYPOT_TLS_CERT_FILE`、`HONEYPOT_TLS_KEY_FILE`），并保持 `tls_enabled: true` 与 HTTPS `public_url`。此模式只替换 Console 证书；`8443` Agent 网关仍使用 Honeynet 节点 CA 和强制 mTLS，节点安装命令则通过操作系统信任链下载，不会嵌入节点 CA 作为 Console 信任根。Server 以 `honeynet` 低权限用户运行，因此证书链必须可读、私钥只应授权 `root:honeynet` 读取；证书续期后需安全地更新文件并重启 `honeynet-server`。两个文件必须同时配置且内容必须匹配，否则 Server 会拒绝启动。

若未传管理员密码，安装程序只在首次安装结束时显示一次随机密码。状态和日志：

```bash
systemctl status honeynet-server honeynet-agent
journalctl -u honeynet-server -u honeynet-agent -f
```

Linux 安装器在覆盖原生 Server/Agent 前会先完成 ClickHouse 迁移与读写冒烟，并保存旧二进制、静态资源、配置、PKI 和 systemd 状态；新版管理控制台和 Agent mTLS 网关必须在 120 秒内同时通过启动探测，否则自动恢复旧版本及原运行/启用状态。`--no-start` 只做安装与验证，不会把原本停止的 ClickHouse 留在运行状态。

> `public_url` 与 `agent_public_url` 必须是其他节点能够访问的地址，分布式部署中不能填写 `localhost`；防火墙需开放管理端口 `8080` 和 Agent mTLS 端口 `8443`。HTTP Console 便于直接部署，但管理登录和安装包下载链路不加密；生产环境建议在条件具备后启用可信 HTTPS。使用私有 Honeynet CA 的新节点还需先安全取得并信任 `ca.crt`，才能校验控制台 HTTPS 下载；禁止用 `curl -k` 绕过校验。内置 Agent 需要监听 22/23/80 等低端口，因此其服务以 root 运行；Server 自身以独立的 `honeynet` 低权限用户运行。

Server、Agent 网关、节点地址上报、攻击事件和被动扫描感知均支持 IPv6。监听地址可配置为 `[::]:8080`、`[::]:8443`，IPv6 URL 必须使用标准方括号形式，例如 `https://[2001:db8::10]:8080` 和 `https://[2001:db8::10]:8443`。节点会同时上报 `public_ips`、`private_ips` 双栈候选并保留兼容字段 `public_ip`；蜜罐实例将 `config.bind` 设置为 `::` 或具体 IPv6 地址即可监听 IPv6，原有 IPv4 配置保持兼容。

## ClickHouse 安全数据分析引擎

默认部署使用官方 `clickhouse/clickhouse-server:25.8.28.1` 固定镜像和独立 Compose 项目 `honeynet-analytics`，不会启动、停止或修改 `deploy/mysql/compose.yaml` 中的 MySQL。ClickHouse 保存攻击事件、扫描、凭据观测和统计聚合；业务数据仍写 MySQL，Agent 也始终只连接 Honeynet Server。

从源码单独安装、迁移和验收：

```bash
sudo ./scripts/install-clickhouse.sh
sudo ./scripts/migrate-clickhouse.sh
sudo ./scripts/smoke-clickhouse.sh
```

安装器首次运行生成随机迁移、应用、写入和查询账户密码，升级时不会覆盖已有 `/etc/honeynet/clickhouse.env` 或 `/etc/honeynet/analytics.yaml`。迁移严格按 `migrations/clickhouse/*.sql` 的文件名顺序执行，并在 `schema_migrations` 中记录版本；Server 使用 `honeynet_app`，仅具备事件表 INSERT/SELECT、聚合视图 SELECT 和迁移台账只读权限，不具备 DDL 权限。Server 启动和 `/healthz` 不只测试端口连通性，还会核对迁移版本、表引擎以及全部取证字段的名称与类型，Schema 缺失或漂移时拒绝启动或标记为不健康。

状态与排障：

```bash
docker compose --env-file /etc/honeynet/clickhouse.env -f /opt/honeynet/deploy/clickhouse/compose.yaml ps
docker compose --env-file /etc/honeynet/clickhouse.env -f /opt/honeynet/deploy/clickhouse/compose.yaml logs -f clickhouse
curl --fail http://127.0.0.1:8123/ping
```

停止并卸载运行实例：

```bash
sudo /opt/honeynet/scripts/uninstall-clickhouse.sh
```

卸载脚本只移除 ClickHouse 容器和独立项目网络，不使用 `--volumes`，因此 `honeynet_analytics_clickhouse_data` 数据卷、凭据和分析配置都会保留。确需永久销毁数据时必须由管理员在确认备份后单独删除该明确卷名。

Windows Server 不在本机安装 ClickHouse；请在 Linux 管理节点上运行上述默认 Docker 部署，或配置一个仅能从 Server 管理网访问的外部 ClickHouse。禁止将 `8123`、`9000` 直接开放到公网。

Windows 管理端发布包与安装方式：

```bash
make release-server VERSION=0.24.0 TARGET_OS=windows TARGET_ARCH=amd64
```

将 `dist/honeynet-server-0.24.0-windows-amd64.zip` 复制到 Windows 并解压，然后在管理员 PowerShell 中执行：

```powershell
.\scripts\install-server.ps1 `
  -DatabaseDSN 'honeynet:replace-with-a-strong-password@tcp(127.0.0.1:3306)/honeynet?charset=utf8mb4&parseTime=True&loc=Local' `
  -ConsoleTLS `
  -PublicURL 'https://192.0.2.20:8080' `
  -AgentPublicURL 'https://192.0.2.20:8443'
```

Windows 安装到 `%ProgramFiles%\Honeynet`，配置保存在 `%ProgramData%\Honeynet\server.yaml`，内部 CA 位于 `%ProgramData%\Honeynet\pki\ca.crt`，Server 与内置 Agent 均注册为支持失败自动重启的 Windows Service。启用 `-ConsoleTLS` 时安装器会使用该 CA 验证本机 TLS 1.3 `/healthz`，管理员仍需按组织策略把 CA 导入浏览器信任库。

## 部署新节点

进入“节点管理”创建节点，复制控制台生成的一次性安装命令并在目标主机执行。命令包含 30 分钟注册令牌和管理端 CA 的 SHA-256 指纹：Agent 在 TLS 1.3 注册连接中固定该指纹，取得独立客户端证书后立即清除令牌。Linux 安装脚本会识别 CPU 架构，下载 Agent 和 `honeypot-templates-server.tar.gz`，分别验证 SHA-256 后原生解压并注册 systemd 服务；Windows 使用对应 ZIP 资源包并注册支持失败恢复的 Windows Service。资源目录记录在 Agent 配置的 `template_root`，运行蜜罐不需要容器。

Agent 注册和心跳会同时上报全部有效本机网卡地址，Server 还会从 8443 连接来源记录 NAT/公网出口地址。“节点管理 → 地址”可选择自动（公网优先）、公网、任一私网或自定义 IP；选中的 `ip` 同时用于服务访问链接。公网、私网和自定义模式属于手动选择，后续心跳只更新候选地址，不会覆盖当前选择。

源码运行 Server 前执行 `make agents VERSION=0.24.0`，即可生成控制台一键安装所需的全部平台 Agent 二进制与模板下载包；正式 Server 发布包会自动携带这些文件。

也可以构建完全独立、无需在线下载安装器的 Agent 包：

```bash
make release-agent VERSION=0.24.0 TARGET_OS=linux TARGET_ARCH=amd64
tar -xzf dist/honeynet-agent-0.24.0-linux-amd64.tar.gz
sudo ./honeynet-agent-0.24.0-linux-amd64/scripts/install-agent.sh \
  --server 'http://192.0.2.10:8080' \
  --agent-url 'https://192.0.2.10:8443' \
  --ca-sha256 '<控制台显示的 CA SHA-256>' \
  --node-id '<节点 ID>' \
  --token '<30 分钟一次性令牌>'
```

Windows 独立包内提供 `scripts/install-agent.ps1`。两个安装器只安装一个 Agent 二进制、Web 蜜罐资源目录和系统服务，不安装数据库、Web 管理端或任何容器运行时。

执行 `make release-agents VERSION=0.24.0` 可一次生成 Linux 386/AMD64/ARM/ARM64/Loong64 与 Windows 386/AMD64 的全部独立节点包。

节点证书默认有效期 400 天，剩余 30 天时自动续期。重新签发安装令牌或删除节点会清除当前证书序列号并主动断开控制连接。CA、证书有效期和续期窗口可在 `server.yaml` 的 `agent` 段配置。

## 自定义 Web 蜜罐

在“自定义 Web 蜜罐”中用受限 YAML 定义精确路由、HTTP 方法、静态响应和需要捕获的请求字段，然后在“蜜罐编排”中选择“自定义 Web 模板”服务和目标模板完成部署。模板不执行脚本、文件引用或任意用户代码；Server 会拒绝未知字段、不安全响应头、重复路由、超限正文和非 `web.*` 事件类型。

```yaml
name: fake-oa-portal
listen:
  port: 8080
pages:
  - path: /login
    method: GET
    response:
      status: 200
      body: "Honeynet login portal"
  - path: /login
    method: POST
    capture:
      fields: [username, password]
      event_type: web.credential
    response:
      status: 401
      body: "Invalid username or password"
```

YAML 中的 `listen.port` 是模板的默认端口说明，实例最终监听端口以蜜罐编排配置为准。每次保存模板版本号自动递增，Server 会重新下发所有引用实例，Agent 检测到版本变化后原地重启对应服务。每次请求产生 `web.request` 事件；配置字段命中后产生指定的 `web.*` 事件，并附带模板 ID、名称和版本。仍被实例引用的模板禁止删除。

自定义 Web 模板由原生 Agent 直接监听实例配置的端口，不经过容器端口映射。

Server 是蜜罐实例的唯一控制面：`GET/POST/PUT/DELETE /api/v1/pots` 维护节点、服务、名称、监听端口、配置与目标状态，启动/停止接口改变期望状态；在线 Agent 通过 mTLS WebSocket 接收完整目标态，在本机创建、迁移或关闭真实 TCP/UDP 监听器。节点离线时操作保持为待同步，重连后自动收敛。可在已运行原生 Server/Agent 的开发环境执行 `make smoke-pot-control`，验证创建监听、查询、修改端口、停止、重新启动及删除释放端口的完整链路。

## 蜜饵投放与命中检测

在“蜜饵管理”选择节点和类型即可创建蜜饵，Server 经严格配置校验后通过 `cmd.decoy.apply` 下发完整目标态。页面展示等待同步、监控中、被动关联、已停止或部署失败，并记录命中次数、最近命中时间、实际文件路径和错误原因。

- 文件蜜饵与凭据蜜饵必须配置非根绝对路径。Agent 默认拒绝已有路径，绝不覆盖生产文件；只有显式开启“仅监控已有文件”时才会监控它，并且不会修改或删除。
- Agent 创建的文件及其 SHA-256 所有权清单持久化在 Agent 状态目录。停用或删除蜜饵时只移除内容未变化的 Agent 自有文件；被修改的文件会保留为取证证据。
- Linux 使用 inotify 捕获打开、读取、修改、属性变更、移动和删除，产生 `decoy.file` 或 `decoy.credential` 事件。Windows 和其他平台采用内容轮询降级，可检测修改、删除和重新创建，但不宣称能够检测纯读取。
- 网络蜜饵生成唯一 Token，用户可将其放入虚假 URL、连接串或文档。当同一节点的蜜罐攻击事件载荷包含该 Token 时，Server 自动关联并产生 `decoy.network` 事件。
- 所有 `decoy.*` 事件进入统一事件、WebSocket、审计和默认 critical 告警链路；重复事件上报不会重复累计命中。

原生部署应选择业务环境中不与真实文件冲突的蜜饵路径；本机测试可使用 Agent 状态目录下的独立 `decoys` 子目录。

## 全端口扫描感知

在“节点管理 → 感知配置”中可以为 Linux 节点启用被动扫描感知。Agent 使用 AF_PACKET 观察到达节点网卡的 TCP 初始 SYN 与 UDP 数据包，按“来源 IP + 协议”在时间窗内累计不同目标端口；达到阈值后生成统一的 `port.scan` 事件，继续进入默认低危告警、IOC、实时 WebSocket 和查询链路。

- 功能默认关闭，不监听新端口，也不写入 iptables/nftables；只处理到达本机或镜像网卡的流量。
- 可选择网卡、TCP/UDP 协议、不同端口阈值、统计窗口、告警冷却时间、排除端口和可信扫描器 IPv4/IPv6 CIDR；IPv6 扩展头和首分片会被正确解析，非首分片不会在未重组时误判。
- Linux Agent 需要 root 或进程有效的 `CAP_NET_RAW`；原生 systemd Agent 以 root 运行。Server 始终使用独立低权限用户。
- 配置会同时持久化在 Server 和 Agent。节点离线时可先保存，重连后自动下发；页面展示待同步、运行中、错误、平台不支持及本次感知运行的检测次数。
- Windows 构建保持可用，但本版本尚未集成 Npcap，启用后会明确回报“平台不支持”。

本地调试 Agent：

```bash
go run ./cmd/agent \
  --config ./tmp/agent.json \
  --server http://127.0.0.1:8080 \
  --agent-url https://127.0.0.1:8443 \
  --ca-sha256 '<控制台显示的 CA SHA-256>' \
  --node-id '<控制台生成的节点 ID>' \
  --registration-token '<一次性注册令牌>' \
  --template-root ./honeypot-templates-server/services
```

## Agent 签名升级与灰度发布

管理员进入“Agent 发布”，先执行“扫描并签名当前构建”。Server 会为 `downloads` 中每个 OS/架构构建计算 SHA-256，并使用持久化 Ed25519 私钥对“版本、平台、摘要、大小”组成的规范清单签名。升级私钥首次启动时生成在 `pki/update-signing.key`（原生安装默认 `/var/lib/honeynet/pki/update-signing.key`），权限为 `0600`；公钥和 Key ID 经已经认证的 mTLS 注册/控制通道下发给 Agent。

扫描成功后，Server 会把本次二进制保存为带版本和摘要的只读快照（原生安装默认位于 `/var/lib/honeynet/pki/agent-releases`），后续下载不再引用可能被下一次安装覆盖的通用 `downloads` 文件。已有发布任务历史的版本禁止重新签名，必须使用新版本号，从而保证历史任务记录、签名清单和实际下载制品始终一致；备份 Server PKI 目录时应一并保留这些快照。

新建灰度发布时可选择目标节点、节点组或将二者组合，Server 会对目标去重并优先把在线节点放入金丝雀波次；随后可配置金丝雀数量、后续批次大小和波次观察时间：

- Agent 先离线验证签名描述符，再通过自己的 mTLS 网关地址下载二进制，随后校验长度与 SHA-256；管理端 URL 不参与签名，可适配各节点不同的可达地址。
- Linux 在专用可写 `bin` 目录暂存并原子替换；只读 `libexec` 中的稳定守护器监督每次启动，目标无法执行、提前退出、卡死或未完成健康确认时都能在 Agent 进程外恢复旧二进制。
- Linux 新版本必须在两分钟内通过 mTLS 建立控制连接并收到 `hello.ack`。版本不符、连续启动失败或健康窗口超时会恢复旧二进制；控制台会区分“升级失败”“已自动回滚”和“回滚失败”，并保留节点、版本变化、尝试次数、完成时间与失败原因。
- Windows Agent 构建仍会签名并提供离线安装，但 0.22 不允许加入远程灰度任务；Windows 稳定服务监督器完成并经过真实 Windows Service E2E 前，不宣称启动失败自动回滚。
- 当前波次全部健康并经过观察时间后才推进下一波。任一任务失败时后续波次不会下发；管理员排除问题后可“继续”重试失败任务，或取消剩余任务。

升级签名私钥必须随 Server 数据目录备份。若主动轮换该密钥，应先确保现有 Agent 在线接收新公钥，再签名新版本。首次引入在线升级能力、仍从 `/usr/local/bin` 启动的旧 Linux Agent，需要用 0.22 安装包一次性迁移到 `/opt/honeynet-agent/bin` 和稳定守护器单元；迁移完成后的后续 Linux 版本才能完全远程升级。

## 告警外发

管理员可在“系统管理 → 告警通道”配置 SMTP、Syslog、通用 Webhook、企业微信、钉钉或飞书，并直接发送测试告警。规则的“投递通道”为空时会发送到全部启用通道，也可以为每条规则指定通道。

- 通用 Webhook 发送版本化 `honeynet.alert` JSON；配置密钥后增加 `X-Honeynet-Timestamp` 与 HMAC-SHA256 签名头
- 邮件支持 STARTTLS、隐式 TLS 和无 TLS 三种模式；Syslog 支持 UDP、TCP 八位字节计数分帧及 TLS
- 企业微信、钉钉和飞书使用群机器人消息格式，钉钉与飞书可配置机器人签名密钥
- Webhook URL、机器人密钥、SMTP 密码及自定义请求头不会通过管理 API 明文返回
- 外发记录写入 MySQL，失败后按 5 秒、30 秒、2 分钟重试，最多四次；Server 重启后会恢复未完成任务

最近投递结果可在系统管理页面查看，失败记录可手动重新入队。

## IPIP 城市定位

Server 支持使用 IPIP.net 免费版 `.ipdb` 城市库离线补全攻击来源，不调用外部定位接口。数据库只需安装在管理端，Agent 无需携带：

```yaml
geoip:
  ipip_db_path: "/var/lib/honeynet/ipip.ipdb"
  language: "CN"
```

也可使用 `HONEYPOT_IPIP_DB_PATH` 和 `HONEYPOT_IPIP_LANGUAGE` 环境变量。启用后，新事件入库前由 Server 覆盖节点提供的地理字段，并输出“国家 / 省 / 市”；私网、回环和链路本地地址统一显示为“内网地址”。Server 启动时会异步补全历史记录中仍为空的归属地，不阻塞 API 和 Agent 连接。

源码开发时可把数据库保存为 `data/ipip.ipdb`。原生安装可向安装程序传入现有数据库；Linux 使用 `--ipip-db /path/to/ipip.ipdb`，Windows 使用 `-IPIPDBPath C:\path\to\ipip.ipdb`。数据库无法加载或不包含指定语言时，配置了该功能的 Server 会拒绝启动并给出明确错误，避免静默产生错误位置。

## 离线 IPv4 / IPv6 威胁情报

Server 可加载免费社区 IP 威胁情报数据库，为新入库事件添加“情报等级 / 网络类型 / 恶意行为”中文标签。事件列表、详情与实时推送也会在响应阶段查询当前离线库，所以数据库启用后，历史事件无需重写即可立即展示情报命中结果。情报库保存在 Server 文件系统，不会把大体积数据写入 MySQL，也不会把攻击 IP 发送到外部服务。

当前原生 Go 读取器与发布方 `intelligence-db 1.0.2`（MIT）格式兼容：逐条验证 AES-256-GCM 认证标签并解析 IPv4 / IPv6 索引。下载包会经过 HTTPS、压缩包路径/大小、ZIP 密码、CRC、记录格式和 GCM 完整性检查；只有完整加载成功才原子替换当前数据库，失败时继续使用上一版。

```yaml
threat_intelligence:
  enabled: true
  database_path: "/var/lib/honeynet/threat-intelligence/intelligence.db"
  download_url: "https://intelligence-0.rivers.chaitin.cn/api/share/download/下载标识/chaitin_threatdb_ipv4_ipv6.zip"
  update_interval: "24h"
```

解压密码属于部署密钥，不写入 `server.yaml`。Linux 原生安装使用 systemd 的 `/etc/honeynet/threat-intelligence.env`：

```sh
install -m 0640 -o root -g honeynet /dev/null /etc/honeynet/threat-intelligence.env
# 使用编辑器写入下一行；不要把真实值放入 shell 历史、源码或工单：
# HONEYPOT_THREAT_INTEL_ARCHIVE_PASSWORD=部署方获得的解压密码
systemctl restart honeynet-server
```

也可在首次安装前设置 `HONEYPOT_THREAT_INTEL_ENABLED`、`HONEYPOT_THREAT_INTEL_DB_PATH`、`HONEYPOT_THREAT_INTEL_DOWNLOAD_URL`、`HONEYPOT_THREAT_INTEL_UPDATE_INTERVAL` 和 `HONEYPOT_THREAT_INTEL_ARCHIVE_PASSWORD`；安装器会把密码单独保存到仅 `root:honeynet` 可读的环境文件。Windows 应把密码写入机器级环境变量，再重启 `HoneynetServer` 服务。

控制台“威胁实体 → 攻击来源”显示数据库状态、IPv4/IPv6 手动查询和管理员手动更新入口。自动更新失败不会中断事件入库；状态接口只返回错误摘要，不返回下载地址或解压密码。

## 双阶段攻击检测规则

Server 启动时读取 `detection.rules_dir` 下的 Yara 规则，将每个规则转换为 Honeynet 的受限跨端格式。当前内置规则库共有 54 个规则块：可等价转换的规则默认启用，带复杂位置/布尔表达式而无法安全等价转换的规则会完整保留并标记“待审核”，不会静默降级。响应模板文件不参与检测，也不会被 Server 或 Agent 执行。

```yaml
detection:
  rules_dir: "/opt/honeynet/rules/builtin"
```

管理员可在“威胁感知 → 检测规则”中新增、修改、停用规则，并分别控制 Agent 本地匹配和 Server 二次确认。规则修改后通过已有 mTLS WebSocket 控制通道实时推送给全部在线 Agent；Agent 先在本地匹配并随事件回传命中证据，Server 使用当前统一规则重新匹配并生成最终告警。自定义正则使用 Go RE2，不允许脚本、外部命令或任意 Yara 扩展。

规则修订号使用纳秒级 `int64`。管理 API 额外提供十进制字符串 `revision_text`，控制台只以该字符串作为权威版本，避免 JavaScript 数值精度丢失；页面以短版本展示，悬浮可查看、复制完整版本，并列出每个节点的当前版本、Agent 目标版本、同步状态、时间和错误。Agent 与 Server 可按职责启用不同规则集，因此页面会把“中心/边缘规则集差异”与“节点规则滞后”分开呈现。

HTTP 蜜罐事件把原始请求快照写入独立的 `raw_packet` 字段；二进制正文同时以 Base64 原样保留。控制台详情按“检测结论 → 原始请求包 → 归一化字段”展示，不再让运营人员阅读 JSON。

事件证据默认不脱敏，MySQL、ClickHouse 的列表/详情以及控制台 WebSocket 推送都会直接返回原始请求包、正文和凭据，无需附加 `include_sensitive`。如部署方需要默认隐藏敏感证据，可显式启用 Server 策略：

```yaml
security:
  redact_sensitive_events: true
```

也可以设置 `HONEYPOT_REDACT_SENSITIVE_EVENTS=true`。启用后，列表、详情和实时推送会统一返回掩码/长度/SHA-256 摘要；`admin` 与 `operator` 可在列表或单条事件请求中使用 `include_sensitive=true` 显式查看原文，每次显式查看都会写入 `READ` 审计。响应中的 `sensitive_redaction_enabled` 表示策略是否启用，`evidence_redacted` 表示本次证据是否已脱敏，`sensitive_reveal_audited` 只在本次响应来自已审计的显式查看时为 `true`。

## 平台管理与一键 OEM

管理员可在“平台管理 → 配置管理”集中维护系统名称、系统版本、版权信息、系统 Logo、公司 Logo、客服热线、客服邮箱、官网地址和产品文档地址。保存后登录页、控制台导航与页脚、仪表盘、攻击态势大屏、浏览器标题和站点图标会统一更新；其他角色只能读取公开品牌信息，不能进入配置页面或修改配置。

OEM 配置作为带修订号的单例业务数据保存在 MySQL。文本保存、Logo 上传和恢复内置 Logo 都使用乐观并发控制，多位管理员同时编辑时不会静默覆盖。Logo 仅接受不超过 2 MiB 的 PNG、JPEG 或 WebP；Server 会完整解码、限制尺寸与像素数并重新编码，拒绝 SVG、伪装文件和像素炸弹。审计日志只记录品牌文本、图片类型、字节数和 SHA-256，不记录图片二进制。

公开接口 `/api/v1/platform/branding` 和带不可变 ETag 的 `/api/v1/platform/assets/{system|company}` 供登录页在认证前加载品牌。管理员接口 `/api/v1/platform/settings` 及同一路径下的图片上传/删除操作需要 `admin` 权限；版本冲突返回 `409` 并要求客户端刷新后重试。

## AI 威胁分析模块

AI 能力位于独立的 `internal/ai` 模块，通过 OpenAI Chat Completions 兼容接口接入 DeepSeek、GLM 或其他模型。配置示例：

```yaml
ai:
  enabled: true
  provider: "deepseek"
  base_url: "https://your-provider.example/v1"
  api_key: "replace-me"
  model: "your-model"
  timeout: "45s"
  send_raw_packet: false
```

也可使用 `HONEYPOT_AI_ENABLED`、`HONEYPOT_AI_PROVIDER`、`HONEYPOT_AI_BASE_URL`、`HONEYPOT_AI_API_KEY`、`HONEYPOT_AI_MODEL`、`HONEYPOT_AI_TIMEOUT` 和 `HONEYPOT_AI_SEND_RAW_PACKET` 环境变量作为首次启动配置。管理员随后可在“威胁感知 → AI 分析 → 在线配置”中修改并测试连接，无需重启 Server。API Key 使用由 Server 密钥派生的 AES-256-GCM 密钥加密后写入 MySQL，查询接口只返回“是否已配置”，永不回显密钥。默认只向模型发送归一化事件证据与规则检测结论，`send_raw_packet` 明确打开后才会附加原始请求包。

“威胁感知 → AI 分析”提供单事件自动分析和按来源 IP 聚合的攻击者画像。模型输入被明确标记为不可信攻击证据，结果持久化到 MySQL，并保留提供商、模型、提示证据哈希和失败原因，便于审计。

AI Agent 采用可审计的 Harness Engineering 闭环：管理员或运营人员先从真实事件中标注恶意/正常样本并选择待优化规则，Server 对样本做确定性分层切分；模型只获得脱敏后的训练集和只读规则/反馈工具，隐藏评估集不会发送给模型。候选规则必须通过受限语法静态校验及隐藏集回放，并达到精准率不低于 80%、召回率不低于 60%、误报率不高于 10% 才进入人工审批。审批与发布是两个独立动作，模型没有线上写权限；发布后复用统一规则下发链路，规则版本血缘、执行轨迹、人工反馈和回滚原因都会持久化。误报、正确命中和漏报反馈可成为下一轮优化证据，但永远作为不可信输入处理。

## 源码开发

前端、Server 和 Agent 都直接在宿主机运行；MySQL 可选择使用容器，ClickHouse 默认使用独立容器：

```bash
docker compose -f deploy/mysql/compose.yaml up -d
sudo ./scripts/install-clickhouse.sh
```

已有 MySQL 8 时不需要 Docker，直接配置 DSN：

```bash
cp .env.example .env
set -a && source .env && set +a
make dev-server
```

另一个终端启动前端：

```bash
make dev-web
```

开发控制台位于 <http://localhost:5173>，Vite 会代理 API 与 WebSocket。也可以先执行 `npm run build`，再由 Server 在 <http://localhost:8080> 同端口托管控制台。

Server 支持 YAML 配置并允许环境变量覆盖：

```bash
go run ./cmd/server --config ./configs/server.example.yaml
```

## 原生蜜罐访问

内置 Web 服务不会仅因加入目录而自动监听，必须先在“蜜罐编排”中为在线节点创建并启动实例。本机内置节点会为 Web 资源优先选择 `20000–20099` 的空闲端口；启动后直接访问 `https://127.0.0.1:<实例端口>/`。节点证书为 Agent 本地生成的蜜罐证书，浏览器首次访问需要确认自签名证书警告。这里没有 Docker 端口映射层。

## 工程结构

```text
cmd/server/          Server 入口
cmd/agent/           Agent 入口
configs/             Server YAML 配置样例
deploy/systemd/      原生 Linux 服务单元
scripts/             原生安装与发布包构建
internal/platformservice/  Linux 信号与 Windows SCM 生命周期适配
internal/agent/      控制客户端、事件队列、运行时与协议蜜罐
internal/agent/sense/ Linux AF_PACKET 被动扫描感知、聚合检测与跨平台降级
internal/agentupdate/ Agent Ed25519 签名清单、自替换、健康确认与 Linux 进程外回滚
internal/alerting/   告警规则投递队列及六类外发适配器
internal/geoip/      IPIP 城市库加载、地址分类与事件归属地解析
internal/detection/  内置规则导入、跨端受限规则与双阶段匹配
internal/ai/         DeepSeek/GLM 等 OpenAI 兼容模型分析边界
internal/nodepki/    节点 CA、服务端证书和客户端证书签发
internal/config/     Server 配置加载
internal/store/      GORM 模型、迁移与种子数据
internal/httpapi/    Gin API、认证、RBAC、WebSocket Hub
web/                 React + Arco Design 控制台
honeypot-templates-server/ Web 蜜罐原始资源与 config.json
docs/                产品架构设计
```

## 验证

```bash
go test ./...
cd web && npm run build
sh -n scripts/install-server.sh scripts/install-agent.sh scripts/build-release.sh scripts/build-agent-downloads.sh scripts/build-agent-release.sh scripts/build-all-agent-releases.sh
node scripts/smoke-test.mjs
node scripts/smoke-sense.mjs
```

## 许可证

项目自有代码及其他明确由本项目版权方授权的内容采用 [PolyForm Noncommercial License 1.0.0](LICENSE)：允许在许可条款规定的非商业目的范围内使用、修改和分发，禁止未经书面授权的商业使用。

仓库包含的第三方依赖、兼容性资源、产品标识和静态页面不因收录于本仓库而改用上述许可证；它们继续适用各自的许可证、版权和商标规则。详见 [第三方声明](THIRD_PARTY_NOTICES.md)。商业授权请通过本仓库的 GitHub Issues 联系版权方。
