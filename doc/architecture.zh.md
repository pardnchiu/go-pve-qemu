# go-pve-qemu — 架構

> 返回 [README](./README.zh.md)

## 目錄

- [概覽](#概覽)
- [模組：Handler（處理器）](#模組handler處理器)
- [模組：Service（服務層）](#模組service服務層)
- [模組：Config（設定層）](#模組config設定層)
- [模組：Model（資料模型）](#模組model資料模型)
- [模組：Util（工具層）](#模組util工具層)
- [資料流](#資料流)
- [VM 生命週期狀態機](#vm-生命週期狀態機)

---

## 概覽

```mermaid
graph TB
    subgraph Client
        API[HTTP Client / curl]
    end

    subgraph go-pve-qemu
        Handler[Handler Layer<br/>請求處理與驗證]
        Service[Service Layer<br/>業務邏輯與指令派送]
        Config[Config Layer<br/>路由註冊與 CORS]
        Model[Model Layer<br/>資料型別定義]
        Util[Util Layer<br/>輔助函式]
    end

    subgraph Proxmox[Proxmox VE Cluster]
        MainNode[Main Node<br/>qm / pvesh 本機執行]
        RemoteNode[Remote Node<br/>SSH 派送 qm 指令]
        Storage[Storage Pool<br/>local-zfs / dir / nfs]
    end

    subgraph External[外部資源]
        CloudImage[OS Cloud Image<br/>cloud.debian.org /<br/>cloud-images.ubuntu.com /<br/>dl.rockylinux.org]
        VM[目標 VM<br/>cloud-init + SSH]
    end

    API -->|HTTP + SSE| Handler
    Handler --> Service
    Service --> Config
    Service --> Model
    Service --> Util
    Service -->|qm create / importdisk / set| MainNode
    Service -->|SSH + qm| RemoteNode
    MainNode --> Storage
    RemoteNode --> Storage
    Service -->|下載 qcow2| CloudImage
    CloudImage --> MainNode
    MainNode -->|cloud-init| VM
```

## 模組：Handler（處理器）

Handler 層負責 HTTP 請求的接收、輸入驗證與回應產生。所有對外暴露的 API Endpoint 皆由此層定義。

```mermaid
graph TB
    subgraph Handler
        Install[Install<br/>POST /api/vm/install]
        Start[Start<br/>POST /api/vm/:id/start]
        Stop[Stop<br/>POST /api/vm/:id/stop]
        Shutdown[Shutdown<br/>POST /api/vm/:id/shutdown]
        Reboot[Reboot<br/>POST /api/vm/:id/reboot]
        Destroy[Destroy<br/>POST /api/vm/:id/destroy]
        CPU[Set CPU<br/>POST /api/vm/:id/set/cpu]
        Memory[Set Memory<br/>POST /api/vm/:id/set/memory]
        Disk[Set Disk<br/>POST /api/vm/:id/set/disk]
        Node[Set Node<br/>POST /api/vm/:id/set/node]
        GetStatus[Get Status<br/>GET /api/vm/:id/status]
        GetVMList[Get VM List<br/>GET /api/vm/list]
    end

    Install -->|SSE 串流| Service
    Start -->|SSE 串流| Service
    Reboot -->|SSE 串流| Service
    Node -->|SSE 串流| Service
    Stop --> Service
    Shutdown --> Service
    Destroy --> Service
    CPU --> Service
    Memory --> Service
    Disk --> Service
    GetStatus --> Service
    GetVMList --> Service
```

**職責：**
- 解析 HTTP 請求與路徑參數
- 透過 `util.CheckIP()` 驗證來源 IP 是否在白名單中
- 透過 `util.CheckID()` 驗證 VMID 是否存在及狀態正確
- 將請求轉發至 Service 層對應方法
- 對 SSE Endpoint 設定 SSE 標頭並串流事件

## 模組：Service（服務層）

Service 層包含所有業務邏輯，是系統的核心。它封裝了 Proxmox CLI 指令的執行、SSH 派送、OS 映像管理與 IP 分配邏輯。

```mermaid
graph TB
    subgraph Service
        InstallSvc[Install<br/>完整 VM 佈建管線]
        StartSvc[Start<br/>啟動 + SSH 等待]
        StopSvc[Stop / Shutdown]
        RebootSvc[Reboot<br/>重啟 + SSH 等待]
        DestroySvc[Destroy<br/>停止 + 清除]
        CPUSvc[Set CPU]
        MemorySvc[Set Memory]
        DiskSvc[Set Disk]
        NodeSvc[Migrate Node<br/>節點遷移]
        GetVMStatusSvc[Get VM Status]
    end

    subgraph Internal[內部方法]
        initConfig[initConfig<br/>設定預設值]
        getStorages[getStorages<br/>列出儲存池]
        getOSImage[getOSImage<br/>解析 OS 映像 URL]
        checkOSImageURL[checkOSImageURL<br/>驗證映像 URL]
        downloadOSImage[downloadOSImage<br/>下載 qcow2]
        createVM[createVM<br/>qm create]
        importDisk[importDisk<br/>qm importdisk]
        initialVM[initialVM<br/>qm set + cloud-init]
        initialWithSSH[initialWithSSH<br/>SSH 初始化腳本]
        CheckAlive[CheckAlive<br/>SSH 輪詢]
        assignIP[assignIP<br/>並發 IP 分配]
        getCommand[getCommand<br/>本機 / SSH 指令選擇]
        runCommandSSE[runCommandSSE<br/>SSE 串流指令輸出]
        checkCPUArch[checkCPUArch<br/>CPU 架構偵測]
        GetClusterCPUType[GetClusterCPUType<br/>叢集 CPU 類型快取]
        getVMIDsNode[getVMIDsNode<br/>查詢 VM 所在節點]
    end

    InstallSvc --> initConfig
    InstallSvc --> getStorages
    InstallSvc --> getOSImage
    InstallSvc --> checkOSImageURL
    InstallSvc --> downloadOSImage
    InstallSvc --> createVM
    InstallSvc --> importDisk
    InstallSvc --> initialVM
    InstallSvc --> initialWithSSH
    InstallSvc --> CheckAlive
    InstallSvc --> assignIP
    InstallSvc --> getCommand
    InstallSvc --> runCommandSSE
    InstallSvc --> GetClusterCPUType
    StartSvc --> CheckAlive
    StartSvc --> getCommand
    NodeSvc --> runCommandSSE
    NodeSvc --> getCommand
```

**職責：**
- 實作 VM 完整生命週期管理（建立、啟動、停止、重啟、刪除）
- 封裝 `qm` 與 `pvesh` CLI 指令的參數組合與執行
- 自動判斷指令應在本機執行或透過 SSH 派送至遠端節點
- 管理 OS Cloud 映像的下載與快取
- 並發掃描可用 VMID 與 IP 位址
- 偵測叢集所有節點的 CPU 架構相容性並快取結果
- 透過 SSE 即時串流指令執行進度

## 模組：Config（設定層）

```mermaid
graph TB
    subgraph Config
        Routes[NewRoutes<br/>路由註冊]
        CORS[CORS Middleware<br/>跨域與 IP 限制]
    end

    Routes -->|GET /api/health| Health[回傳 ok]
    Routes -->|POST /api/vm/install| HandlerInstall[Handler.Install]
    Routes -->|GET /api/vm/list| HandlerList[Handler.GetVMList]
    Routes -->|GET /api/vm/:id/status| HandlerStatus[Handler.GetStatus]
    Routes -->|POST /api/vm/:id/start| HandlerStart[Handler.Start]
    Routes -->|POST /api/vm/:id/stop| HandlerStop[Handler.Stop]
    Routes -->|POST /api/vm/:id/shutdown| HandlerShutdown[Handler.Shutdown]
    Routes -->|POST /api/vm/:id/reboot| HandlerReboot[Handler.Reboot]
    Routes -->|POST /api/vm/:id/destroy| HandlerDestroy[Handler.Destroy]
    Routes -->|POST /api/vm/:id/set/cpu| HandlerCPU[Handler.CPU]
    Routes -->|POST /api/vm/:id/set/memory| HandlerMemory[Handler.Memory]
    Routes -->|POST /api/vm/:id/set/disk| HandlerDisk[Handler.Disk]
    Routes -->|POST /api/vm/:id/set/node| HandlerNode[Handler.Node]
```

**職責：**
- 將所有 API 路由註冊至 Gin Engine
- 設定 CORS Middleware，僅允許私有 IP 範圍存取

## 模組：Model（資料模型）

```mermaid
classDiagram
    class Config {
        +int ID
        +string Name
        +string Node
        +string Storage
        +string OS
        +string Version
        +int CPU
        +string Disk
        +int RAM
        +string IP
        +string Gateway
        +string User
        +string Passwd
        +string Pubkey
    }

    class Response {
        +bool Success
        +string Message
        +int VMID
        +string IP
    }

    class SSE {
        +string Step
        +string Status
        +string Message
        +int VMID
        +string IP
    }

    class VM {
        +int ID
        +string Name
        +string OS
        +bool Running
        +string Node
        +int CPU
        +int Disk
        +int Memory
        +int MemoryUsed
    }

    class Node {
        +string Node
        +float64 MaxCPU
        +float64 MaxMemory
        +float64 CPU
        +float64 Memory
        +float64 MemoryUsed
        +float64 Disk
        +bool Running
    }

    class Status {
        +string IP
        +bool Available
        +int VMID
    }
```

**職責：**
- 定義所有資料傳輸物件（DTO）與內部型別
- `Config`：VM 佈建請求的完整設定
- `Response`：API 統一回應格式
- `SSE`：Server-Sent Events 事件格式
- `VM`：VM 狀態與資源資訊
- `Node`：Proxmox 節點資源資訊
- `Status`：IP 可用性檢查結果

## 模組：Util（工具層）

```mermaid
graph TB
    subgraph Util
        CheckIP[CheckIP<br/>來源 IP 白名單檢查]
        CheckID[CheckID<br/>VMID 存在性與狀態驗證]
        GetVMMap[GetVMMap<br/>取得叢集 VM 清單]
        GetNodeMap[GetNodeMap<br/>取得叢集節點清單]
        GetOSUser[GetOSUser<br/>查詢 VM 的 OS 使用者]
        IncludeVM[IncludeVM<br/>VM 過濾器]
    end

    GetVMMap -->|pvesh get /cluster/resources| ProxmoxCluster[Proxmox Cluster]
    GetNodeMap -->|pvesh get /cluster/resources| ProxmoxCluster
    CheckID --> GetVMMap
```

**職責：**
- 提供跨 Service 與 Handler 層共用的輔助函式
- 透過 `pvesh` CLI 查詢 Proxmox 叢集資源
- 管理 `.go_qemu_disabled` 檔案中的停用 VM 清單

## 資料流

### 一般請求流程

```mermaid
sequenceDiagram
    participant Client
    participant Handler
    participant Service
    participant Proxmox as Proxmox Cluster
    participant VM as 目標 VM

    Client->>Handler: HTTP 請求
    Handler->>Handler: CheckIP（IP 白名單）
    Handler->>Handler: CheckID（VMID 驗證）
    Handler->>Service: 呼叫對應服務方法
    Service->>Service: getVMIDsNode（查詢 VM 所在節點）
    alt 本機節點
        Service->>Proxmox: 本機執行 qm / pvesh
    else 遠端節點
        Service->>Proxmox: SSH 派送 qm 指令
    end
    Proxmox-->>Service: 指令輸出
    Service-->>Handler: 結果
    Handler-->>Client: HTTP 回應
```

### VM 安裝流程

```mermaid
sequenceDiagram
    participant Client
    participant API as go-pve-qemu
    participant Proxmox
    participant Cloud as Cloud Image Repo
    participant VM as 新 VM

    Client->>API: POST /api/vm/install（JSON）
    API-->>Client: SSE: 開始安裝

    API->>API: 分配 VMID（並發掃描）
    API-->>Client: SSE: VMID 已分配

    API->>API: 分配 IP（Gateway + VMID）
    API-->>Client: SSE: IP 已分配

    API->>API: 驗證 CPU / RAM / Disk 上限
    API-->>Client: SSE: 資源已驗證

    API->>Cloud: 下載 OS Cloud 映像
    Cloud-->>API: qcow2 檔案
    API-->>Client: SSE: 映像下載完成

    API->>Proxmox: qm create（VM 建立）
    Proxmox-->>API: VM 建立成功
    API-->>Client: SSE: VM 已建立

    API->>Proxmox: qm importdisk（磁碟匯入）
    Proxmox-->>API: 磁碟匯入成功
    API-->>Client: SSE: 磁碟已匯入

    API->>Proxmox: qm set（cloud-init、SSH 金鑰、開機順序）
    Proxmox-->>API: VM 設定完成
    API-->>Client: SSE: VM 初始化完成

    alt 指定節點
        API->>Proxmox: qm migrate（節點遷移）
        Proxmox-->>API: 遷移完成
        API-->>Client: SSE: 遷移完成
    end

    API->>Proxmox: qm start
    Proxmox-->>API: VM 啟動中

    loop SSH 輪詢（最長 60 次）
        API->>VM: SSH 連線測試
        alt 連線成功
            VM-->>API: ready
        else 連線失敗
            API->>API: 等待 5 秒
        end
    end

    API->>VM: SSH 執行初始化腳本
    VM-->>API: 初始化完成
    API-->>Client: SSE: SSH 初始化完成

    API->>Proxmox: qm reboot
    Proxmox-->>API: VM 重啟中

    loop SSH 輪詢（最長 60 次）
        API->>VM: SSH 連線測試
        alt 連線成功
            VM-->>API: ready
        else 連線失敗
            API->>API: 等待 5 秒
        end
    end

    API-->>Client: SSE: VM 安裝完成（VMID、IP、User）
    API-->>Client: event: close
```

## VM 生命週期狀態機

```mermaid
stateDiagram-v2
    [*] --> Creating: POST /api/vm/install
    Creating --> Created: qm create + importdisk 完成
    Created --> Running: qm start
    Running --> Stopped: qm stop / shutdown
    Running --> Rebooting: POST /api/vm/:id/reboot
    Rebooting --> Running: SSH 重新就緒
    Running --> Migrating: POST /api/vm/:id/set/node
    Migrating --> Running: 遷移完成
    Stopped --> Running: POST /api/vm/:id/start
    Stopped --> Destroyed: POST /api/vm/:id/destroy
    Running --> Destroyed: POST /api/vm/:id/destroy
    Destroyed --> [*]
```

---

©️ 2025 [邱敬幃 Pardn Chiu](https://www.linkedin.com/in/pardnchiu)