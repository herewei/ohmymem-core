package detector

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/herewei/ohmymem-core/internal/domain"
)

// GoDetector Go项目检测器
type GoDetector struct{}

// NewGoDetector 创建Go检测器
func NewGoDetector() *GoDetector {
	return &GoDetector{}
}

func (d *GoDetector) Name() string {
	return "go"
}

func (d *GoDetector) Priority() int {
	return 10
}

// Detect 检测Go项目
func (d *GoDetector) Detect(rootPath string) (*domain.ProjectInfo, error) {
	goModPath := filepath.Join(rootPath, "go.mod")

	if !fileExists(goModPath) {
		return &domain.ProjectInfo{Language: ""}, nil
	}

	info := &domain.ProjectInfo{
		Language: "go",
		RootPath: rootPath,
		Features: []string{},
	}

	// 读取go.mod分析依赖
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return info, nil
	}

	goModContent := string(content)

	// 检测框架
	info.Framework = d.detectFramework(goModContent)

	// 检测数据库
	info.Database = d.detectDatabase(goModContent)

	// 检测项目类型
	info.ProjectType = d.detectProjectType(rootPath, goModContent)

	// 检测特性
	info.Features = d.detectFeatures(goModContent)

	return info, nil
}

// detectFramework 检测Web框架
func (d *GoDetector) detectFramework(goMod string) string {
	frameworks := map[string]string{
		"github.com/labstack/echo": "echo",
		"github.com/gin-gonic/gin": "gin",
		"github.com/gofiber/fiber": "fiber",
		"github.com/gorilla/mux":   "gorilla",
		"github.com/go-chi/chi":    "chi",
	}

	for pkg, name := range frameworks {
		if strings.Contains(goMod, pkg) {
			return name
		}
	}

	return ""
}

// detectDatabase 检测数据库
func (d *GoDetector) detectDatabase(goMod string) string {
	databases := map[string]string{
		"github.com/jackc/pgx":           "postgresql",
		"github.com/lib/pq":              "postgresql",
		"github.com/go-sql-driver/mysql": "mysql",
		"go.mongodb.org/mongo-driver":    "mongodb",
		"github.com/go-redis/redis":      "redis",
		"github.com/redis/go-redis":      "redis",
	}

	for pkg, name := range databases {
		if strings.Contains(goMod, pkg) {
			return name
		}
	}

	return ""
}

// detectProjectType 检测项目类型
func (d *GoDetector) detectProjectType(rootPath, goMod string) string {
	// 有cmd目录通常是CLI或服务
	if dirExists(filepath.Join(rootPath, "cmd")) {
		// 检查是否有web相关包
		if strings.Contains(goMod, "echo") ||
			strings.Contains(goMod, "gin") ||
			strings.Contains(goMod, "fiber") ||
			strings.Contains(goMod, "net/http") {
			return "backend"
		}
		return "cli"
	}

	// 有internal但没有cmd，可能是库
	if dirExists(filepath.Join(rootPath, "internal")) {
		return "backend"
	}

	return "library"
}

// detectFeatures 检测项目特性
func (d *GoDetector) detectFeatures(goMod string) []string {
	var features []string

	featureMap := map[string]string{
		"google.golang.org/grpc":       "grpc",
		"github.com/grpc-ecosystem":    "grpc",
		"github.com/swaggo/swag":       "swagger",
		"github.com/golang-jwt/jwt":    "jwt",
		"github.com/prometheus/client": "prometheus",
		"go.opentelemetry.io/otel":     "opentelemetry",
	}

	for pkg, feature := range featureMap {
		if strings.Contains(goMod, pkg) {
			features = append(features, feature)
		}
	}

	return features
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
