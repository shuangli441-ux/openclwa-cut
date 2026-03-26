package cli

import "fmt"

// RenderProgress 以阶段为单位输出渲染进度。
type RenderProgress struct {
	total int
	step  int
}

// NewRenderProgress 创建一个简单的命令行进度提示器。
func NewRenderProgress(total int) *RenderProgress {
	if total <= 0 {
		total = 1
	}
	return &RenderProgress{total: total}
}

// Step 推进一步并输出当前阶段说明。
func (p *RenderProgress) Step(message string) {
	p.step++
	if p.step > p.total {
		p.total = p.step
	}
	fmt.Printf("[%.0f%%] (%d/%d) %s\n", float64(p.step)*100/float64(p.total), p.step, p.total, message)
}
