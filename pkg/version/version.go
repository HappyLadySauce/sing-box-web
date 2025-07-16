// Package version provides version information for sing-box-web.
package version

import (
	"fmt"
	"runtime"
	"time"
)

// 版本信息变量，在构建时通过 ldflags 注入
var (
	// Version 应用版本号
	Version = "dev"
	
	// BuildTime 构建时间
	BuildTime = "unknown"
	
	// GitCommit Git提交哈希
	GitCommit = "unknown"
	
	// GitBranch Git分支
	GitBranch = "unknown"
)

// Info 包含完整的版本信息
type Info struct {
	Version   string `json:"version"`
	BuildTime string `json:"build_time"`
	GitCommit string `json:"git_commit"`
	GitBranch string `json:"git_branch"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
	Arch      string `json:"arch"`
}

// Get 返回完整的版本信息
func Get() *Info {
	return &Info{
		Version:   Version,
		BuildTime: BuildTime,
		GitCommit: GitCommit,
		GitBranch: GitBranch,
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

// String 返回版本信息的字符串表示
func (i *Info) String() string {
	return fmt.Sprintf("Version: %s\nBuild Time: %s\nGit Commit: %s\nGit Branch: %s\nGo Version: %s\nPlatform: %s/%s",
		i.Version, i.BuildTime, i.GitCommit, i.GitBranch, i.GoVersion, i.Platform, i.Arch)
}

// Short 返回简短的版本信息
func (i *Info) Short() string {
	return fmt.Sprintf("%s (%s)", i.Version, i.GitCommit[:8])
}

// GetVersion 返回版本号
func GetVersion() string {
	return Version
}

// GetBuildTime 返回构建时间
func GetBuildTime() time.Time {
	if BuildTime == "unknown" {
		return time.Time{}
	}
	
	// 尝试解析构建时间
	if t, err := time.Parse("2006-01-02_15:04:05", BuildTime); err == nil {
		return t
	}
	
	return time.Time{}
}

// IsRelease 判断是否为正式版本
func IsRelease() bool {
	return Version != "dev" && GitBranch != "unknown"
}

// GetUserAgent 返回用户代理字符串
func GetUserAgent() string {
	return fmt.Sprintf("sing-box-web/%s (%s/%s)", Version, runtime.GOOS, runtime.GOARCH)
} 