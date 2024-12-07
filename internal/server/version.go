package server

import (
	"fmt"
	"strings"
)

type APIVersion string

const (
	V1 APIVersion = "v1"
	V2 APIVersion = "v2"
)

// VersionManager 版本管理器
type VersionManager struct {
	versions       map[APIVersion]bool
	defaultVersion APIVersion
	versionParam   string // URL参数中的版本标识
}

// NewVersionManager 创建版本管理器
func NewVersionManager() *VersionManager {
	vm := &VersionManager{
		versions:       make(map[APIVersion]bool),
		defaultVersion: V1,
		versionParam:   "version",
	}
	vm.RegisterVersion(V1)
	return vm
}

// RegisterVersion 注册版本
func (vm *VersionManager) RegisterVersion(version APIVersion) {
	vm.versions[version] = true
}

// SetDefaultVersion 设置默认版本
func (vm *VersionManager) SetDefaultVersion(version APIVersion) {
	if vm.IsValidVersion(version) {
		vm.defaultVersion = version
	}
}

// IsValidVersion 判断版本是否有效
func (vm *VersionManager) IsValidVersion(version APIVersion) bool {
	return vm.versions[version]
}

// GetVersionPath 获取带版本的路径
func (vm *VersionManager) GetVersionPath(version APIVersion, path string) string {
	if !vm.IsValidVersion(version) {
		version = vm.defaultVersion
	}
	return fmt.Sprintf("/api/%s%s", version, path)
}

// ParseVersionFromPath 从路径中解析版本
func (vm *VersionManager) ParseVersionFromPath(path string) APIVersion {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "api" && i+1 < len(parts) {
			if vm.IsValidVersion(APIVersion(parts[i+1])) {
				return APIVersion(parts[i+1])
			}
		}
	}
	return vm.defaultVersion
}

// GetAvailableVersions 获取所有可用版本
func (vm *VersionManager) GetAvailableVersions() []APIVersion {
	versions := make([]APIVersion, 0, len(vm.versions))
	for version := range vm.versions {
		versions = append(versions, version)
	}
	return versions
}
