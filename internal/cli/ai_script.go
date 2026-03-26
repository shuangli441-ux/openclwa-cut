package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/shuangli441-ux/openclwa-cut/internal/ffmpeg"
)

// AIScriptOptions 描述 AI 脚本生成命令的可选覆盖项。
type AIScriptOptions struct {
	Provider  string
	Model     string
	Force     bool
	PrintOnly bool
}

type codexScriptResponse struct {
	ScriptLines []string `json:"scriptLines"`
}

// GenerateAIScriptProject 调用 AI 提供方生成脚本文案，并回写项目配置。
func GenerateAIScriptProject(projectPath string, opts AIScriptOptions) error {
	rawProject, workingProject, baseDir, err := loadProjectMutationPair(projectPath)
	if err != nil {
		return err
	}

	if strings.TrimSpace(opts.Provider) != "" {
		rawProject.AIEdit.Provider = strings.TrimSpace(opts.Provider)
		workingProject.AIEdit.Provider = strings.TrimSpace(opts.Provider)
	}
	if strings.TrimSpace(opts.Model) != "" {
		rawProject.AIEdit.Model = strings.TrimSpace(opts.Model)
		workingProject.AIEdit.Model = strings.TrimSpace(opts.Model)
	}
	if workingProject.ResolvedAIProvider() == AIProviderBuiltin {
		rawProject.AIEdit.Provider = AIProviderCodex
		workingProject.AIEdit.Provider = AIProviderCodex
	}
	if !opts.Force && len(workingProject.ResolvedAIEditScriptLines()) > 0 {
		return fmt.Errorf("项目已存在 aiEdit.scriptLines，如需重新生成请追加 -force")
	}

	scriptLines, err := generateAIScriptLines(&workingProject, baseDir)
	if err != nil {
		return err
	}

	kind := workingProject.ResolveAIEditTemplateKind()
	duration, assetID, err := aiTimelineSource(&workingProject)
	if err != nil {
		return err
	}
	timeline := BuildSmartTemplateTimeline(duration, assetID, scriptLines, kind, workingProject.AIEdit)
	if len(timeline) == 0 {
		return fmt.Errorf("AI 脚本已生成，但未能构建时间线，请检查素材和项目配置")
	}

	rawProject.AIEdit.Provider = workingProject.ResolvedAIProvider()
	rawProject.AIEdit.Command = strings.TrimSpace(workingProject.AIEdit.Command)
	rawProject.AIEdit.Model = strings.TrimSpace(workingProject.AIEdit.Model)
	rawProject.AIEdit.PromptHint = strings.TrimSpace(workingProject.AIEdit.PromptHint)
	rawProject.AIEdit.AutoGenerate = workingProject.AIEdit.AutoGenerate
	rawProject.AIEdit.Enabled = workingProject.AIEdit.Enabled
	rawProject.AIEdit.Mode = workingProject.AIEdit.Mode
	rawProject.AIEdit.TemplateKind = kind
	rawProject.AIEdit.MaxDurationSeconds = workingProject.AIEdit.MaxDurationSeconds
	rawProject.AIEdit.HookSeconds = workingProject.AIEdit.HookSeconds
	rawProject.AIEdit.CTASeconds = workingProject.AIEdit.CTASeconds
	rawProject.AIEdit.ScriptLines = scriptLines
	rawProject.Timeline = timeline

	if opts.PrintOnly {
		for i, line := range scriptLines {
			fmt.Printf("%d. %s\n", i+1, line)
		}
		return nil
	}
	return writeProjectJSON(projectPath, rawProject)
}

func loadProjectMutationPair(projectPath string) (Project, Project, string, error) {
	data, err := os.ReadFile(projectPath)
	if err != nil {
		return Project{}, Project{}, "", fmt.Errorf("读取项目文件失败：%w", err)
	}

	var rawProject Project
	if err := json.Unmarshal(data, &rawProject); err != nil {
		return Project{}, Project{}, "", fmt.Errorf("解析项目 JSON 失败：%w", err)
	}
	workingProject := rawProject
	workingProject.ApplyDefaults()
	baseDir := filepath.Dir(projectPath)
	workingProject.ResolvePaths(baseDir)
	return rawProject, workingProject, baseDir, nil
}

