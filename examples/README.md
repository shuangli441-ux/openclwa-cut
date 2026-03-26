# 示例说明

这里放的是 `clawcut` 常用项目配置模板。你可以直接复制一份，替换自己的素材路径，然后执行 `clawcut render -project ...`。

注意：
- 示例里的 `./assets/...`、`./audio/...` 都是占位路径，需要换成你自己的素材。
- 抖音类示例默认会输出成片、封面、字幕和发布文案。
- 如果示例里只有 `aiEdit.scriptLines`、没有手工 `timeline`，说明这个项目会在渲染前自动生成剪辑时间线。

## douyin_qa_with_subtitles.json

文件：
- [douyin_qa_with_subtitles.json](./douyin_qa_with_subtitles.json)
- [douyin_qa_with_subtitles.README.md](./douyin_qa_with_subtitles.README.md)

适合：
- 问答口播
- 教程拆解
- 知识分享短视频

## douyin_goods_recommend.json

文件：
- [douyin_goods_recommend.json](./douyin_goods_recommend.json)
- [douyin_goods_recommend.README.md](./douyin_goods_recommend.README.md)

适合：
- 好物推荐
- 产品种草
- 口播带货

## douyin_ads_project.json

文件：
- [douyin_ads_project.json](./douyin_ads_project.json)
- [douyin_ads_project.README.md](./douyin_ads_project.README.md)

适合：
- 抖音投放短视频
- SaaS、工具、服务类广告
- 需要自动生成钩子、卖点、CTA 节奏的项目

## image_mix_project.json

文件：
- [image_mix_project.json](./image_mix_project.json)
- [image_mix_project.README.md](./image_mix_project.README.md)

适合：
- 封面图 + 视频正文混剪
- 静态图过场

## douyin_qa_project.json

文件：
- [douyin_qa_project.json](./douyin_qa_project.json)
- [douyin_qa_project.README.md](./douyin_qa_project.README.md)

适合：
- 纯 QA 节奏模板
- 口播问答项目最小示例

## douyin_template_project.json

文件：
- [douyin_template_project.json](./douyin_template_project.json)
- [douyin_template_project.README.md](./douyin_template_project.README.md)

适合：
- 模板初始化后继续二次修改
- QA 模板的手工扩展版本
