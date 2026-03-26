# douyin_goods_recommend

输入：
- `./assets/input.mp4`
- 音乐库里至少有 `goods-recommend` 风格的 BGM

说明：
- 视频路径是占位示例，使用前请替换成你自己的素材。
- `aiEdit.scriptLines` 会自动生成好物推荐节奏，不需要手工计算每段 clip 的时间。
- 如果没有手动指定 `music.path`，会优先按 `music.style` 自动选曲。

输出：
- `./output/video/douyin_goods_recommend.mp4`
- `./output/cover/douyin_goods_recommend_cover.jpg`
- `./output/subtitles/douyin_goods_recommend.ass`
- `./output/report/douyin_goods_recommend.render.json`
- `./output/report/douyin_goods_recommend.publish.txt`

执行命令：

```bash
./clawcut render -project ./examples/douyin_goods_recommend.json
```

适合：
- 好物推荐
- 产品种草
- 口播带货模板