func generateAIScriptLines(project *Project, workDir string) ([]string, error) {
	if project == nil {
		return nil, fmt.Errorf("项目为空，无法生成 AI 脚本")
	}
	if len(project.Assets) == 0 {
		return nil, fmt.Errorf("AI 脚本生成需要至少一个素材文件")
	}

	switch project.ResolvedAIProvider() {
	case AIProviderCodex:
		return generateAIScriptLinesWithCodex(project, workDir)
	case AIProviderBuiltin:
		return nil, fmt.Errorf("当前项目使用的是内置模板脚本，请把 aiEdit.provider 改成 codex，或使用 -provider codex")
	default:
		return nil, fmt.Errorf("AI 提供方不支持：%s", project.ResolvedAIProvider())
	}
}

func generateAIScriptLinesWithCodex(project *Project, workDir string) ([]string, error) {
	_, asset, ok := project.primaryAIAsset()
	if !ok {
		return nil, fmt.Errorf("AI 脚本生成需要至少一个主视频素材")
	}
	duration, _, err := aiTimelineSource(project)
	if err != nil {
		return nil, err
	}
	prompt := buildCodexScriptPrompt(project, asset, duration)
	return runCodexScriptGenerator(project.ResolvedAICommand(), workDir, strings.TrimSpace(project.AIEdit.Model), prompt)
}

func aiTimelineSource(project *Project) (float64, string, error) {
	assetID, asset, ok := project.primaryAIAsset()
	if !ok {
		return 0, "", fmt.Errorf("AI 时间线生成需要至少一个视频素材")
	}
	duration := project.TotalDuration()
	if probed, err := ffmpeg.ProbeDuration(asset.Path); err == nil && probed > 0 {
		duration = probed
	}
	if duration <= 0 {
		duration = project.AIEdit.MaxDurationSeconds
	}
	if duration <= 0 {
		duration = float64(len(project.ResolvedAIEditScriptLines())) * 3
	}
	if duration <= 0 {
		return 0, "", fmt.Errorf("无法确定主素材时长，请检查视频文件是否可读取")
	}
	return duration, assetID, nil
}

func buildCodexScriptPrompt(project *Project, asset Asset, duration float64) string {
	kind := project.ResolveAIEditTemplateKind()
	expectedCount := expectedScriptLineCount(kind)
	title := project.ResolvedPublishTitle()
	if title == "" {
		title = strings.TrimSpace(project.Project)
	}

	lines := []string{
		"你是工业化短视频脚本策划，请为抖音成片生成可直接用作字幕和镜头节奏的口播文案。",
		"只输出一个 JSON 对象，不要输出解释、Markdown 或代码块。",
		fmt.Sprintf("模板类型：%s。", kind),
		fmt.Sprintf("目标句数：%d 句。", expectedCount),
		fmt.Sprintf("单句建议：控制在 %d 个中文字符以内，适合竖屏字幕。", maxInt(project.Subtitles.MaxCharsPerLine, 14)),
		fmt.Sprintf("素材文件：%s。", filepath.Base(asset.Path)),
		fmt.Sprintf("素材时长：%.1f 秒。", duration),
	}
	if title != "" {
		lines = append(lines, "主题标题："+title+"。")
	}
	if desc := strings.TrimSpace(project.Publish.Description); desc != "" {
		lines = append(lines, "项目说明："+desc+"。")
	}
	if hashtags := project.ResolvedPublishHashtags(); len(hashtags) > 0 {
		lines = append(lines, "参考话题："+strings.Join(hashtags, " ")+"。")
	}
	if hint := strings.TrimSpace(project.AIEdit.PromptHint); hint != "" {
		lines = append(lines, "额外要求："+hint+"。")
	}
	if existing := normalizeScriptLines(project.AIEdit.ScriptLines); len(existing) > 0 {
		lines = append(lines, "已有草稿："+strings.Join(existing, " / ")+"。请重写成更适合抖音成片的版本。")
	}

	switch kind {
	case TemplateDouyinAds:
		lines = append(lines,
			"结构要求：第 1 句是钩子，第 2 句讲痛点或场景，第 3 句讲方案或收益，最后 1 句是 CTA。",
			"文案要求：更像真实投放口播，不要空话，不要抽象术语，不要写镜头说明。",
		)
	case TemplateDouyinGoods:
		lines = append(lines,
			"结构要求：先讲使用场景，再讲核心卖点，最后给购买理由或行动建议。",
			"文案要求：适合好物推荐和种草，不要堆参数。",
		)
	default:
		lines = append(lines,
			"结构要求：先抛问题，中段给结论或步骤，结尾补提醒或引导收藏。",
			"文案要求：适合教程、答疑、知识类口播。",
		)
	}

	lines = append(lines, `返回格式示例：{"scriptLines":["第一句","第二句","第三句"]}`)
	return strings.Join(lines, "\n")
}

