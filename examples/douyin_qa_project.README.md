# douyin_qa_project

输入：
- 一个口播主视频
- 可选的本地音乐库

说明：
- 请先把 `douyin_qa_project.json` 里的 `./assets/input.mp4` 替换成自己的素材路径。

输出：
- `./output/video/example_render.mp4`
- `./output/cover/example_render_cover.jpg`
- `./output/report/example_render.render.json`
- `./output/report/example_render.publish.txt`

执行命令：

```bash
./clawcut render -project ./examples/douyin_qa_project.json
```

适合：
- 问答类口播模板
- 快速起号内容
