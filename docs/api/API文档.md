# WeaveNet 织网穿透 API 接口文档

> 项目：准商业级 FRP 全套程序（服务端、客户端、管理端、官网）
> 品牌：WeaveNet・织网穿透
> 文档：CQFISH&喵酱出品
> 日期：2026-08-15

本文档依据《WeaveNet 织网穿透 设计文档》（2026-08-15）编写，覆盖用户侧 API、管理后台 API、内核联动 API、客户端 API 与应急 API 的接口划分、状态码规划与认证体系说明。所有内容均取自设计文档，未作超出设计文档的扩展。

## 目录

1. 通用约定
2. 用户侧 API
3. 管理后台 API
4. 内核联动 API（frps 专用 Token 鉴权）
5. 客户端 API（Electron 专用）
6. 认证体系说明
7. 应急 API（逃生通道）

## 第 1 章 通用约定

### 1.1 Base URL

- 面板 API（含用户侧、管理后台、内核联动、客户端分组）：`https://<面板域名>/`，生产环境经 HTTPS 反向代理对外提供，开发环境为 `http://localhost:8000`（8000 端口为面板与官网共用）。
- 应急 API：`http://<内网IP>:9001/`，独立端口，仅内网/VPN 访问。

### 1.2 统一响应格式

所有面板 API 返回统一的 JSON 结构：

成功：`{ "code": 0, "message": "ok", "data": {...} }`

失败：`{ "code": 409, "business_code": 3001, "message": "远程端口已被占用", "data": null }`

字段说明：

| 字段 | 说明 |
| --- | --- |
| code | HTTP 状态码，0 表示成功 |
| business_code | 业务码，仅在失败时出现，用于精确定位错误原因 |
| message | 中文友好提示信息 |
| data | 业务数据，失败时为 null |

### 1.3 HTTP 状态码

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

### 1.4 业务码（响应体 business_code 字段）

| 业务码 | 含义 | 分类 |
| --- | --- | --- |
| 1001 | 邮箱验证码错误 | 认证 |
| 1002 | 验证码过期 | 认证 |
| 1003 | 邮箱已注册 | 认证 |
| 1004 | 用户名已注册 | 认证 |
| 1005 | 邮箱未验证 | 认证 |
| 1006 | 账号已封禁 | 认证 |
| 2001 | 套餐额度不足 | 套餐 |
| 2002 | 带宽超限 | 套餐 |
| 3001 | 远程端口冲突 | 隧道/域名/节点 |
| 3002 | 子域名被占用 | 隧道/域名/节点 |
| 3003 | 免费域名额度不足 | 隧道/域名/节点 |
| 3004 | 节点离线或维护中 | 隧道/域名/节点 |
| 4001 | 签到重复 | 签到/积分 |
| 4002 | 积分不足 | 签到/积分 |
| 5001 | 工单已关闭 | 工单/公告 |
| 5002 | 公告已下线 | 工单/公告 |
| 9001 | 内部错误 | 内部 |

### 1.5 认证方式

- 用户与管理员使用会话 Token 鉴权，请求头携带 `Authorization: Bearer <token>`。
- frps 节点使用独立的 Agent Token，仅允许访问 `/api/agent/*` 分组。
- 中间件按角色（user / admin）校验接口访问权限。

## 第 2 章 用户侧 API

用户侧 API 共 10 个分组，面向注册用户。以下主要端点为按设计文档功能划分的接口路径规划，具体参数细节以最终实现为准。

### 2.1 /api/auth

功能：注册、登录、邮箱验证、找回密码、登出。

主要端点：

| 方法 | 端点 | 说明 |
| --- | --- | --- |
| POST | /api/auth/register | 用户注册 |
| POST | /api/auth/login | 用户登录 |
| POST | /api/auth/email-verify | 邮箱验证 |
| POST | /api/auth/forgot-password | 找回密码 |
| POST | /api/auth/logout | 登出（响应 204） |

### 2.2 /api/user

功能：个人资料、改密码/邮箱、套餐额度、操作日志。

主要端点：

