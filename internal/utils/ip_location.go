package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/golang-pay-core/internal/logger"
	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"go.uber.org/zap"
)

// IPLocationInfo IP 归属地信息
type IPLocationInfo struct {
	Address string // 归属地（格式：省/市，如：广东省/深圳市）
	PID     int    // 省份ID（代理省ip）
	CID     int    // 城市ID（代理城市ip）
}

var (
	ipSearcher     *xdb.Searcher
	ipSearcherOnce sync.Once
	ipSearcherErr  error
)

// initIPSearcher 初始化 IP 归属地查询器
// 使用 NewWithBuffer 将整个 xdb 文件加载到内存，实现并发安全的查询
// 优势：
// 1. 并发安全：完全基于内存的查询是并发安全的，适合高并发场景
// 2. 极致性能：10微秒级别的查询效率，无磁盘 IO 操作
// 3. 内存占用：约 10-20MB（等同于 xdb 文件大小）
func initIPSearcher() {
	ipSearcherOnce.Do(func() {
		// 尝试多个可能的 xdb 文件路径
		possiblePaths := []string{
			"data/ip2region.xdb",
			"data/ip2region_v4.xdb",
			"./data/ip2region.xdb",
			"./data/ip2region_v4.xdb",
			"../data/ip2region.xdb",
			"../data/ip2region_v4.xdb",
		}

		var dbPath string
		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				absPath, err := filepath.Abs(path)
				if err == nil {
					dbPath = absPath
					break
				}
			}
		}

		if dbPath == "" {
			ipSearcherErr = fmt.Errorf("未找到 ip2region.xdb 文件，请下载并放置在 data/ 目录下")
			logger.Logger.Warn("IP 归属地查询器初始化失败",
				zap.Error(ipSearcherErr),
				zap.Strings("尝试的路径", possiblePaths))
			return
		}

		// 加载 Header 获取版本信息
		header, err := xdb.LoadHeaderFromFile(dbPath)
		if err != nil {
			ipSearcherErr = fmt.Errorf("加载 Header 失败: %w", err)
			logger.Logger.Warn("IP 归属地查询器初始化失败",
				zap.String("db_path", dbPath),
				zap.Error(ipSearcherErr))
			return
		}

		// 从 Header 获取版本对象
		version, err := xdb.VersionFromHeader(header)
		if err != nil {
			ipSearcherErr = fmt.Errorf("获取版本信息失败: %w", err)
			logger.Logger.Warn("IP 归属地查询器初始化失败",
				zap.String("db_path", dbPath),
				zap.Error(ipSearcherErr))
			return
		}

		// 将整个 xdb 文件加载到内存（并发安全，性能最优）
		// 内存占用等同于 xdb 文件大小（约 10-20MB）
		cBuff, err := xdb.LoadContentFromFile(dbPath)
		if err != nil {
			ipSearcherErr = fmt.Errorf("加载 xdb 内容失败: %w", err)
			logger.Logger.Warn("IP 归属地查询器初始化失败",
				zap.String("db_path", dbPath),
				zap.Error(ipSearcherErr))
			return
		}

		// 创建查询器（完全基于内存，并发安全）
		searcher, err := xdb.NewWithBuffer(version, cBuff)
		if err != nil {
			ipSearcherErr = fmt.Errorf("创建查询器失败: %w", err)
			logger.Logger.Warn("IP 归属地查询器初始化失败",
				zap.String("db_path", dbPath),
				zap.Error(ipSearcherErr))
			return
		}

		ipSearcher = searcher
		logger.Logger.Info("IP 归属地查询器初始化成功（完全基于内存，并发安全）",
			zap.String("db_path", dbPath),
			zap.Int("memory_size_mb", len(cBuff)/(1024*1024)))
	})
}

