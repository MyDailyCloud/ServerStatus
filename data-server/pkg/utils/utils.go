package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// GenerateUUID 生成UUID
func GenerateUUID() string {
	return uuid.New().String()
}

// GenerateSessionID 生成会话ID
func GenerateSessionID() string {
	return GenerateUUID()
}

// GenerateAccessKey 生成访问密钥
func GenerateAccessKey(serverKey, projectKey string) string {
	combinedKey := serverKey + projectKey + time.Now().Format("20060102")
	hash := sha256.Sum256([]byte(combinedKey))
	return hex.EncodeToString(hash[:])
}

// GenerateRandomString 生成随机字符串
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		// 如果随机数生成失败，使用时间戳作为后备方案
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}
	return string(bytes)
}

// IsValidHost 验证主机地址格式
func IsValidHost(host string) bool {
	if host == "" {
		return false
	}

	// IPv4地址
	if net.ParseIP(host) != nil {
		return true
	}

	// 特殊主机名
	if host == "localhost" || host == "0.0.0.0" {
		return true
	}

	// 域名格式检查
	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	return domainRegex.MatchString(host)
}

// IsValidPort 验证端口号格式
func IsValidPort(port string) bool {
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return false
	}
	return portNum >= 1 && portNum <= 65535
}

// IsValidRedisAddr 验证Redis地址格式
func IsValidRedisAddr(addr string) bool {
	// 支持 host:port 格式
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return false
	}

	host := parts[0]
	port := parts[1]

	if !IsValidHost(host) && host != "" {
		return false
	}

	return IsValidPort(port)
}

// IsValidPath 验证文件路径格式
func IsValidPath(path string) bool {
	// 基本路径检查
	if path == "" {
		return false
	}

	// 检查是否包含非法字符
	illegalChars := []string{"\x00", "\x01", "\x02", "\x03", "\x04", "\x05", "\x06", "\x07", "\x08", "\x09", "\x0a", "\x0b", "\x0c", "\x0d", "\x0e", "\x0f"}
	for _, char := range illegalChars {
		if strings.Contains(path, char) {
			return false
		}
	}

	// 检查路径长度
	if len(path) > 260 { // Windows路径长度限制
		return false
	}

	return true
}

// EnsureDir 确保目录存在
func EnsureDir(dirPath string) error {
	return os.MkdirAll(dirPath, 0755)
}

// FileExists 检查文件是否存在
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// DirExists 检查目录是否存在
func DirExists(dirPath string) bool {
	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// GetFileSize 获取文件大小
func GetFileSize(filePath string) (int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// RemoveFile 删除文件
func RemoveFile(filePath string) error {
	return os.Remove(filePath)
}

// CopyFile 复制文件
func CopyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = destination.ReadFrom(source)
	return err
}

// GetHostname 获取主机名
func GetHostname() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown", err
	}
	return hostname, nil
}

// FormatBytes 格式化字节数
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatDuration 格式化时间间隔
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}

// ParseDuration 解析时间间隔字符串
func ParseDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}

// StringToBool 将字符串转换为布尔值
func StringToBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "1" || s == "yes" || s == "on"
}

// BoolToString 将布尔值转换为字符串
func BoolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// StringToInt 将字符串转换为整数
func StringToInt(s string, defaultValue int) int {
	if i, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
		return i
	}
	return defaultValue
}

// StringToFloat 将字符串转换为浮点数
func StringToFloat(s string, defaultValue float64) float64 {
	if f, err := strconv.ParseFloat(strings.TrimSpace(s), 64); err == nil {
		return f
	}
	return defaultValue
}

// TruncateString 截断字符串
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// SanitizeString 清理字符串
func SanitizeString(s string) string {
	// 移除控制字符
	re := regexp.MustCompile(`[\x00-\x1f\x7f]`)
	return re.ReplaceAllString(s, "")
}

// Contains 检查字符串切片是否包含某个元素
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// RemoveFromSlice 从字符串切片中移除元素
func RemoveFromSlice(slice []string, item string) []string {
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

// UniqueStrings 去除字符串切片中的重复元素
func UniqueStrings(slice []string) []string {
	keys := make(map[string]bool)
	result := make([]string, 0, len(slice))
	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}
	return result
}

// MergeMaps 合并map
func MergeMaps(maps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// MapKeys 获取map的所有键
func MapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// GetEnv 获取环境变量，带默认值
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SetEnv 设置环境变量
func SetEnv(key, value string) error {
	return os.Setenv(key, value)
}

// GetTempDir 获取临时目录
func GetTempDir() string {
	if dir := os.Getenv("TEMP"); dir != "" {
		return dir
	}
	if dir := os.Getenv("TMP"); dir != "" {
		return dir
	}
	return os.TempDir()
}

// CreateTempFile 创建临时文件
func CreateTempFile(prefix, suffix string) (*os.File, error) {
	return os.CreateTemp(GetTempDir(), prefix+"*"+suffix)
}

// GetWorkingDir 获取当前工作目录
func GetWorkingDir() (string, error) {
	return os.Getwd()
}

// GetExecutablePath 获取可执行文件路径
func GetExecutablePath() (string, error) {
	return os.Executable()
}

// ResolvePath 解析相对路径为绝对路径
func ResolvePath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}

	wd, err := GetWorkingDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(wd, path), nil
}

// IsValidJSON 检查是否为有效的JSON字符串
func IsValidJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

// MinInt 返回最小整数
func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MaxInt 返回最大整数
func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// MinUint64 返回最小uint64
func MinUint64(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

// MaxUint64 返回最大uint64
func MaxUint64(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

// MinFloat64 返回最小浮点数
func MinFloat64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// MaxFloat64 返回最大浮点数
func MaxFloat64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// ClampFloat64 限制浮点数在指定范围内
func ClampFloat64(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// RoundFloat64 四舍五入浮点数到指定小数位
func RoundFloat64(value float64, precision int) float64 {
	factor := math.Pow(10, float64(precision))
	return math.Round(value*factor) / factor
}

func SplitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, sep)
}
