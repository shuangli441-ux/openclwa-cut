# clawcut

Go 版本地强剪辑工具。

当前命令：
- `clawcut health`
- `clawcut trim`
- `clawcut compress`
- `clawcut concat`
- `clawcut project-init`
- `clawcut validate`
- `clawcut render`
- `clawcut render-dir`
- `clawcut music-init`
- `clawcut music-match`
- `clawcut music-mix`
- `clawcut template-douyin-qa`

现在这版已经支持：
- 统一项目 JSON 配置
- 相对路径资产解析
- 分辨率 / 帧率统一渲染
- 自动 scale + pad 到目标画布
- 样式化字幕输出
- 图片素材片段
- 更稳定的重编码拼接
- 背景音乐选择、混音和人声 ducking
- 项目预校验
- 批量渲染目录内项目

## 快速开始

```bash
cd /Users/apple/.openclaw/workspace/tools/clawcut
go build -o clawcut ./cmd/clawcut
./clawcut health
./clawcut project-init -dir ./projects/demo -name demo
./clawcut validate -project ./projects/demo/project.json
./clawcut render -project ./examples/douyin_qa_with_subtitles.json
./clawcut render-dir -dir ./projects -pattern '*.json'
```

## 项目结构

```json
{
  "project": "douyin-qa-subtitles",
  "format": "douyin-vertical",
  "fps": 30,
  "resolution": "1080x1920",
  "assets": [
    {
      "id": "clip1",
      "type": "video",
      "path": "./assets/input.mp4"
    }
  ],
  "timeline": [
    { "type": "clip", "asset": "clip1", "start": 0, "end": 3 },
    { "type": "subtitle", "start": 0, "end": 3, "text": "第一句字幕" }
  ],
  "subtitles": {
    "fontName": "PingFang SC",
    "fontSize": 60,
    "primaryColor": "#FFFFFF",
    "outlineColor": "#000000",
    "outline": 3,
    "marginV": 160,
    "maxCharsPerLine": 18
  },
  "music": {
    "path": "./audio/bgm.m4a",
    "volume": 0.14,
    "fadeOutSeconds": 1.2,
    "duckingThreshold": 0.035,
    "duckingRatio": 10,
    "duckingAttackMs": 20,
    "duckingReleaseMs": 350,
    "voiceBoost": 1.0
  },
  "output": {
    "path": "./output/final.mp4",
    "platform": "douyin",
    "videoCodec": "libx264",
    "audioCodec": "aac",
    "audioBitrate": "160k",
    "preset": "medium",
    "crf": 21
  }
}
```

## 渲染规则

- `clip` 的 `start/end` 是源素材时间。
- 当 `clip` 引用的是 `image` 资产时，`end-start` 会被当作该图片片段时长。
- `subtitle` 的 `start/end` 是最终成片时间。
- 所有视频片段会先统一转到目标分辨率、帧率和编码参数，再拼接。
- 若当前 FFmpeg 支持 `subtitles` 或 `drawtext` 滤镜，会直接烧录硬字幕；否则会自动输出同名 `.ass` 字幕文件。
- 若配置了 `music.path` 或 `music.style`，渲染结束后自动混入 BGM。
- 当原视频里存在人声时，默认启用 sidechain ducking；BGM 会在说话时自动下沉，尽量不盖住人声。

## 音乐库

可以用 `music.style` + `music.library` 从音乐库里挑选 BGM。

```bash
./clawcut music-match -path ./config/music_library.json -style qa-short
```
