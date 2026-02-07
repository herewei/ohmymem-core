package domain

// ProjectInfo 检测到的项目信息
type ProjectInfo struct {
	Language    string   // go, typescript, python, rust, unknown
	Framework   string   // echo, gin, express, fastapi, etc. 空字符串表示未检测到
	ProjectType string   // backend, frontend, cli, library
	Database    string   // postgresql, mysql, mongodb, etc.
	Features    []string // 检测到的特性
	RootPath    string   // 项目根目录
}

// IsDetected 是否成功检测到语言
func (p *ProjectInfo) IsDetected() bool {
	return p.Language != "" && p.Language != "unknown"
}

// ProjectDetector 项目检测器接口（Port）
type ProjectDetector interface {
	// 检测项目信息
	Detect(rootPath string) (*ProjectInfo, error)
}
