请同步优化以下文档：
1. README.md
   - 开头放 Slogan：「爪切 (clawcut) —— 一行命令，从素材到抖音精品成片」
   - 核心命令按功能分组展示（环境/剪辑/项目/音乐/模板）
   - 加「5 分钟快速开始」步骤：安装 → 初始化 → 渲染 → 看结果
   - 补充「常见问题」板块：比如 FFmpeg 未安装、路径报错等

2. 示例文件 (examples/)
   - 完善 douyin_qa_with_subtitles.json：包含视频、字幕、BGM 完整配置
   - 新增示例：douyin_goods_recommend.json（好物推荐模板）
   - 每个示例都加 README，说明输入、输出、执行命令

3. 代码注释
   - 所有公开函数、结构体加中文注释
   - 关键逻辑（比如路径解析、音频 ducking）加详细说明
