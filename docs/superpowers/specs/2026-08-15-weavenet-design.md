# WeaveNet 织网穿透 设计文档

> 项目：准商业级 FRP 全套程序（服务端、客户端、管理端、官网）
> 品牌：WeaveNet・织网穿透
> 对标：https://www.chmlfrp.net/
> 文档：CQFISH&喵酱出品
> 日期：2026-08-15

## 1. 项目概述

### 1.1 定位
准商业核心版内网穿透服务平台，包含四个子系统：FRP 服务端、桌面客户端、管理面板、官网。

### 1.2 目标规模
- 注册用户 1 万以内，在线隧道 500 以内
- 单机部署（Docker Compose），保留中间件支撑业务架构
- 整体内存占用目标 200MB 以内

### 1.3 商业化模式
- 四档套餐：免费版 / 普通会员 / 高级会员 / 超级会员
- 无真实支付，虚拟积分 + 每日签到体系，后续可平滑升级到真实支付

## 2. 技术栈

| 组件 | 技术 |
| --- | --- |
| 服务端/客户端内核 | Go，基于 fatedier/frp 二次开发（Apache 2.0 可商用） |
| 面板后端 | Python 3.12 + FastAPI，uvicorn 单进程 |
| 数据库 | SQLite（WAL 模式 + busy_timeout） |
| 缓存/会话 | 轻量 Redis（单实例） |
| 异步任务 | asyncio 后台任务（不引入 Celery/消息队列） |
| 面板前端 | Vue3 + Naive UI，构建为静态文件 |
| 官网 | FastAPI + Jinja2 服务端渲染 |
| 桌面客户端 | Electron（Windows 10/11 x86_64），托管 frpc 子进程 |
| 应急服务 | 独立轻量进程（标准库实现，零依赖） |
| 部署 | Docker Compose（4 容器） |

## 3. 系统架构

### 3.1 组件拓扑
- 用户浏览器：官网访问、面板操作、外部服务访问
- 应用层：官网（FastAPI+Jinja2）、面板 API（FastAPI）、面板前端（Vue3 静态）
- 数据层：SQLite（WAL）+ Redis（会话/缓存/实时状态）+ asyncio 后台任务
- 穿透层：frps 服务端（Go 二开，全家桶隧道类型）
- 客户端：Electron 桌面程序 + frpc（Windows 10/11）

### 3.2 数据流
1. 用户在面板创建隧道，面板写入 SQLite，隧道配置缓存到 Redis
2. frps 每 10s 轮询面板拉取本节点名下隧道与用户限速配置（热更新，无需重启）
3. Electron 客户端登录面板拉取隧道列表，生成 frpc.toml，启动 frpc 子进程
4. frpc 携带用户 Token 连接 frps，鉴权通过建立隧道
5. frps 每 30s 上报隧道在线状态、连接数、流量增量，面板入库并刷新 Redis
6. 面板心跳检测 frps，超过 60s 无心跳标记节点离线

### 3.3 容器拓扑（4 容器）
- weave-panel：FastAPI 面板+官网+静态前端，挂载 sqlite 数据卷
- weave-redis：redis:7-alpine
- weave-frps：Go 穿透服务端，暴露 7000（控制）/80/443（隧道）
- weave-emergency：应急服务，端口 9001（仅内网/VPN 访问）

### 3.4 端口规划
- 8000：面板 + 官网（HTTPS 反代后）
- 7000：frps 控制通道
- 80/443：HTTP(S) 隧道
- 9001：应急服务（仅内网/VPN）

## 4. 数据库模型（SQLite）

