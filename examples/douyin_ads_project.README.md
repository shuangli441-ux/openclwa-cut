# douyin_ads_project

输入：
- `./assets/input.mp4`
- `./assets/logo.png`
- 可选的本地音乐库，建议先执行 `clawcut music-scan -dir ~/Music/DouyinBGM`

说明：
- 这是投放短视频模板示例，时间线不需要手工填写，`aiEdit.scriptLines` 会自动生成 clip 和字幕。
- `watermarkPath` 是品牌 Logo 占位路径；如果暂时没有 Logo，可以删除 `branding` 整段配置。
- `music.style = "douyin-ads"` 会优先从已扫描的本地音乐库里匹配适合投放素材的 BGM。

输出：
- `./output/video/douyin_ads_project.mp4`
- `./output/cover/douyin_ads_project_cover.jpg`
- `./output/subtitles/douyin_ads_project.ass`
- `./output/report/douyin_ads_project.render.json`
- `./output/report/douyin_ads_project.publish.txt`

执行命令：

```bash
./clawcut render -project ./examples/douyin_ads_project.json
```

适合：
- 抖音信息流投放
- 本地推广短视频
- SaaS、工具类、服务类广告口播
