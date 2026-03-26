# douyin_qa_with_subtitles

输入：
- `./assets/input.mp4`
- `./audio/bgm.m4a`

说明：
- 这两个路径是占位示例，使用前请替换成你自己的素材。
- 如果本机 FFmpeg 不支持硬字幕，渲染时会自动降级输出 `.ass` 外挂字幕。

输出：
- `./output/video/douyin_qa_with_subtitles.mp4`
- `./output/cover/douyin_qa_with_subtitles_cover.jpg`
- `./output/subtitles/douyin_qa_with_subtitles.ass`
- `./output/report/douyin_qa_with_subtitles.render.json`
- `./output/report/douyin_qa_with_subtitles.publish.txt`

执行命令：

```bash
./clawcut render -project ./examples/douyin_qa_with_subtitles.json
```

适合：
- 抖音问答
- 教程说明
- 口播知识视频
