package main

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"strconv"
	"strings"
)

// deriveSessionKey 返回用于存储的session标识。
// 1) 若客户端已提供SessionID，则直接使用；2) 否则基于主机指纹与请求来源生成可重复的ID，
// 以避免仅用hostname导致的冲突。
func deriveSessionKey(info *SystemInfo, projectKey, remoteAddr string) string {
	if info == nil {
		return ""
	}
	if info.SessionID != "" {
		return info.SessionID
	}

	parts := []string{
		strings.TrimSpace(projectKey),
		strings.TrimSpace(info.Hostname),
		strings.TrimSpace(info.OS.Platform),
		strings.TrimSpace(info.OS.Architecture),
		strings.TrimSpace(info.OS.Version),
		strconv.Itoa(info.CPU.CoreCount),
		strconv.FormatUint(info.Memory.Total, 10),
		strconv.FormatUint(info.Disk.Total, 10),
	}

	if remoteAddr != "" {
		if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
			parts = append(parts, host)
		} else {
			parts = append(parts, remoteAddr)
		}
	}

	fingerprint := strings.Join(parts, "|")
	sum := sha256.Sum256([]byte(fingerprint))

	return "auto-" + hex.EncodeToString(sum[:])[:16]
}
