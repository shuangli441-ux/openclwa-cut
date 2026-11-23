# MEP Core 1 · 表格示例 / Table example

[核心规范 / Core spec](../../CORE.md) · [报文样例 / Wire fixtures](fixtures.json) · [参考代码 / Reference code](demo.mjs) · [测试 / Tests](demo.test.mjs)

这是非视频领域的**可运行示例**，不是某个表格软件已经接入的声明。助手读取清单与表格快照，按稳定的行 ID 请求把两条记录改为「完成」。宿主一次性校验并更新两行，返回结果版本和变更行。旧版本冲突不会写入；重复请求不应再次改表。单个无效行导致整批拒绝。

This runnable example shows a non-video domain, not an integration with a shipping spreadsheet app. The assistant requests an update by stable row ID. The host validates the whole batch, applies it atomically, and returns the resulting revision and changed rows. Stale writes cause no mutation; replays do not apply twice. An invalid row rejects the whole batch.

Run from the repository root with Node.js 20+:

```sh
node examples/core-table/demo.mjs
node --test examples/core-table/demo.test.mjs
```

The demo uses synthetic in-memory state and a fixed test credential. It does not implement OS permissions, directory transport or durable restart recovery. Production hosts must implement those requirements from Core. The domain JSON Schemas in [schemas/](schemas/) describe the same fixture values; the executable example checks that subset directly and has no external dependencies.
