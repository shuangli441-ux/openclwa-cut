# 爪切 (clawcut) —— 一行命令，从素材到抖音精品成片

`clawcut` 是一个面向工业化短视频生产的 Go 语言 CLI 工具。它不是临时脚本，而是一个强调稳定、默认值合理、对小白友好的本地出片工具：给它素材，5 分钟内能跑通抖音标准成片、封面、字幕和发布文案。

当前版本重点能力：
- 默认输出抖音标准：`1080x1920`、`30fps`、`libx264 + aac`
- 项目 JSON 自动补默认值，减少必填字段
- 内置 AI 剪辑模块：按模板自动生成更适合抖音投放的时间线
- 支持只写 `aiEdit.scriptLines`，不手工写 `timeline` 也能自动出片
- 自动处理相对路径、中文路径、空格路径、`~` 路径、环境变量路径
- 字幕多级降级：硬字幕优先，失败时自动输出 `.ass`，不中断成片
- BGM 自动裁剪、自动淡出、ducking 压低人声背景乐
- 封面导出、品牌水印、发布文案交付文件
- 批量渲染目录项目、健康检查、Docker 镜像运行
- 支持 `go install github.com/shuangli441-ux/openclwa-cut/cmd/clawcut@latest`

## 命令总览

环境检查：
- `clawcut health`

剪辑渲染：
- `clawcut trim`
- `clawcut compress`
- `clawcut concat`
- `clawcut render`
- `clawcut render-dir`

项目管理：
- `clawcut project-init`
- `clawcut validate`

音乐能力：
- `clawcut music-init`
- `clawcut music-scan`
- `clawcut music-match`
- `clawcut music-mix`

模板能力：
- `clawcut template-init`
- `clawcut template-douyin-qa`
- `clawcut template-douyin-goods`
- `clawcut template-douyin-ads`

## 5 分钟快速开始

### 1. 安装

推荐直接安装可执行文件：

```bash
go install github.com/shuangli441-ux/openclwa-cut/cmd/clawcut@latest
```

本地开发调试也可以直接编译：

```bash
git clone https://github.com/shuangli441-ux/openclwa-cut.git
cd openclwa-cut
go build -o clawcut ./cmd/clawcut
```

如果你不想本地装 FFmpeg，也可以直接用 Docker：

```bash
docker build -t clawcut:local .
```

### 2. 检查环境

```bash
clawcut health
```

### 3. 一键初始化模板项目

不懂 JSON 配置时，优先用模板：

```bash
clawcut template-init \
  -kind douyin-qa \
  -dir ./projects/demo \
  -name 财税答疑-demo \
  -input /path/to/input.mp4 \
  -title "误操作预收款怎么撤销"
```

好物推荐场景：

```bash
clawcut template-init \
  -kind douyin-goods \
  -dir ./projects/goods-demo \
  -name 办公室好物-demo \
  -input /path/to/input.mp4 \
  -title "办公室高频好物推荐"
```

投放广告场景：

```bash
clawcut template-init \
  -kind douyin-ads \
  -dir ./projects/ads-demo \
  -name 秒账投放-demo \
  -input /path/to/input.mp4 \
  -title "进销存混乱？3 步帮你理顺" \
  -brand "秒账" \
  -cta "现在私信领取试用版"
```

### 4. 渲染

```bash
clawcut validate -project ./projects/demo/project.json
clawcut render -project ./projects/demo/project.json
```

### 5. 查看结果

渲染完成后，交付物会自动分开存放：
- 成片：`output/video/`
- 封面：`output/cover/`
- 字幕：`output/subtitles/`
- 报告与发布文案：`output/report/`

其中 `output/report/*.publish.txt` 可以直接复制到抖音发布后台。

## 推荐工作流

最省心的方式是模板初始化后再渲染：

```bash
clawcut template-init -kind douyin-qa -dir ./projects/demo -input /path/to/input.mp4
clawcut render -project ./projects/demo/project.json
```

如果你要从零建一个最小项目：

```bash
clawcut project-init \
  -dir ./projects/demo \
  -name demo \
  -input /path/to/input.mp4 \
  -title "误操作预收款怎么撤销" \
  -ai-kind douyin-qa \
  -scripts "先抛问题|中段给结论|结尾引导收藏"

clawcut render -project ./projects/demo/project.json
```

说明：
- `project-init` 现在默认会生成 AI 剪辑脚手架，包括 `aiEdit.templateKind`、`aiEdit.scriptLines` 和自动时间线。
- 如果你只想保留传统的整段 clip 时间线，可以追加 `-disable-ai`。

批量渲染整个目录：

```bash
clawcut render-dir -dir ./projects -pattern '*.json'
```

扫描本地音乐库：

```bash
clawcut music-scan -dir /path/to/music
clawcut music-match -style qa-short
```

说明：
- `qa-short`、`goods-recommend`、`promo-short`、`douyin-ads` 会自动映射到音乐库里的 `electronic / pop / upbeat / trap` 等目录风格。
- 如果项目里没有手工 `clip` 时间线，但填写了 `aiEdit.scriptLines`，`clawcut render` 和 `clawcut validate` 会在执行前自动补齐剪辑片段。
- `project-init` 支持 `-ai-kind`、`-scripts`、`-brand`、`-cta`，可以直接初始化 AI 口播项目。

默认音乐库路径会写到系统配置目录：
- macOS: `~/Library/Application Support/clawcut/music_library.json`
- Linux: `~/.config/clawcut/music_library.json`
- Windows: `%AppData%\clawcut\music_library.json`