### 4.1 核心表（15 张）
| 表 | 说明 | 关键字段 |
| --- | --- | --- |
| users | 用户 | username, email, password_hash, email_verified, status, plan_id, plan_expires_at, points, created_at |
| plans | 套餐 | name, speed_limit_mbps, tunnel_limit, domain_limit, sort |
| nodes | 穿透节点 | name, address, port, status, speed_limit_mbps, remark |
| tunnels | 隧道 | user_id, node_id, name, type, local_ip, local_port, remote_port, subdomain, custom_domain, kcp, encryption, compression, status |
| domains | 免费域名 | user_id, tunnel_id, subdomain, full_domain, status, created_at |
| signin_logs | 签到记录 | user_id, signin_date, points, continuous_days |
| points_logs | 积分流水 | user_id, change, reason, created_at |
| tickets | 工单 | user_id, title, content, status, admin_reply, created_at |
| announcements | 公告 | title, content, **author**（发布人/部门名称）, status, created_at |
| traffic_stats | 流量统计 | tunnel_id, date, in_bytes, out_bytes |
| system_configs | 系统配置 | key, value（SMTP、签到规则、占位域名、兑换价） |
| sessions | 会话 | token, user_id, expires_at |
| email_codes | 邮箱验证码 | email, code, purpose, expires_at, used |
| operation_logs | 操作日志 | admin_id, action, target_type, target_id, detail, created_at |
| user_plan_logs | 套餐变更记录 | user_id, plan_id, reason, created_at |

### 4.2 存储策略
- 在线状态与实时流量缓存于 Redis，不入库
- 历史流量按日聚合入 traffic_stats

## 5. 套餐与积分体系

### 5.1 套餐档位
| 档位 | 限速 | 隧道数 | 免费域名 | 流量 |
| --- | --- | --- | --- | --- |
| 免费版 | 8Mbps | 3 条 | 1 条 | 不限 |
| 普通会员 | 16Mbps | 6 条 | 4 条 | 不限 |
| 高级会员 | 24Mbps | 10 条 | 8 条 | 不限 |
| 超级会员 | 32Mbps | 14 条 | 16 条 | 不限 |

### 5.2 签到与积分
- 每日签到得 10 积分，连续 7 天额外 +30 积分
- 积分兑换：300 积分 = 1 个月普通会员
- 管理员可后台调整规则
- 会员到期自动降级回免费版（每小时检查）

## 6. 隧道类型（全家桶）

- TCP / UDP / HTTP / HTTPS / STCP 安全隧道 / XTCP 点对点 / KCP 加速 / 负载均衡 / 免费三级域名

## 7. frps/frpc 内核二开设计

### 7.1 frps 二开点
1. 控制通道 API：向面板注册节点 + 心跳上报
2. 隧道配置热拉取：从面板实时获取本节点隧道与用户套餐限速
3. 客户端鉴权对接：frpc 登录校验面板签发的用户 Token
4. 限速联动：按用户套餐带宽上限动态限速
5. 状态上报：在线隧道、连接数、流量增量定时上报
6. 远程端口分配校验：防止端口冲突

### 7.2 frpc 对接
- Token 鉴权连接 frps
- 配置由 Electron 客户端生成 frpc.toml（用户无需手写）
- Electron 管理 frpc 子进程生命周期

### 7.3 限速链路
套餐限速 → 面板配置下发 → frps 按用户 Token 带宽限制 → 隧道吞吐受限

### 7.4 免费域名链路
用户申请子域名 → 面板校验占用 → 生成 xxx.weave.test（占位域名）→ frps 按域名路由

## 8. 面板功能模块与 API 划分

### 8.1 用户侧 API
| 分组 | 功能 |
| --- | --- |
| /api/auth | 注册、登录、邮箱验证、找回密码、登出 |
| /api/user | 个人资料、改密码/邮箱、套餐额度、操作日志 |
| /api/tunnels | 隧道 CRUD、启停、配置生成、状态 |
| /api/nodes | 节点列表与状态 |
| /api/domains | 免费域名申请/释放 |
| /api/stats | 流量统计、在线概览 |
| /api/signin | 每日签到 |
| /api/points | 积分流水、兑换会员 |
| /api/announcements | 公告列表/详情 |
| /api/tickets | 工单创建/回复/关闭 |