| 方法 | 端点 | 说明 |
| --- | --- | --- |
| GET | /api/user/profile | 获取个人资料 |
| PUT | /api/user/profile | 修改个人资料 |
| PUT | /api/user/password | 修改密码 |
| PUT | /api/user/email | 修改邮箱 |
| GET | /api/user/quota | 查询套餐额度 |
| GET | /api/user/logs | 查询操作日志 |

### 2.3 /api/tunnels

功能：隧道 CRUD、启停、配置生成、状态。

主要端点：

| 方法 | 端点 | 说明 |
| --- | --- | --- |
| GET | /api/tunnels | 隧道列表 |
| POST | /api/tunnels | 创建隧道（响应 201） |
| GET | /api/tunnels/{id} | 隧道详情 |
| PUT | /api/tunnels/{id} | 修改隧道 |
| DELETE | /api/tunnels/{id} | 删除隧道（响应 204） |
| POST | /api/tunnels/{id}/start | 启动隧道 |
| POST | /api/tunnels/{id}/stop | 停止隧道 |
| POST | /api/tunnels/{id}/config | 生成隧道配置 |
| GET | /api/tunnels/{id}/status | 查询隧道状态 |

隧道类型为全家桶：TCP / UDP / HTTP / HTTPS / STCP 安全隧道 / XTCP 点对点 / KCP 加速 / 负载均衡 / 免费三级域名。创建失败时按业务码明确报错（端口冲突/超限/节点离线等）。

### 2.4 /api/nodes

功能：节点列表与状态。

主要端点：

| 方法 | 端点 | 说明 |
| --- | --- | --- |
| GET | /api/nodes | 节点列表 |
| GET | /api/nodes/{id}/status | 节点状态 |

节点超过 60s 无心跳将被标记离线。

### 2.5 /api/domains

功能：免费域名申请/释放。

主要端点：

| 方法 | 端点 | 说明 |
| --- | --- | --- |
| GET | /api/domains | 免费域名列表 |
| POST | /api/domains | 申请免费域名（响应 201） |
| DELETE | /api/domains/{id} | 释放免费域名（响应 204） |

免费域名链路：用户申请子域名，面板校验占用后生成 `xxx.weave.test` 占位域名，由 frps 按域名路由。免费域名额度由套餐决定（免费版 1 条，普通会员 4 条，高级会员 8 条，超级会员 16 条）。

### 2.6 /api/stats

功能：流量统计、在线概览。

主要端点：

| 方法 | 端点 | 说明 |
| --- | --- | --- |
| GET | /api/stats/traffic | 流量统计 |
| GET | /api/stats/overview | 在线概览 |

历史流量按日聚合存储于 traffic_stats，在线状态与实时流量缓存于 Redis。

### 2.7 /api/signin

功能：每日签到。

主要端点：

| 方法 | 端点 | 说明 |
| --- | --- | --- |
| POST | /api/signin | 每日签到 |

规则：每日签到得 10 积分，连续 7 天额外 +30 积分；重复签到返回业务码 4001。签到规则由管理员在后台配置（system_configs 的签到规则项）。

### 2.8 /api/points

功能：积分流水、兑换会员。

主要端点：

| 方法 | 端点 | 说明 |
| --- | --- | --- |
| GET | /api/points/logs | 积分流水 |
| POST | /api/points/exchange | 积分兑换会员 |

兑换规则：300 积分 = 1 个月普通会员；积分不足返回业务码 4002。兑换价由管理员在后台配置。

### 2.9 /api/announcements

功能：公告列表/详情。

主要端点：

| 方法 | 端点 | 说明 |
| --- | --- | --- |
| GET | /api/announcements | 公告列表（标题/author/时间） |
| GET | /api/announcements/{id} | 公告详情 |

公告包含 author（发布人/部门名称）字段；已下线公告返回业务码 5002。

### 2.10 /api/tickets

功能：工单创建/回复/关闭。

主要端点：

