# go-pve-qemu — 技術文件

> 返回 [README](./README.zh.md)

## 目錄

- [前置需求](#前置需求)
- [安裝](#安裝)
- [設定](#設定)
- [使用方式](#使用方式)
- [API 參考](#api-參考)
- [SSE 事件格式](#sse-事件格式)
- [安裝流程階段](#安裝流程階段)

---

## 前置需求

- Proxmox 主節點上已安裝 **Go ≥ 1.20**
- 標準 Proxmox VE CLI 工具 **`qm`** 與 **`pvesh`** 可用
- 主節點至所有遠端節點的**金鑰式 SSH 存取**（無 Passphrase）
- 網路閘道設定使 **VMID 等於 IP 最後一個位元組**（例如 `192.168.0.X`，其中 `X` = VMID）
- 目標 Proxmox 儲存池（`ASSIGN_STORAGE`）必須存在且為啟用狀態

## 安裝

### 原始碼建置

```bash
git clone https://github.com/pardnchiu/go-pve-qemu.git
cd go-pve-qemu
cp .env.example .env
# 編輯 .env（參閱設定章節）
go build -o go-pve-qemu ./cmd/api
./go-pve-qemu
```

伺服器啟動時會列印所有已註冊的路由，並在 `PORT` 指定的埠號上監聽。

### Systemd 服務（選用）

```ini
[Unit]
Description=go-pve-qemu API
After=network.target

[Service]
WorkingDirectory=/opt/go-pve-qemu
ExecStart=/opt/go-pve-qemu/go-pve-qemu
Restart=on-failure
EnvironmentFile=/opt/go-pve-qemu/.env

[Install]
WantedBy=multi-user.target
```

## 設定

所有設定均透過環境變數完成（啟動時由 `godotenv` 從 `.env` 載入）。

### 必填

| 變數 | 說明 |
|------|------|
| `PORT` | API 伺服器監聽埠號 |
| `GATEWAY` | 網路閘道 — 驅動 IP 分配（例如 `192.168.0.1`） |
| `MAIN_NODE` | 主要 Proxmox 節點名稱（指令在此節點本機執行） |
| `NODE_<name>` | 各叢集節點 IP 位址（例如 `NODE_pve2=192.168.0.12`） |
| `ASSIGN_STORAGE` | 磁碟匯入用 Proxmox 儲存池（例如 `local-zfs`） |

### 選填

| 變數 | 預設值 | 說明 |
|------|--------|------|
| `ASSIGN_IP_START` | `100` | IP 範圍起始值 — 最後一位元組（最小 `100`） |
| `ASSIGN_IP_END` | `254` | IP 範圍結束值 — 最後一位元組（最大 `254`） |
| `ALLOW_IPS` | `0.0.0.0` | 以逗號分隔的允許來源 IP 清單；`0.0.0.0` 允許所有來源 |
| `VM_MAX_CPU` | 無限制 | 每台 VM 最大 vCPU 核心數（1–32） |
| `VM_MAX_RAM` | 無限制 | 每台 VM 最大記憶體（MB） |
| `VM_MAX_DISK` | 無限制 | 每台 VM 最大磁碟大小（GB） |
| `VM_BALLOON_MIN` | — | Balloon 裝置最小記憶體（MB）；設定後若 RAM 超過此值則啟用 Balloon |
| `VM_ROOT_PASSWORD` | — | 透過 OS 初始化腳本注入的預設 root 密碼 |

### `.env` 範例

```bash
PORT=8080
GATEWAY=192.168.0.1
MAIN_NODE=pve1
NODE_pve1=192.168.0.11
NODE_pve2=192.168.0.12
NODE_pve3=192.168.0.13
ASSIGN_STORAGE=local-zfs
ASSIGN_IP_START=100
ASSIGN_IP_END=254
ALLOW_IPS=0.0.0.0
VM_MAX_CPU=32
VM_MAX_RAM=32768
VM_MAX_DISK=64
VM_BALLOON_MIN=2048
VM_ROOT_PASSWORD=
```

## 使用方式

### 健康檢查

```bash
curl http://localhost:8080/api/health
# 回應: ok
```

### 佈建 VM

使用單一 API 呼叫即可完成完整 VM 佈建：

```bash
curl -X POST http://localhost:8080/api/vm/install \
  -H "Content-Type: application/json" \
  -d '{
    "os": "ubuntu",
    "version": "22.04",
    "name": "web-01",
    "cpu": 2,
    "ram": 4096,
    "disk": "40G",
    "user": "deploy",
    "passwd": "s3cr3t"
  }'
```

回應為 SSE 串流，即時顯示每個階段的進度與耗時。

### 列出 VM

```bash
curl http://localhost:8080/api/vm/list
```

回應包含叢集範圍的 VM 清單及各節點資源使用率。

### 管理 VM 生命週期

```bash
# 啟動 VM（SSE 串流直至 SSH 就緒）
curl -X POST http://localhost:8080/api/vm/101/start

# 優雅關機
curl -X POST http://localhost:8080/api/vm/101/shutdown

# 強制停止
curl -X POST http://localhost:8080/api/vm/101/stop

# 重新啟動（SSE 串流）
curl -X POST http://localhost:8080/api/vm/101/reboot

# 刪除 VM
curl -X POST http://localhost:8080/api/vm/101/destroy
```

### 調整資源

```bash
# 設定 vCPU
curl -X POST http://localhost:8080/api/vm/101/set/cpu \
  -H "Content-Type: application/json" \
  -d '{"cpu": 4}'

# 設定記憶體
curl -X POST http://localhost:8080/api/vm/101/set/memory \
  -H "Content-Type: application/json" \
  -d '{"memory": 8192}'

# 擴充磁碟
curl -X POST http://localhost:8080/api/vm/101/set/disk \
  -H "Content-Type: application/json" \
  -d '{"disk": "10G"}'
```

### 節點遷移

```bash
curl -X POST http://localhost:8080/api/vm/101/set/node \
  -H "Content-Type: application/json" \
  -d '{"node": "pve2"}'
```

回應為 SSE 串流，即時顯示遷移進度。

## API 參考

所有 Endpoint 均以 `/api` 為前綴。

### 指令表

| 方法 | 路徑 | 回應類型 | 說明 |
|------|------|---------|------|
| GET | `/api/health` | `text/plain` | 健康檢查 — 回傳 `ok` |
| POST | `/api/vm/install` | SSE | 建立並完整佈建 VM |
| GET | `/api/vm/list` | JSON | 列出所有 VM 及各節點叢集統計 |
| GET | `/api/vm/:id/status` | `text/plain` | VM 電源狀態（`running` 或 `stopped`） |
| POST | `/api/vm/:id/start` | SSE | 啟動 VM，串流直至 SSH 就緒 |
| POST | `/api/vm/:id/stop` | `text/plain` | 強制立即停止 VM |
| POST | `/api/vm/:id/shutdown` | `text/plain` | ACPI 優雅關機 |
| POST | `/api/vm/:id/reboot` | SSE | 重新啟動 VM，串流直至 SSH 再次就緒 |
| POST | `/api/vm/:id/destroy` | `text/plain` | 刪除 VM 並清除磁碟 |
| POST | `/api/vm/:id/set/cpu` | `text/plain` | 設定 vCPU 核心數 |
| POST | `/api/vm/:id/set/memory` | `text/plain` | 設定記憶體大小 |
| POST | `/api/vm/:id/set/disk` | `text/plain` | 擴充磁碟 |
| POST | `/api/vm/:id/set/node` | SSE | 將 VM 遷移至另一叢集節點 |

### `POST /api/vm/install` 請求參數

| 欄位 | 類型 | 必填 | 預設值 | 說明 |
|------|------|------|--------|------|
| `os` | string | 是 | — | `debian`、`ubuntu` 或 `rockylinux` |
| `version` | string | 是 | — | OS 版本（例如 `22.04`、`12`、`9`） |
| `name` | string | 否 | VMID | VM 名稱 |
| `id` | int | 否 | 自動 | VMID（100–254）；省略時自動分配 |
| `node` | string | 否 | 主節點 | 目標 Proxmox 節點 |
| `cpu` | int | 否 | `2` | vCPU 核心數（受 `VM_MAX_CPU` 限制） |
| `ram` | int | 否 | `2048` | 記憶體（MB），最小 512（受 `VM_MAX_RAM` 限制） |
| `disk` | string | 否 | `16G` | 磁碟大小（例如 `20G`），最小 `16G` |
| `user` | string | 否 | OS 預設 | SSH 使用者名稱（debian/ubuntu/rocky） |
| `passwd` | string | 否 | `passwd` | SSH 密碼 |
| `pubkey` | string | 否 | — | 欲注入的 SSH 公鑰 |

### `POST /api/vm/:id/set/cpu` 請求參數

```json
{ "cpu": 4 }
```

有效範圍：1–32（另受 `VM_MAX_CPU` 限制）。

### `POST /api/vm/:id/set/memory` 請求參數

```json
{ "memory": 8192 }
```

單位 MB，最小 512，最大 `VM_MAX_RAM`。

### `POST /api/vm/:id/set/disk` 請求參數

```json
{ "disk": "10G" }
```

擴充指定大小的磁碟空間。不支援磁碟縮小。

### `POST /api/vm/:id/set/node` 請求參數

```json
{ "node": "pve2" }
```

以本機磁碟傳輸方式將 VM 遷移至另一叢集節點。

## SSE 事件格式

所有 SSE Endpoint 發送帶有 JSON Payload 的 `data:` 行。

```
data: {"step":"<階段>","status":"<狀態>","message":"<文字>","vm_id":<id>,"ip":"<ip>"}
```

| 欄位 | 值 | 說明 |
|------|-----|------|
| `step` | string | 流程階段名稱（例如 `"preparation > checking VMID"`） |
| `status` | `processing` \| `success` \| `error` | 事件嚴重性 |
| `message` | string | 包含耗用時間的人類可讀細節 |
| `vm_id` | int | VMID（出現於最終成功事件） |
| `ip` | string | 已分配的 IP（出現於最終成功事件） |

### 安裝流程 SSE 範例

```
data: {"step":"preparation > checking VMID","status":"processing","message":"[*] start VM installation"}
data: {"step":"preparation > checking VMID","status":"success","message":"[+] auto-assigned VMID: 101 (0.02s)"}
data: {"step":"preparation > assigning IP","status":"success","message":"[+] assigned IP: 192.168.0.101/24 (0.00s)"}
data: {"step":"OS preparation > downloading OS image","status":"processing","message":"[*] start downloading OS image"}
data: {"step":"OS preparation > downloading OS image","status":"success","message":"[+] completed OS image download (18.34s)"}
data: {"step":"VM creation > creating VM","status":"success","message":"[+] VM created successfully (2.45s)"}
data: {"step":"VM creation > importing disk image","status":"success","message":"[+] disk image imported successfully (8.12s)"}
data: {"step":"VM initialization > initializing configuration","status":"success","message":"[+] VM initialized successfully (1.03s)"}
data: {"step":"VM initialization > waiting for ready","status":"processing","message":"[*] VM starting"}
data: {"step":"VM initialization > SSH initialization","status":"success","message":"[+] SSH initialization completed (5.21s)"}
data: {"step":"VM initialization > finalizing","status":"success","message":"[+] VM installation completed in 67.83s"}
data: {"step":"VM initialization > finalizing","status":"success","message":"[*] VMID: 101"}
data: {"step":"VM initialization > finalizing","status":"success","message":"[*] IP: 192.168.0.101"}
data: {"step":"VM initialization > finalizing","status":"success","message":"[*] User: deploy"}
```

## 安裝流程階段

`POST /api/vm/install` 流程按序執行以下階段：

| 階段 | 說明 |
|------|------|
| `preparation > checking VMID` | 自動分配或驗證指定的 VMID |
| `preparation > assigning IP` | 從閘道前綴 + VMID 推導靜態 IP |
| `preparation > validating CPU and RAM` | 將資源限制在設定上限內 |
| `preparation > validating disk size` | 強制最小 16G 及最大 `VM_MAX_DISK` |
| `preparation > setting default config values` | 套用儲存池與 CPU 類型預設值 |
| `preparation > checking storage pool` | 驗證儲存池存在且為啟用狀態 |
| `OS preparation > getting OS image` | 解析官方 Cloud 映像 URL |
| `OS preparation > validating OS image URL` | 對映像 URL 執行 HTTP HEAD 檢查 |
| `OS preparation > downloading OS image` | 若未快取則下載 |
| `SSH preparation > checking SSH key` | 驗證或生成主機 SSH 金鑰對 |
| `VM creation > creating VM` | 以 Cloud-init 硬體設定執行 `qm create` |
| `VM creation > importing disk image` | 執行 `qm importdisk` 匯入至儲存池 |
| `VM initialization > initializing configuration` | 掛載磁碟、設定開機順序、cloud-init 磁碟機 |
| `VM initialization > migrating VM` | 遷移至目標節點（若指定 `node`） |
| `VM initialization > waiting for ready` | 輪詢 SSH 直至 VM 可存取 |
| `VM initialization > SSH initialization` | 透過 SSH 執行 OS 特定初始化腳本 |
| `VM initialization > rebooting VM` | 重新啟動以套用所有設定 |
| `VM initialization > finalizing` | 確認重啟後 SSH 就緒，發送 VMID/IP |

---

©️ 2025 [邱敬幃 Pardn Chiu](https://www.linkedin.com/in/pardnchiu)