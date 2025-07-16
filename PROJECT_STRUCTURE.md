# 项目目录结构

本项目已重构为模块化的目录结构，便于维护和扩展。

## 目录说明

```
alpha_wx_notify/
├── cmd/                    # 程序入口
│   └── main.go            # 主程序文件
├── internal/              # 内部包
│   ├── airdrop.go         # 空投相关功能
│   └── utils.go           # 通用工具函数
├── config/                # 配置文件
│   └── config.json        # 应用配置
├── data/                  # 数据文件
│   └── last_snapshot.txt  # 快照数据
├── .github/               # GitHub Actions
│   └── workflows/
│       └── go.yml
├── go.mod                 # Go模块定义
├── go.sum                 # 依赖校验
├── .gitignore
├── LICENSE
└── README.md
```

## 模块说明

### cmd/main.go
- 程序入口点
- 调用internal包中的功能
- 保持简洁，只负责程序启动

### internal/airdrop.go
- 空投数据获取和处理
- AirdropService服务类
- 快照生成和比较逻辑
- 价格获取功能

### internal/utils.go
- 配置文件加载
- Server酱推送功能
- 文件操作工具
- HTTP响应处理

### config/config.json
- 应用配置文件
- Server酱推送密钥
- 功能开关设置

### data/last_snapshot.txt
- 存储上次的快照数据
- 用于比较检测变化

## 编译和运行

```bash
# 编译
cd cmd
go build -o alpha.exe

# 运行
./alpha.exe
```

## 优势

1. **模块化**: 功能分离，便于维护
2. **可扩展**: 易于添加新功能模块
3. **标准化**: 遵循Go项目标准目录结构
4. **清晰**: 文件职责明确，代码组织清晰
5. **可测试**: 内部包便于单元测试

## 后续扩展

可以轻松添加新的功能模块，如：
- `internal/news.go` - 新闻公告功能
- `internal/notification.go` - 通知管理
- `internal/database.go` - 数据库操作