### 8.2 管理后台 API
| 分组 | 功能 |
| --- | --- |
| /api/admin/users | 用户列表/封禁/重置密码/调套餐 |
| /api/admin/nodes | 节点增删改/启停/限速配置 |
| /api/admin/tunnels | 全局隧道/强制下线 |
| /api/admin/plans | 套餐档位配置 |
| /api/admin/announcements | 公告发布（含 author）/下线 |
| /api/admin/config | 签到积分规则、SMTP、占位域名 |
| /api/admin/logs | 操作日志/运行日志 |
| /api/admin/dashboard | 统计看板 |

### 8.3 内核联动 API（frps 专用 Token 鉴权）
- /api/agent/register 节点注册
- /api/agent/heartbeat 心跳上报（状态/流量）
- /api/agent/tunnels 拉取本节点隧道+限速配置

### 8.4 客户端 API（Electron 专用）
- /api/client/login 客户端登录
- /api/client/tunnels 拉取用户隧道+节点信息
- /api/client/config 生成 frpc.toml 配置

### 8.5 认证体系
- 用户/管理员：会话 Token（Redis + SQLite 双写），请求头 Bearer 鉴权
- 角色：user / admin，中间件校验
- frps 节点：独立 Agent Token（仅允许 /api/agent/*）
- 邮箱验证码：注册激活 + 找回密码，5 分钟有效

### 8.6 后台任务（asyncio）
- 会员过期检查（每小时）：到期自动降级免费版
- 流量日汇总（每 5 分钟）：Redis 增量 → traffic_stats
- 邮件发送队列（内存队列 + 重试 3 次）
- 节点心跳超时标记（每 60s 扫描）

## 9. 应急 API（逃生通道）

### 9.1 设计要点
- 独立进程/容器，与面板完全解耦，面板/数据库/Redis 全挂也能响应
- 独立端口 9001，预共享强随机应急密钥鉴权
- 零依赖（标准库实现），不连接数据库/Redis
- 每次操作写本地审计日志

### 9.2 接口清单
| 接口 | 功能 |
| --- | --- |
| GET /health | 存活检查 |
| GET /status | 系统状态（CPU/内存/磁盘/容器状态） |
| GET /logs?service=panel&lines=200 | 拉取服务日志 |
| GET /data?type=users\|tunnels\|nodes | 直连 SQLite 拉取只读数据 |
| POST /restart?service=panel\|frps\|redis\|all | 重启服务 |
| POST /stop?service=all | 停止服务 |
| POST /start?service=panel\|frps | 启动服务 |
| POST /reboot | 重启服务器 |
| POST /exec?cmd=... | 受限白名单命令执行（可选） |

### 9.3 安全防护
- 强随机密钥（>=32 位），独立配置文件
- 可选 IP 白名单
- 连续失败 5 次锁定 10 分钟
- 危险操作（reboot）需二次确认参数

## 10. 官网页面（11 页）

| 页面 | 路径 | 内容 |
| --- | --- | --- |
| 首页 | / | Hero、特色、场景、三步教程、定价摘要、FAQ 摘要 |
| 下载中心 | /download | Windows 客户端下载、版本、更新说明 |
| 文档 | /docs | 快速开始、隧道类型、免费域名、FAQ |
| 公告 | /announcements | 公告列表（标题/author/时间）+ 详情 |
| 定价 | /pricing | 四档套餐完整对比 + 签到兑换说明 |
| 关于我们 | /about | 品牌介绍、联系方式 |
| 服务条款 | /terms | 用户协议、使用规范、违规处理 |
| 隐私政策 | /privacy | 数据收集、加密、存储说明 |
| 更新日志 | /changelog | 客户端/面板版本历史 |
| 服务状态 | /status | 节点在线状态与历史可用率（公开） |
| 帮助中心 | /help | FAQ 汇总 + 工单入口 |

## 11. 视觉规范

- 风格：玻璃拟态浅色（半透明玻璃卡片、渐变边缘羽化、无阴影、无浮动）
- 主色：青蓝渐变（#0ea5e9 → #06b6d4）
- 点缀：暖橙（#f59e0b → #f97316），用于 CTA 按钮、徽章
- 面板布局：顶部栏 + 左侧菜单 + 中间内容（经典左导航）
- 禁用紫色系、禁用 Emoji（图标用 SVG）

## 12. API 状态码规划

### 12.1 HTTP 状态码
| 状态码 | 语义 |
| --- | --- |
| 200 | 查询/操作成功 |
| 201 | 资源创建成功 |
| 204 | 删除成功、登出成功 |
| 400 | 参数格式错误 |
| 401 | 未登录或 Token 失效 |
| 403 | 无权限 |
| 404 | 资源不存在 |
| 409 | 业务冲突 |
| 422 | Pydantic 校验失败 |
| 429 | 限流触发 |
| 500 | 服务端未预期异常 |
| 503 | 依赖不可用（frps 离线/Redis 不可用） |

### 12.2 业务码（响应体 code 字段）
- 1001 邮箱验证码错误 / 1002 验证码过期 / 1003 邮箱已注册 / 1004 用户名已注册 / 1005 邮箱未验证 / 1006 账号已封禁
- 2001 套餐额度不足 / 2002 带宽超限
- 3001 远程端口冲突 / 3002 子域名被占用 / 3003 免费域名额度不足 / 3004 节点离线或维护中
- 4001 签到重复 / 4002 积分不足
- 5001 工单已关闭 / 5002 公告已下线
- 9001 内部错误

### 12.3 统一响应格式
```
成功：{ "code": 0, "message": "ok", "data": {...} }
失败：{ "code": 409, "business_code": 3001, "message": "远程端口已被占用", "data": null }
```

### 12.4 应急 API 独立约定
200 成功 / 401 密钥错误 / 423 已锁定 / 400 参数错误

## 13. 错误处理与测试策略

### 13.1 错误处理
- 统一错误格式（code + message），中文友好提示
- 隧道创建失败：端口冲突/超限/节点离线均有明确报错
- frps 掉线：标记节点离线，不阻塞其他功能
- SMTP 失败：队列重试 3 次
- SQLite 锁冲突：WAL + busy_timeout，写操作重试
- 客户端网络中断：Electron 指数退避重连

### 13.2 测试策略
- 单元测试（pytest）：模型、限速计算、签到规则、套餐校验、Token 生成
- 接口测试：pytest + httpx 覆盖全部 API（含鉴权、越权用例）
- 内核测试（Go）：frps/frpc 基础隧道连通性
- 端到端：docker compose 起全套 → 注册 → 建隧道 → 客户端连接 → 公网访问验证
- 应急演练：模拟面板崩溃，验证应急 API 流程
- 性能冒烟：500 隧道在下面板响应、Redis 状态刷新

### 13.3 安全设计
- 密码 bcrypt（>=10 轮）
- 会话 64 位随机 Token，30 天过期
- 验证码防暴力（次数限制）
- 登录/注册/签到限流（IP + 账号维度）
- 隧道越权防护（ORM 层强制 user_id 过滤）
- 远程端口池管理，禁止系统保留端口

## 14. 项目目录结构（Monorepo）

```
e:\FRPserver\
├── docs/                  # 需求/设计/计划文档
├── server/                # Go frps/frpc 二开（fork fatedier/frp）
├── panel/                 # 面板+官网（FastAPI）
│   ├── app/               # 主应用（API/模型/任务）
│   ├── web/               # 面板前端（Vue3+Naive UI）
│   └── templates/         # 官网 Jinja2 模板
├── client/                # Electron 桌面客户端（Windows）
├── emergency/             # 应急服务（独立轻量进程）
└── deploy/                # Docker Compose + 配置（.env.example 等）
```

## 15. 交付范围说明

- 开放 API：暂不对外开放，内部接口架构预留
- 界面语言：简体中文（暂不做 i18n）
- 客户端平台：首发仅 Windows 10/11，其他平台后续扩展
- 免费域名：占位域名（weave.test）开发，上线前替换正式域名并配置泛解析
- HTTPS 证书：手动上传证书配置（不支持自动签发）
