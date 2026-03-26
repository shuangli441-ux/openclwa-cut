# image_mix_project

输入：
- `./assets/cover.png`
- `./assets/input.mp4`

说明：
- 两个路径都是占位示例，渲染前请替换成你自己的图片和视频素材。

输出：
- `./output/video/image_mix_render.mp4`
- `./output/cover/image_mix_render_cover.jpg`
- `./output/subtitles/image_mix_render.ass`
- `./output/report/image_mix_render.render.json`
- `./output/report/image_mix_render.publish.txt`

执行命令：

```bash
./clawcut render -project ./examples/image_mix_project.json
```

适合：
- 封面图和视频素材混剪
- 静态图过场视频
