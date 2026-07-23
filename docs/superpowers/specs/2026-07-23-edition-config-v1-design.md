# EditionConfig V1 后端设计

## 目标

为 PandaWiki 增加部署级行业版本配置，支持 `common`、`research`、`legal`、`finance` 四套预置版本。第一阶段只增加后端模型、预置配置、JSON 存储和管理员 API，不修改前端、PromptRepo、ReportProfile、知识库、文档、权限、RAG 或向量数据。

## 存储模型

复用 `system_settings` 表，使用 key `edition_config`。数据库只保存当前选择和管理员覆盖项，不复制整套预置配置：

```json
{
  "schema_version": 1,
  "edition_id": "legal",
  "edition_version": "1.0.0",
  "overrides": {
    "product_name": "XX法律知识平台",
    "home_description": "企业内部法律资料与报告平台"
  },
  "updated_at": "2026-07-23T01:00:00Z",
  "updated_by": 1001
}
```

`schema_version` 是持久化结构版本；`edition_id` 是预置配置标识；`edition_version` 是预置配置版本；`overrides` 只允许修改公开品牌、术语、功能、文档类型、关系类型和默认 Prompt 字段。`updated_at`、`updated_by` 由服务端写入，客户端不能覆盖。

## 完整配置结构

最终读取结果为预置配置与 overrides 合并后的对象：

```json
{
  "schema_version": 1,
  "edition_id": "legal",
  "edition_version": "1.0.0",
  "product_name": "法律知识库",
  "short_name": "Legal Wiki",
  "branding": { "logo": "/assets/editions/legal.svg" },
  "home_description": "企业内部法律资料与报告平台",
  "terminology": { "knowledge_base": "案件资料库", "document": "案卷", "folder": "案卷目录", "report": "法律报告" },
  "enabled_features": ["chat", "search"],
  "document_types": ["案件资料", "案卷", "合同", "证据", "法条", "判例"],
  "relation_types": ["引用", "关联案件", "相关法条"],
  "default_prompts": { "chat": "", "summary": "" }
}
```

JSON Schema 约束：顶层必需 `schema_version`、`edition_id`、`edition_version`、`product_name`、`short_name`、`branding`、`home_description`、`terminology`、`enabled_features`、`document_types`、`relation_types`、`default_prompts`；字符串不得为空（Prompt 允许为空以触发系统默认回退）；数组元素必须为非空字符串；`edition_id` 只能是四个预置 ID；`schema_version` 只能为 1；服务端拒绝未知字段和未知 override 字段。

## API

### GET `/api/v1/system/edition`

要求已认证用户，但不要求管理员角色。读取失败、缺少配置、JSON 损坏、schema 版本未知、edition ID 未知或合并结果非法时，返回 `common`，并记录结构化 warning/error。

公开 Wiki 前台当前通过 `/share/v1/app/web/info` 使用 `X-KB-ID` 且不要求后台登录。后续前台接入应新增只返回安全展示字段的 `/share/v1/edition`，或把安全字段附加到现有 AppInfo；不得让公开接口返回管理员审计字段或敏感 Prompt。该接口不属于第一阶段实现。

### PUT `/api/v1/system/edition`

请求只接受：

```json
{ "edition_id": "legal", "overrides": { "product_name": "XX法律知识平台" } }
```

服务端加载目标预置配置，校验 edition ID、override 字段和值，合并并校验最终结构后，在一次事务内写入。任何校验或写入失败都不更新数据库。客户端不能提交 `schema_version`、`edition_version`、`updated_at`、`updated_by` 或任意未知字段。

## 权限与审计

路由使用现有 `Authorize` 和 `ValidateUserRole(consts.UserRoleAdmin)`。PUT 从上下文获取当前用户 ID，记录修改前 edition/version、修改后 edition/version、用户 ID、时间和失败原因。项目暂不新增审计表，使用现有 logger；后续若已有统一审计机制，再接入统一事件。

## 兼容策略

不存在 `edition_config` 时按 common 返回，不自动写入数据库，避免初始化覆盖现有配置。未来 schema 升级通过显式迁移函数处理；未知 schema 只回退 common。合法 edition 的旧 `edition_version` 不回退 common，而是使用当前同 edition 预置、继续合并 overrides，并记录迁移日志；预置配置升级通过代码中的 `edition_version` 完成。

## 测试范围

- 预置配置完整性和四个 edition ID 校验。
- 缺失、损坏、未知 schema、未知 edition 均回退 common。
- PUT 接受合法选择与 overrides。
- PUT 拒绝未知字段、非法 edition、非法字段类型、客户端系统字段。
- 普通用户 PUT 返回 403。
- 数据库更新失败时旧配置保持不变。
- 更新内容包含服务端 updated_at/updated_by，且不覆盖其他 system_settings key。