// GetIPLocation 查询 IP 地址归属地信息
// 使用 ip2region 离线 IP 定位库（支持 IPv4 和 IPv6）
// 查询效率：10微秒级别（完全基于内存，无磁盘 IO）
// 并发安全：完全基于内存的查询是并发安全的，适合高并发场景
func GetIPLocation(ip string) (*IPLocationInfo, error) {
	if ip == "" {
		return &IPLocationInfo{
			Address: "",
			PID:     -1,
			CID:     -1,
		}, nil
	}

	// 过滤本地 IP 和私有 IP
	if isLocalIP(ip) {
		return &IPLocationInfo{
			Address: "本地",
			PID:     -1,
			CID:     -1,
		}, nil
	}

	// 初始化查询器（只初始化一次）
	initIPSearcher()

	// 检查初始化是否成功
	if ipSearcherErr != nil || ipSearcher == nil {
		logger.Logger.Warn("IP 归属地查询器未初始化，跳过查询",
			zap.String("ip", ip),
			zap.Error(ipSearcherErr))
		return &IPLocationInfo{
			Address: "",
			PID:     -1,
			CID:     -1,
		}, nil // 失败时返回空值，不阻塞主流程
	}

	// 查询 IP 归属地
	region, err := ipSearcher.SearchByStr(ip)
	if err != nil {
		logger.Logger.Warn("查询 IP 归属地失败",
			zap.String("ip", ip),
			zap.Error(err))
		return &IPLocationInfo{
			Address: "",
			PID:     -1,
			CID:     -1,
		}, nil // 失败时返回空值，不阻塞主流程
	}

	// 解析 region 字符串（格式：国家|省份|城市|ISP）
	// 例如：中国|0|广东省|深圳市|0
	parts := strings.Split(region, "|")
	if len(parts) < 3 {
		logger.Logger.Warn("IP 归属地格式异常",
			zap.String("ip", ip),
			zap.String("region", region))
		return &IPLocationInfo{
			Address: region,
			PID:     -1,
			CID:     -1,
		}, nil
	}

	// 构建归属地字符串（省/市）
	var addressParts []string
	country := parts[0]
	province := parts[2]
	city := ""
	if len(parts) > 3 {
		city = parts[3]
	}

	// 过滤掉 "0" 和空字符串
	if province != "" && province != "0" {
		addressParts = append(addressParts, province)
	}
	if city != "" && city != "0" {
		addressParts = append(addressParts, city)
	}

	address := strings.Join(addressParts, "/")
	if address == "" && country != "" && country != "0" {
		address = country
	}

	// 注意：pid 和 cid 需要根据实际的省份/城市映射表来设置
	// 这里暂时返回 -1，如果需要可以后续添加映射逻辑
	return &IPLocationInfo{
		Address: address,
		PID:     -1, // TODO: 根据省份名称映射到省份ID
		CID:     -1, // TODO: 根据城市名称映射到城市ID
	}, nil
}

// isLocalIP 检查是否为本地 IP 或私有 IP
func isLocalIP(ip string) bool {
	// 检查常见的本地 IP
	localIPs := []string{
		"127.0.0.1",
		"localhost",
		"::1",
		"0.0.0.0",
	}

	for _, local := range localIPs {
		if ip == local {
			return true
		}
	}

	// 检查私有 IP 地址段
	if strings.HasPrefix(ip, "192.168.") ||
		strings.HasPrefix(ip, "10.") ||
		strings.HasPrefix(ip, "172.16.") ||
		strings.HasPrefix(ip, "172.17.") ||
		strings.HasPrefix(ip, "172.18.") ||
		strings.HasPrefix(ip, "172.19.") ||
		strings.HasPrefix(ip, "172.20.") ||
		strings.HasPrefix(ip, "172.21.") ||
		strings.HasPrefix(ip, "172.22.") ||
		strings.HasPrefix(ip, "172.23.") ||
		strings.HasPrefix(ip, "172.24.") ||
		strings.HasPrefix(ip, "172.25.") ||
		strings.HasPrefix(ip, "172.26.") ||
		strings.HasPrefix(ip, "172.27.") ||
		strings.HasPrefix(ip, "172.28.") ||
		strings.HasPrefix(ip, "172.29.") ||
		strings.HasPrefix(ip, "172.30.") ||
		strings.HasPrefix(ip, "172.31.") {
		return true
	}

	return false
}
