# Go Qemu - Proxmox VM API

> Go Qemu 是基於 Go 語言開發的 Proxmox VE 虛擬機管理 API 服務，提供虛擬機創建、管理和控制功能。支援 Debian、Ubuntu、RockyLinux 等作業系統的自動化部署。

[![pkg](https://pkg.go.dev/badge/github.com/pardnchiu/go-qemu.svg)](https://pkg.go.dev/github.com/pardnchiu/go-qemu)
[![version](https://img.shields.io/github/v/tag/pardnchiu/go-qemu?label=release)](https://github.com/pardnchiu/go-qemu/releases)
[![license](https://img.shields.io/github/license/pardnchiu/go-qemu)](LICENSE)<br>
[![readme](https://img.shields.io/badge/readme-EN-white)](README.md)
[![readme](https://img.shields.io/badge/readme-ZH-white)](README.zh.md)

## 核心特色

### 完整虛擬機生命週期管理
- 支援多種 Linux 發行版本自動安裝
- 即時 SSE 串流安裝進度回饋
- 智能 IP 地址分配與管理
- 完整的 SSH 金鑰配置

### 多節點叢集支援
- 支援 Proxmox VE 叢集環境
- 遠端節點 SSH 操作支援
- 多節點統一請求 API

### 安全性設計
- IP 白名單存取控制
- 私有網路限制存取
- SSH 金鑰認證機制

## 如何使用

### 健康檢查
```
GET /api/health
```

### 虛擬機管理

#### 創建虛擬機
> 支援 OS
> - Debian: 11, 12, 13
> - RockyLinux: 8, 9, 10
> - Ubuntu: 20.04, 22.04, 24.04  
```
POST /api/vm/install
```

```json
{
  "id": 101,                                          // 可選，VM ID，不指定則自動分配
  "name": "test-vm",                                  // VM 名稱
  "node": "pve1",                                     // 可選，指定節點名稱
  "os": "ubuntu",                                     // 必填，支援: debian, ubuntu, rockylinux
  "version": "22.04",                                 // 必填，作業系統版本
  "cpu": 2,                                           // vCPU 核心數，預設 2
  "ram": 2048,                                        // RAM 大小 (MB)，預設 2048
  "disk": "20G",                                      // 硬碟大小，預設 16G
  "passwd": "password123",                            // SSH 密碼，預設 "passwd"
  "pubkey": "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5..."     // 可選，SSH 公鑰
}
```

**SSE 回應**
```javascript
data: {"step":"preparation > checking VMID","status":"processing","message":"[*] using specified VMID: 101"}
data: {"step":"VM creation > creating VM","status":"success","message":"[+] VM created successfully (2.45s)"}
```

#### 啟動虛擬機
```
POST /api/vm/{id}/start
```

**SSE 回應**
```javascript
data: {"step":"starting VM","status":"success","message":"[+] VM starting (1.23s)"}
data: {"step":"waiting for SSH","status":"success","message":"[+] VM is ready (15.67s)"}
```

#### 重啟虛擬機
```
POST /api/vm/{id}/reboot
```

**SSE 回應**
```javascript
data: {"step":"rebooting VM","status":"success","message":"[+] VM rebooting (1.23s)"}
data: {"step":"waiting for SSH","status":"success","message":"[+] VM is ready (15.67s)"}
```

#### 虛擬機狀態
```
GET /api/vm/{id}/status
```

#### 關閉虛擬機
```
POST /api/vm/{id}/shutdown
```

#### 停止虛擬機（強制關機）
```
POST /api/vm/{id}/stop
```

#### 移除虛擬機
```
POST /api/vm/{id}/destroy
```

## 環境配置

### 必要環境變數
```bash
PORT=8080

# 主要節點名稱
MAIN_NODE=PVE1

# 網路閘道
GATEWAY=10.0.0.1

# 允許存取的 IP（逗號分隔，0.0.0.0 允許所有）
ALLOW_IPS=0.0.0.0

# 節點 IP 配置
NODE_PVE1=10.0.0.2
NODE_PVE2=10.0.0.3
NODE_PVE3=10.0.0.4

# IP 分配範圍
ASSIGN_IP_START=100
ASSIGN_IP_END=254
```

### 初始化腳本

系統會自動執行對應的初始化腳本：
```
sh/
├── debian_11.sh
├── debian_12.sh  
├── debian_13.sh
├── ubuntu_20.04.sh
├── ubuntu_22.04.sh
├── ubuntu_24.04.sh
├── rockylinux_8.sh
├── rockylinux_9.sh
└── rockylinux_10.sh
```

## 使用範例

### 創建虛擬機
```bash
curl -X POST http://localhost:8080/api/vm/install \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-vm",
    "os": "debian",
    "version": "12",
    "cpu": 2,
    "ram": 2048,
    "disk": "20G",
    "passwd": "password123"
  }'
```

### 管理虛擬機
```bash
# 獲取狀態
curl http://localhost:8080/api/vm/101/status

# 啟動
curl -X POST http://localhost:8080/api/vm/101/start

# 停止
curl -X POST http://localhost:8080/api/vm/101/stop

# 重啟
curl -X POST http://localhost:8080/api/vm/101/reboot

# 銷毀
curl -X POST http://localhost:8080/api/vm/101/destroy
```

## 授權條款

此原始碼專案採用 [MIT](LICENSE) 授權條款。

## 作者

<img src="https://avatars.githubusercontent.com/u/25631760" align="left" width="96" height="96" style="margin-right: 0.5rem;">

<h4 style="padding-top: 0">邱敬幃 Pardn Chiu</h4>

<a href="mailto:dev@pardn.io" target="_blank">
  <img src="https://pardn.io/image/email.svg" width="48" height="48">
</a> <a href="https://linkedin.com/in/pardnchiu" target="_blank">
  <img src="https://pardn.io/image/linkedin.svg" width="48" height="48">
</a>

***

©️ 2025 [邱敬幃 Pardn Chiu](https://pardn.io)
