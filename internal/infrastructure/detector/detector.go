package detector

import (
	"github.com/herewei/ohmymem-core/internal/domain"
)

// CompositeDetector 组合检测器
type CompositeDetector struct {
	detectors []LanguageDetector
}

// LanguageDetector 语言检测器接口
type LanguageDetector interface {
	Name() string
	Detect(rootPath string) (*domain.ProjectInfo, error)
	Priority() int // 优先级，数字越小优先级越高
}

// NewCompositeDetector 创建组合检测器
func NewCompositeDetector() *CompositeDetector {
	return &CompositeDetector{
		detectors: []LanguageDetector{
			NewGoDetector(),
			// 未来添加更多：NewTypeScriptDetector(), NewPythonDetector()
		},
	}
}

// Detect 检测项目
func (d *CompositeDetector) Detect(rootPath string) (*domain.ProjectInfo, error) {
	for _, detector := range d.detectors {
		info, err := detector.Detect(rootPath)
		if err != nil {
			continue
		}
		if info.IsDetected() {
			return info, nil
		}
	}

	// 未检测到
	return &domain.ProjectInfo{
		Language: "unknown",
		RootPath: rootPath,
	}, nil
}
