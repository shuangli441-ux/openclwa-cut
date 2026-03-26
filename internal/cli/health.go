package cli

import (
	"fmt"
	"strings"

	"github.com/shuangli441-ux/openclwa-cut/internal/ffmpeg"
)

type healthItem struct {
	Name    string
	OK      bool
	Detail  string
	Warning bool
}

// RunHealthCheck 检查 FFmpeg、滤镜和字体环境是否满足渲染要求。
func RunHealthCheck() error {
	items := make([]healthItem, 0, 8)

	ffmpegVersion, err := binaryVersion("ffmpeg")
	if err != nil {
		return fmt.Errorf("未检测到 ffmpeg，请先安装后再运行 clawcut：%w", err)
	}
	items = append(items, healthItem{Name: "ffmpeg", OK: true, Detail: ffmpegVersion})

	ffprobeVersion, err := binaryVersion("ffprobe")
	if err != nil {
		return fmt.Errorf("未检测到 ffprobe，请先安装后再运行 clawcut：%w", err)
	}
	items = append(items, healthItem{Name: "ffprobe", OK: true, Detail: ffprobeVersion})

	for _, filterName := range []string{"subtitles", "drawtext", "sidechaincompress", "overlay", "highpass", "lowpass", "acompressor", "alimiter"} {
		ok, err := ffmpeg.HasFilter(filterName)
		if err != nil {
			items = append(items, healthItem{Name: "filter:" + filterName, OK: false, Detail: err.Error(), Warning: true})
			continue
		}
		items = append(items, healthItem{
			Name:    "filter:" + filterName,
			OK:      ok,
			Detail:  filterStatusDetail(filterName, ok),
			Warning: !ok,
		})
	}

	fontPath := DefaultSubtitleFontFile()
	if fontPath == "" {
		items = append(items, healthItem{Name: "font", OK: false, Detail: "未找到默认字幕字体，将使用 FFmpeg 自带字体降级", Warning: true})
	} else {
		items = append(items, healthItem{Name: "font", OK: true, Detail: fontPath})
	}

	fmt.Println("clawcut 环境检查")
	for _, item := range items {
		status := "OK"
		if !item.OK {
			if item.Warning {
				status = "WARN"
			} else {
				status = "FAIL"
			}
		}
		fmt.Printf("- [%s] %s: %s\n", status, item.Name, item.Detail)
	}

	fmt.Println("建议")
	fmt.Println("- 推荐优先使用 Docker 镜像运行，避免本地 FFmpeg 功能不完整")
	fmt.Println("- 如果字幕不能烧录，clawcut 会自动降级为外挂字幕，不会中断成片输出")
	return nil
}

func binaryVersion(name string) (string, error) {
	out, err := ffmpeg.RunCapture(name, "-version")
	if err != nil {
		return "", err
	}
	line := strings.TrimSpace(strings.SplitN(string(out), "\n", 2)[0])
	if line == "" {
		return "已安装", nil
	}
	return line, nil
}

func filterStatusDetail(filterName string, ok bool) string {
	if ok {
		switch filterName {
		case "subtitles":
			return "支持硬字幕烧录"
		case "drawtext":
			return "支持封面标题和文案叠字"
		case "sidechaincompress":
			return "支持人声出现时自动压低 BGM"
		case "overlay":
			return "支持品牌水印和 Logo 叠加"
		case "highpass", "lowpass", "acompressor", "alimiter":
			return "支持人声增强音频链"
		default:
			return "可用"
		}
	}
	switch filterName {
	case "subtitles":
		return "不支持硬字幕烧录，将自动输出 .ass 字幕文件"
	case "drawtext":
		return "不支持 drawtext，封面大字会自动降级"
	case "sidechaincompress":
		return "不支持 sidechaincompress，将退回普通混音"
	case "overlay":
		return "不支持 overlay，水印功能会不可用"
	default:
		return "该能力缺失，将自动降级"
	}
}