func runCodexScriptGenerator(commandPath, workDir, model, prompt string) ([]string, error) {
	if strings.TrimSpace(commandPath) == "" {
		commandPath = "codex"
	}
	if _, err := exec.LookPath(commandPath); err != nil {
		return nil, fmt.Errorf("未找到本机 Codex CLI，请先安装 codex，或在 aiEdit.command 指定命令路径：%w", err)
	}

	tempDir, err := os.MkdirTemp("", "clawcut-codex-")
	if err != nil {
		return nil, fmt.Errorf("创建 Codex 临时目录失败：%w", err)
	}
	defer os.RemoveAll(tempDir)

	outputFile := filepath.Join(tempDir, "codex-output.json")
	schemaFile := filepath.Join(tempDir, "schema.json")
	if err := os.WriteFile(schemaFile, []byte(codexScriptOutputSchema), 0644); err != nil {
		return nil, fmt.Errorf("写入 Codex 输出约束失败：%w", err)
	}

	args := []string{
		"exec",
		"--skip-git-repo-check",
		"--ephemeral",
		"-C", workDir,
		"-o", outputFile,
		"--output-schema", schemaFile,
		"-",
	}
	if strings.TrimSpace(model) != "" {
		args = []string{
			"exec",
			"--skip-git-repo-check",
			"--ephemeral",
			"-C", workDir,
			"-o", outputFile,
			"--output-schema", schemaFile,
			"-m", model,
			"-",
		}
	}

	cmd := exec.Command(commandPath, args...)
	cmd.Stdin = strings.NewReader(prompt)
	output, err := cmd.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			return nil, fmt.Errorf("调用本机 Codex 失败：%w", err)
		}
		return nil, fmt.Errorf("调用本机 Codex 失败：%v\n%s", err, message)
	}

	data, err := os.ReadFile(outputFile)
	if err != nil {
		return nil, fmt.Errorf("读取 Codex 输出失败：%w", err)
	}
	lines, err := parseCodexScriptResponse(data)
	if err != nil {
		return nil, err
	}
	return lines, nil
}

func parseCodexScriptResponse(data []byte) ([]string, error) {
	content := strings.TrimSpace(string(data))
	if content == "" {
		return nil, fmt.Errorf("Codex 没有返回脚本文案")
	}
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var response codexScriptResponse
	if err := json.Unmarshal([]byte(content), &response); err != nil {
		start := strings.Index(content, "{")
		end := strings.LastIndex(content, "}")
		if start < 0 || end <= start {
			return nil, fmt.Errorf("Codex 返回内容无法解析为 JSON：%s", content)
		}
		if err := json.Unmarshal([]byte(content[start:end+1]), &response); err != nil {
			return nil, fmt.Errorf("Codex 返回内容无法解析为 JSON：%w", err)
		}
	}

	response.ScriptLines = normalizeScriptLines(response.ScriptLines)
	if len(response.ScriptLines) == 0 {
		return nil, fmt.Errorf("Codex 没有返回有效的 scriptLines")
	}
	return response.ScriptLines, nil
}

func expectedScriptLineCount(kind string) int {
	switch normalizeTemplateKind(kind) {
	case TemplateDouyinAds:
		return 4
	case TemplateDouyinGoods:
		return 3
	default:
		return 3
	}
}

func maxInt(value int, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}

const codexScriptOutputSchema = `{
  "type": "object",
  "additionalProperties": false,
  "properties": {
    "scriptLines": {
      "type": "array",
      "minItems": 3,
      "maxItems": 6,
      "items": {
        "type": "string",
        "minLength": 2,
        "maxLength": 40
      }
    }
  },
  "required": ["scriptLines"]
}`