| 方法 | 端点 | 说明 |
| --- | --- | --- |
| POST | /api/tickets | 创建工单（响应 201） |
| GET | /api/tickets | 工单列表 |
| GET | /api/tickets/{id} | 工单详情 |
| POST | /api/tickets/{id}/reply | 回复工单 |
| POST | /api/tickets/{id}/close | 关闭工单 |

已关闭的工单不可回复/关闭，返回业务码 5001。

## 第 3 章 管理后台 API

管理后台 API 共 8 个分组，仅 admin 角色可访问，中间件校验权限。

### 3.1 /api/admin/users

功能：用户列表/封禁/重置密码/调套餐。

主要端点：

| 方法 | 端点 | 说明 |
| --- | --- | --- |
| GET | /api/admin/users | 用户列表 |
| POST | /api/admin/users/{id}/ban | 封禁用户 |
| POST | /api/admin/users/{id}/unban | 解除封禁 |
| POST | /api/admin/users/{id}/reset-password | 重置密码 |
| PUT | /api/admin/users/{id}/plan | 调整套餐 |

封禁用户对应业务码 1006（账号已封禁）；调整套餐同时记录 user_plan_logs 套餐变更记录。

### 3.2 /api/admin/nodes

功能：节点增删改/启停/限速配置。

主要端点：

| 方法 | 端点 | 说明 |
| --- | --- | --- |
| GET | /api/admin/nodes | 节点列表 |
| POST | /api/admin/nodes | 新增节点（响应 201） |
| PUT | /api/admin/nodes/{id} | 修改节点 |
| DELETE | /api/admin/nodes/{id} | 删除节点（响应 204） |
| POST | /api/admin/nodes/{id}/start | 启用节点 |
| POST | /api/admin/nodes/{id}/stop | 停用节点 |
| PUT | /api/admin/nodes/{id}/speed | 节点限速配置 |

节点字段包括 name、address、port、status、speed_limit_mbps、remark。

### 3.3 /api/admin/tunnels

功能：全局隧道/强制下线。

主要端点：

| 方法 | 端点 | 说明 |
| --- | --- | --- |
| GET | /api/admin/tunnels | 全局隧道列表 |
| POST | /api/admin/tunnels/{id}/offline | 强制下线隧道 |

### 3.4 /api/admin/plans

功能：套餐档位配置。

主要端点：

| 方法 | 端点 | 说明 |
| --- | --- | --- |
| GET | /api/admin/plans | 套餐列表 |
| PUT | /api/admin/plans/{id} | 修改套餐档位 |

套餐档位：免费版（8Mbps/3 隧道/1 域名）、普通会员（16Mbps/6 隧道/4 域名）、高级会员（24Mbps/10 隧道/8 域名）、超级会员（32Mbps/14 隧道/16 域名），流量均不限。套餐字段包括 name、speed_limit_mbps、tunnel_limit、domain_limit、sort。

### 3.5 /api/admin/announcements

功能：公告发布（含 author）/下线。

主要端点：

| 方法 | 端点 | 说明 |
| --- | --- | --- |
| POST | /api/admin/announcements | 发布公告（含 author 发布人字段，响应 201） |
| GET | /api/admin/announcements | 公告列表 |
| PUT | /api/admin/announcements/{id} | 修改公告 |
| POST | /api/admin/announcements/{id}/offline | 公告下线 |

公告字段包括 title、content、author（发布人/部门名称）、status、created_at。

### 3.6 /api/admin/config

功能：签到积分规则、SMTP、占位域名。

主要端点：

| 方法 | 端点 | 说明 |
| --- | --- | --- |
| GET | /api/admin/config | 获取系统配置 |
| PUT | /api/admin/config | 修改系统配置 |

配置项存储于 system_configs（key/value），包括 SMTP、签到规则、占位域名、兑换价。

### 3.7 /api/admin/logs

功能：操作日志/运行日志。

主要端点：

| 方法 | 端点 | 说明 |
| --- | --- | --- |
| GET | /api/admin/logs/operation | 操作日志 |
| GET | /api/admin/logs/runtime | 运行日志 |

操作日志存储于 operation_logs（admin_id、action、target_type、target_id、detail、created_at）。