## 常用模板命令

直接生成 QA 模板项目文件：

```bash
clawcut template-douyin-qa \
  -project ./projects/demo/project.json \
  -input /path/to/input.mp4 \
  -output ./projects/demo/output/video/final.mp4 \
  -title "误操作预收款怎么撤销"
```

直接生成好物推荐模板项目文件：

```bash
clawcut template-douyin-goods \
  -project ./projects/goods/project.json \
  -input /path/to/input.mp4 \
  -output ./projects/goods/output/video/final.mp4 \
  -title "办公室高频好物推荐"
```

直接生成投放广告模板项目文件：

```bash
clawcut template-douyin-ads \
  -project ./projects/ads/project.json \
  -input /path/to/input.mp4 \
  -output ./projects/ads/output/video/final.mp4 \
  -title "3 步解决库存混乱" \
  -brand "秒账" \
  -cta "现在私信领取试用版"
```

如果有品牌 Logo，可以继续加：

```bash
clawcut template-init \
  -kind douyin-qa \
  -dir ./projects/demo \
  -input /path/to/input.mp4 \
  -logo /path/to/logo.png
```

## 项目配置示例

```json
{
  "project": "douyin-demo",
  "assets": [
    {
      "id": "main",
      "type": "video",
      "path": "./assets/input.mp4"
    }
  ],
  "music": {
    "style": "douyin-ads"
  },
  "aiEdit": {
    "enabled": true,
    "mode": "smart",
    "templateKind": "douyin-ads",
    "maxDurationSeconds": 28,
    "hookSeconds": 3,
    "ctaSeconds": 5,
    "scriptLines": [
      "开头先抛痛点",
      "中段直接给方案",
      "结尾明确 CTA"
    ]
  },
  "cover": {
    "enabled": true,
    "title": "误操作预收款怎么撤销"
  },
  "publish": {
    "title": "误操作预收款怎么撤销",
    "hashtags": ["#抖音问答", "#知识分享"],
    "description": "适合三段式答疑短视频，适合财税、运营、知识讲解场景。"
  },
  "output": {
    "path": "./output/video/final.mp4",
    "platform": "douyin"
  }
}
```

说明：
- 不写的字段会自动补默认值。
- `aiEdit` 会优先把原素材裁成更适合短视频投放的钩子、主体、CTA 节奏，而不是平均切段。
- 不写 `timeline` 时，可以直接通过 `aiEdit.scriptLines` 生成 clip + subtitle 时间线。
- `music.path` 为空但 `music.style` 有值时，会按视频时长和人声音量自动选曲。
- 字幕滤镜不可用时，会自动输出 `output/subtitles/*.ass`。
- 抖音项目会自动生成 `output/report/*.publish.txt`。

## 示例文件

示例总览见 [examples/README.md](./examples/README.md)。

常用示例：
- [examples/douyin_qa_with_subtitles.json](./examples/douyin_qa_with_subtitles.json)
- [examples/douyin_goods_recommend.json](./examples/douyin_goods_recommend.json)
- [examples/douyin_ads_project.json](./examples/douyin_ads_project.json)
- [examples/image_mix_project.json](./examples/image_mix_project.json)

## Docker 运行

构建镜像：

```bash
docker build -t clawcut:local .
```

健康检查：

```bash
docker run --rm clawcut:local health
```

挂载当前工作目录并渲染：

```bash
docker run --rm \
  -u "$(id -u):$(id -g)" \
  -v "$PWD:/workspace" \
  clawcut:local render -project /workspace/projects/demo/project.json
```

拉取已发布镜像：

```bash
docker pull ghcr.io/shuangli441-ux/openclwa-cut:latest
docker run --rm ghcr.io/shuangli441-ux/openclwa-cut:latest health
```

## 常见问题

### 1. 提示找不到 FFmpeg

先执行：

```bash
clawcut health
```

如果显示 `ffmpeg` 或 `ffprobe` 缺失，优先使用 Docker 镜像运行，或者先安装 FFmpeg。

### 2. 路径里有中文、空格，渲染失败

新版本已经自动处理：
- 相对路径
- 中文路径
- 空格路径
- `~` 家目录路径
- 环境变量路径，例如 `$HOME/videos/input.mp4`

如果仍然报错，优先检查文件是否真实存在，再执行：

```bash
clawcut validate -project ./projects/demo/project.json
```

### 3. 字幕没有烧进视频

通常是当前 FFmpeg 缺少 `subtitles` 或 `drawtext` 滤镜。`clawcut` 会自动降级成外挂字幕，生成：

```bash
output/subtitles/*.ass
```

如果你必须烧录硬字幕，建议优先使用 Docker 镜像运行。

### 4. 为什么封面、字幕、报告是分目录的

这是为了工业化批量生产和二次分发：
- `output/video/` 放成片
- `output/cover/` 放封面
- `output/subtitles/` 放字幕
- `output/report/` 放渲染报告和发布文案

### 5. 为什么 BGM 没自动匹配到

先扫描音乐库：

```bash
clawcut music-scan -dir /path/to/music
```

再检查：

```bash
clawcut music-match -style qa-short
```

如果音乐库为空，或没有符合当前风格的曲目，就不会自动选曲。

### 6. 为什么 `go install` 之前装不上

旧版本的模块路径还是本地名，外部安装会失败。现在已经切到 GitHub 模块路径，可以直接执行：

```bash
go install github.com/shuangli441-ux/openclwa-cut/cmd/clawcut@latest
```
