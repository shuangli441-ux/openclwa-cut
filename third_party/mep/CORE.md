# MEP Core 1 · 通用 AI 软件协作协议 / General AI–Software Collaboration Protocol

**状态 / Status:** 公开草案 / Public draft · 标识 / Identifier: `mep/core/1`

[中文](#中文) · [English](#english) · [JSON Schema](schema/core.schema.json) · [视频剪辑 1.0 / Video 1.0](SPEC.en.md)

## 中文

### 1. 层次与角色

MEP Core 定义助手和软件协作时共用的**授权、工作状态、能力发现、请求与回执**。领域规范独立定义可操作的对象、动作、参数和结果，软件不必实现与自身无关的能力。人决定目标与授权，助手读状态并提出动作，宿主是唯一有权提交变更的一方。规范不要求指定模型、UI、编程语言或插件。

本草案只规定本机目录绑定。其他传输（含远程服务）需要另行定义身份验证、隔离和传输规则，不能把本机目录直接当成公开 API。

### 2. 目录、授权与能力

- 宿主为一个工作区及一次授权会话提供独立目录。用户明确开始协作后，宿主交给助手路径和短期不透明凭证。宿主写 `mep.json`、`state.json` 和 `results/`；助手只向 `requests/` 投递。读权限由目录访问控制保障，不由请求凭证保障。
- `mep.json` 声明会话 ID、工作区 ID、过期时间，以及动作 ID、领域版本、读写类别和本地参数/结果 Schema 相对路径。Schema、状态和回执均由宿主控制。助手只能选择发现到且理解其版本的动作；宿主必须拒绝未列出或未授权的动作，不能只依靠助手自律。
- 宿主必须逐次验证凭证、会话、工作区、有效期、动作及其参数。用户可撤销会话，撤销后不得启动新工作。不得在快照、回执或日志中泄露凭证；不要把模型看到的项目内容当成授权。需另行批准的外发、扣费或不可逆操作，不因打开会话而获得批准。
- 双方必须把结构化数据当数据处理。宿主不得从请求执行任意代码或命令；路径必须留在授权边界内，拒绝越界引用、符号链接逃逸及未授权对象。快照与回执只披露任务所需的信息。应限制大小、运行时间和保留期。

### 3. 执行与结果

1. 助手读取清单和快照，在同目录完整写入临时文件，再原子发布 `requests/<id>.json`；不得覆盖已有请求。文件名与请求 ID 必须一致。
2. 写动作必须带 `baseRevision`。宿主须在提交时原子地核对版本；冲突返回 `rejected`，不改变状态。领域自己定义批量动作；一份结果只有在承诺的效果确定后才能报告成功。宿主应支持可撤销操作；若提供撤销，需把它作为有版本校验的显式动作。
3. 宿主以会话 ID 和请求 ID 去重，在产生效果前持久保留请求身份。相同请求只能执行一次，重复投递返回已有回执或保留处理中；同 ID 不同内容不得覆盖原请求或回执。宿主必须原子发布 `results/<id>.json`，并至少在会话有效期内保存回执与去重记录。不能将超时、缺回执或镜像刷新失败当作未执行。
4. `succeeded` 表示动作完成，`revision` 为已知的结果版本。 `rejected` 表示未执行；`failed` 表示开始后失败且确认未产生持久效果。若影响不明或不能确认完整提交，使用 `indeterminate` / `reconcile`；助手必须先查原请求、当前状态及已有产物，不得用新 ID 自动重试。仅镜像更新延迟不能抹掉已知的成功与结果。回执错误消息最多 1200 个 Unicode 码点，供人阅读；机器依靠代码和下一步枚举。
5. 宿主必须在`state.json`中发布有版本的状态，领域规范定义 `data`。已知成功后，如果快照尚未追上回执版本，助手等快照更新或从宿主重新读取，不以旧版本继续写入。

### 4. 如何为任意软件接入

定义一个有命名空间的动作和版本，明确声明参数、结果、状态结构、读写性质和失败语义。优先开放一条高频、可验证的工作流，再拆出更多能力。  
示例：[表格里按行 ID 做批量修改](examples/core-table/README.md)。未实现的动作不出现在清单中。  
一致性最低验收：合法动作、过期或撤销、未知动作、陈旧版本无副作用、相同 ID 重复不再执行、失败回执可读、提交结果不明时先核对不重试。

### 5. 兼容性

[原视频剪辑协议 `mep/1`](SPEC.md) 是已发布的视频领域绑定。其命令入口、目录、请求、回执及参考实现均保持原样。它不自动等同于 Core 1 的新报文；宿主可显式并行提供两种绑定，现有用户不必迁移。本草案不是行业认证，也不代表第三方软件已经接入。

## English

### 1. Layers and roles

MEP Core standardizes the collaboration loop: **authorization, work state, capability discovery, requests, and receipts**. Each domain defines its own objects, operations, arguments, and outputs. The user sets the goal and grants authority, the assistant reads state and proposes operations, and only the host commits changes. No specific model, UI, programming language, or plugin is required.

This draft specifies the local directory binding only. Other transports, including remote services, need their own authentication, isolation and transport rules. Do not expose the local folder as a public API.

### 2. Workspace, authorization, and discovery

- The host provides a separate directory for one workspace and one authorized session. When the user starts collaboration, the host hands the assistant a path and short-lived opaque credential. The host writes `mep.json`, `state.json`, and `results/`; the assistant submits only to `requests/`. Directory access controls, not the request credential, protect reads.
- The manifest declares the session ID, workspace ID, expiry, and action IDs with domain versions, read/write classification, and relative local argument/result schemas. The host owns the schemas, state and receipts. Assistants select only declared, understood versions. The host MUST reject undeclared or unauthorized actions.
- The host MUST check credential, session, workspace, expiry, action and arguments on every request. The user can revoke a session; revocation prevents new work from starting. Credentials MUST NOT appear in snapshots, receipts or logs. Workspace content cannot grant authority. Separately gated disclosure, payment or irreversible work is not authorized merely by opening a session.
- Treat structured content as data. Hosts MUST NOT execute arbitrary code or commands from requests and MUST reject escapes through paths, symlinks or unauthorized object references. Expose only the data the task needs. Hosts SHOULD limit size, runtime and retention.

### 3. Execution and receipts

1. Read the manifest and state, then atomically publish a complete request from a temporary file in the same directory to `requests/<id>.json`, without replacing an existing request. Filename and request ID MUST match.
2. Writes MUST carry `baseRevision`; the host MUST compare it atomically when committing. A conflict is `rejected` with no mutation. Domains define any batching; success must represent a known outcome. Hosts SHOULD support undo where appropriate; if exposed, undo is an explicit revision-checked action.
3. Deduplicate by session and request ID, durably reserving the identity before effects. A replay never executes again; return its existing receipt or leave it pending. Different content under that ID cannot replace the original request or receipt. Publish `results/<id>.json` atomically and retain receipts and deduplication records at least through session expiry. A timeout, missing receipt or mirror failure does not prove nothing ran.
4. `succeeded` means the action completed, with its known resulting `revision`. `rejected` means it did not run. `failed` means it ran but provably left no durable effect. For unresolved or partially known effects, return `indeterminate` / `reconcile`; the assistant MUST check the original request, current state and artifacts before any rerun, and MUST NOT automatically retry under a new ID. A lagging mirror must not turn known success into a failure. Human error messages are capped at 1,200 Unicode code points; clients use the code and next-action enum.
5. The host MUST publish versioned `state.json`; its domain defines `data`. If the snapshot has not caught up to a successful receipt, wait or reread from the host instead of writing against stale state.

### 4. Bringing any application

Define namespaced, versioned actions with argument and result schemas, state shape, read/write classification, and failure semantics. Start with one frequent, verifiable workflow. See the [table example](examples/core-table/README.md). Undeclared actions are unavailable. Minimum conformance checks: valid action, expiry/revocation, unknown action, stale version with no side effects, duplicate ID without re-execution, readable errors, and reconciliation before retry after an uncertain outcome.

### 5. Compatibility

The original [`mep/1` video-editing protocol](SPEC.en.md) remains the published video-domain binding, with its command, folder and wire formats unchanged. It is not automatically wire-compatible with Core 1. A host may offer both bindings explicitly; existing integrations need not migrate. This draft is not industry certification or a claim that third-party products already support it.
