# IP 归属地数据文件说明

## 概述

本项目使用 [ip2region](https://github.com/lionsoul2014/ip2region) 离线 IP 定位库来查询 IP 地址归属地。

## 下载 xdb 数据文件

请将 `ip2region.xdb` 或 `ip2region_v4.xdb` 文件下载并放置在此目录下。

### 下载方式

1. **从 GitHub Releases 下载**：
   - 访问：https://github.com/lionsoul2014/ip2region/releases
   - 下载最新版本的 `ip2region.xdb` 或 `ip2region_v4.xdb` 文件

2. **从 ip2region 官方社区下载**：
   - 访问：https://ip2region.net
   - 下载最新的 xdb 数据文件

3. **从项目源码目录下载**：
   - 访问：https://github.com/lionsoul2014/ip2region/tree/master/data
   - 下载 `ip2region_v4.xdb` 文件

### 文件位置

将下载的 xdb 文件重命名为 `ip2region.xdb` 或 `ip2region_v4.xdb`，并放置在此目录下：

```
data/
  ├── README.md
  └── ip2region.xdb  (或 ip2region_v4.xdb)
```

### 注意事项

1. **文件大小**：xdb 文件大小约为 10-20MB
2. **更新频率**：建议定期更新 xdb 文件以获取最新的 IP 归属地数据
3. **性能**：使用 Vector Index 缓存方式，查询效率为 10 微秒级别
4. **内存占用**：Vector Index 缓存占用约 512KB 内存

## 验证安装

启动应用后，检查日志中是否有以下信息：

```
IP 归属地查询器初始化成功 db_path=/path/to/data/ip2region.xdb
```

如果看到初始化失败的错误，请检查：
1. xdb 文件是否存在于 `data/` 目录下
2. 文件权限是否正确
3. 文件是否损坏