### 3.8 /api/admin/dashboard

功能：统计看板。

主要端点：

| 方法 | 端点 | 说明 |
| --- | --- | --- |
| GET | /api/admin/dashboard | 统计看板 |

## 第 4 章 内核联动 API（frps 专用 Token 鉴权）

frps 使用独立 Agent Token 鉴权，仅允许访问 `/api/agent/*`，与用户会话体系隔离。

### 4.1 POST /api/agent/register

功能：节点注册。frps 启动时向面板注册本节点信息。

### 4.2 POST /api/agent/heartbeat

功能：心跳上报（状态/流量）。frps 每 30s 上报隧道在线状态、连接数、流量增量；面板入库并刷新 Redis。面板超过 60s 无心跳标记节点离线。

### 4.3 GET /api/agent/tunnels

功能：拉取本节点隧道 + 限速配置。frps 每 10s 轮询面板获取本节点名下隧道与用户限速配置，热更新，无需重启。

## 第 5 章 客户端 API（Electron 专用）

Electron 桌面客户端（Windows 10/11 x86_64）专用，托管 frpc 子进程。

### 5.1 POST /api/client/login

功能：客户端登录。登录成功后拉取隧道列表并生成配置。

### 5.2 GET /api/client/tunnels

功能：拉取用户隧道 + 节点信息。客户端据此生成 frpc.toml，启动 frpc 子进程。

### 5.3 POST /api/client/config

功能：生成 frpc.toml 配置。用户无需手写配置，由 Electron 客户端生成并管理 frpc 子进程生命周期。

## 第 6 章 认证体系说明

- 用户与管理员使用会话 Token 鉴权，请求头携带 `Authorization: Bearer <token>`。会话 Token 在 Redis 与 SQLite 双写存储（sessions 表：token、user_id、expires_at），会话 Token 为 64 位随机值，30 天过期。
- 角色分为 user / admin，由中间件校验接口权限。
- frps 节点使用独立 Agent Token，仅允许访问 `/api/agent/*` 分组，实现权限隔离。
- 邮箱验证码用于注册激活与找回密码两个场景（email_codes 表含 purpose 字段），5 分钟有效，并做防暴力（次数限制）。
- 登录/注册/签到接口按 IP + 账号维度限流（对应 HTTP 429）。
- 密码使用 bcrypt 存储（>=10 轮）。
- 隧道访问在 ORM 层强制 user_id 过滤，防止越权。

## 第 7 章 应急 API（逃生通道）

### 7.1 设计要点

- 独立进程/容器，与面板完全解耦，面板/数据库/Redis 全部宕机也能响应。
- 独立端口 9001，仅内网/VPN 访问，预共享强随机应急密钥鉴权。
- 零依赖（标准库实现），不连接数据库/Redis。
- 每次操作写本地审计日志。

### 7.2 独立约定

应急 API 使用独立的状态码约定，与面板 API 不同：

| 状态码 | 语义 |
| --- | --- |
| 200 | 成功 |
| 401 | 密钥错误 |
| 423 | 已锁定 |
| 400 | 参数错误 |

### 7.3 接口清单

| 方法 | 端点 | 功能 |
| --- | --- | --- |
| GET | /health | 存活检查 |
| GET | /status | 系统状态（CPU/内存/磁盘/容器状态） |
| GET | /logs?service=panel&lines=200 | 拉取服务日志 |
| GET | /data?type=users\|tunnels\|nodes | 直连 SQLite 拉取只读数据 |
| POST | /restart?service=panel\|frps\|redis\|all | 重启服务 |
| POST | /stop?service=all | 停止服务 |
| POST | /start?service=panel\|frps | 启动服务 |
| POST | /reboot | 重启服务器 |
| POST | /exec?cmd=... | 受限白名单命令执行（可选） |

### 7.4 安全防护

- 强随机密钥（>=32 位），独立配置文件存储。
- 可选 IP 白名单。
- 连续失败 5 次锁定 10 分钟（对应 423 已锁定）。
- 危险操作（reboot）需二次确认参数。
