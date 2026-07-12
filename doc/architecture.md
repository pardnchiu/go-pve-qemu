# go-pve-qemu — Architecture

> Back to [README](../README.md)

## Table of Contents

- [Overview](#overview)
- [Module: Handler](#module-handler)
- [Module: Service](#module-service)
- [Module: Config](#module-config)
- [Module: Model](#module-model)
- [Module: Util](#module-util)
- [Data Flow](#data-flow)
- [VM Lifecycle State Machine](#vm-lifecycle-state-machine)

---

## Overview

```mermaid
graph TB
    subgraph Client
        API[HTTP Client / curl]
    end

    subgraph go-pve-qemu
        Handler[Handler Layer<br/>Request Processing & Validation]
        Service[Service Layer<br/>Business Logic & Command Dispatch]
        Config[Config Layer<br/>Route Registration & CORS]
        Model[Model Layer<br/>Data Type Definitions]
        Util[Util Layer<br/>Helper Functions]
    end

    subgraph Proxmox[Proxmox VE Cluster]
        MainNode[Main Node<br/>Local qm / pvesh Execution]
        RemoteNode[Remote Node<br/>SSH-Dispatched qm Commands]
        Storage[Storage Pool<br/>local-zfs / dir / nfs]
    end

    subgraph External[External Resources]
        CloudImage[OS Cloud Image<br/>cloud.debian.org /<br/>cloud-images.ubuntu.com /<br/>dl.rockylinux.org]
        VM[Target VM<br/>cloud-init + SSH]
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
    Service -->|download qcow2| CloudImage
    CloudImage --> MainNode
    MainNode -->|cloud-init| VM
```

## Module: Handler

The Handler layer receives HTTP requests, validates inputs, and produces responses. All externally exposed API endpoints are defined here.

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

    Install -->|SSE Stream| Service
    Start -->|SSE Stream| Service
    Reboot -->|SSE Stream| Service
    Node -->|SSE Stream| Service
    Stop --> Service
    Shutdown --> Service
    Destroy --> Service
    CPU --> Service
    Memory --> Service
    Disk --> Service
    GetStatus --> Service
    GetVMList --> Service
```

**Responsibilities:**
- Parse HTTP requests and path parameters
- Validate source IP against the allowlist via `util.CheckIP()`
- Verify VMID existence and state via `util.CheckID()`
- Forward requests to the corresponding Service layer method
- Set SSE headers and stream events for SSE endpoints

## Module: Service

The Service layer contains all business logic and is the core of the system. It encapsulates Proxmox CLI command execution, SSH dispatch, OS image management, and IP allocation logic.

```mermaid
graph TB
    subgraph Service
        InstallSvc[Install<br/>Full VM Provisioning Pipeline]
        StartSvc[Start<br/>Boot + SSH Wait]
        StopSvc[Stop / Shutdown]
        RebootSvc[Reboot<br/>Restart + SSH Wait]
        DestroySvc[Destroy<br/>Stop + Purge]
        CPUSvc[Set CPU]
        MemorySvc[Set Memory]
        DiskSvc[Set Disk]
        NodeSvc[Migrate Node<br/>Cross-Node Migration]
        GetVMStatusSvc[Get VM Status]
    end

    subgraph Internal[Internal Methods]
        initConfig[initConfig<br/>Set Default Values]
        getStorages[getStorages<br/>List Storage Pools]
        getOSImage[getOSImage<br/>Resolve OS Image URL]
        checkOSImageURL[checkOSImageURL<br/>Validate Image URL]
        downloadOSImage[downloadOSImage<br/>Download qcow2]
        createVM[createVM<br/>qm create]
        importDisk[importDisk<br/>qm importdisk]
        initialVM[initialVM<br/>qm set + cloud-init]
        initialWithSSH[initialWithSSH<br/>SSH Init Script]
        CheckAlive[CheckAlive<br/>SSH Polling]
        assignIP[assignIP<br/>Concurrent IP Allocation]
        getCommand[getCommand<br/>Local / SSH Command Selector]
        runCommandSSE[runCommandSSE<br/>SSE-Streamed Command Output]
        checkCPUArch[checkCPUArch<br/>CPU Architecture Detection]
        GetClusterCPUType[GetClusterCPUType<br/>Cluster CPU Type Cache]
        getVMIDsNode[getVMIDsNode<br/>Lookup VM Node Location]
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

**Responsibilities:**
- Implement full VM lifecycle management (create, start, stop, reboot, destroy)
- Encapsulate `qm` and `pvesh` CLI command parameter assembly and execution
- Automatically decide whether to run commands locally or dispatch via SSH to remote nodes
- Manage OS cloud image download and caching
- Concurrently scan available VMID and IP addresses
- Detect CPU architecture compatibility across all cluster nodes and cache results
- Stream command execution progress in real-time via SSE

## Module: Config

```mermaid
graph TB
    subgraph Config
        Routes[NewRoutes<br/>Route Registration]
        CORS[CORS Middleware<br/>Cross-Origin & IP Restriction]
    end

    Routes -->|GET /api/health| Health[Returns ok]
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

**Responsibilities:**
- Register all API routes on the Gin engine
- Configure CORS middleware to restrict access to private IP ranges

## Module: Model

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

**Responsibilities:**
- Define all data transfer objects (DTOs) and internal types
- `Config`: Complete VM provisioning request configuration
- `Response`: Unified API response format
- `SSE`: Server-Sent Events event format
- `VM`: VM status and resource information
- `Node`: Proxmox node resource information
- `Status`: IP availability check result

## Module: Util

```mermaid
graph TB
    subgraph Util
        CheckIP[CheckIP<br/>Source IP Allowlist Check]
        CheckID[CheckID<br/>VMID Existence & State Validation]
        GetVMMap[GetVMMap<br/>Fetch Cluster VM List]
        GetNodeMap[GetNodeMap<br/>Fetch Cluster Node List]
        GetOSUser[GetOSUser<br/>Lookup VM OS User]
        IncludeVM[IncludeVM<br/>VM Filter]
    end

    GetVMMap -->|pvesh get /cluster/resources| ProxmoxCluster[Proxmox Cluster]
    GetNodeMap -->|pvesh get /cluster/resources| ProxmoxCluster
    CheckID --> GetVMMap
```

**Responsibilities:**
- Provide shared helper functions across Service and Handler layers
- Query Proxmox cluster resources via `pvesh` CLI
- Manage the disabled VM list from the `.go_qemu_disabled` file

## Data Flow

### Standard Request Flow

```mermaid
sequenceDiagram
    participant Client
    participant Handler
    participant Service
    participant Proxmox as Proxmox Cluster
    participant VM as Target VM

    Client->>Handler: HTTP Request
    Handler->>Handler: CheckIP (IP Allowlist)
    Handler->>Handler: CheckID (VMID Validation)
    Handler->>Service: Call Corresponding Service Method
    Service->>Service: getVMIDsNode (Lookup VM Node)
    alt Local Node
        Service->>Proxmox: Execute qm / pvesh Locally
    else Remote Node
        Service->>Proxmox: SSH-Dispatched qm Command
    end
    Proxmox-->>Service: Command Output
    Service-->>Handler: Result
    Handler-->>Client: HTTP Response
```

### VM Installation Flow

```mermaid
sequenceDiagram
    participant Client
    participant API as go-pve-qemu
    participant Proxmox
    participant Cloud as Cloud Image Repo
    participant VM as New VM

    Client->>API: POST /api/vm/install (JSON)
    API-->>Client: SSE: Installation Started

    API->>API: Allocate VMID (Concurrent Scan)
    API-->>Client: SSE: VMID Allocated

    API->>API: Assign IP (Gateway + VMID)
    API-->>Client: SSE: IP Assigned

    API->>API: Validate CPU / RAM / Disk Caps
    API-->>Client: SSE: Resources Validated

    API->>Cloud: Download OS Cloud Image
    Cloud-->>API: qcow2 File
    API-->>Client: SSE: Image Downloaded

    API->>Proxmox: qm create (VM Creation)
    Proxmox-->>API: VM Created
    API-->>Client: SSE: VM Created

    API->>Proxmox: qm importdisk (Disk Import)
    Proxmox-->>API: Disk Imported
    API-->>Client: SSE: Disk Imported

    API->>Proxmox: qm set (cloud-init, SSH Keys, Boot Order)
    Proxmox-->>API: VM Configured
    API-->>Client: SSE: VM Initialized

    alt Node Specified
        API->>Proxmox: qm migrate (Node Migration)
        Proxmox-->>API: Migration Complete
        API-->>Client: SSE: Migration Complete
    end

    API->>Proxmox: qm start
    Proxmox-->>API: VM Starting

    loop SSH Polling (up to 60 attempts)
        API->>VM: SSH Connection Test
        alt Connection Successful
            VM-->>API: ready
        else Connection Failed
            API->>API: Wait 5 seconds
        end
    end

    API->>VM: SSH Execute Init Script
    VM-->>API: Initialization Complete
    API-->>Client: SSE: SSH Init Complete

    API->>Proxmox: qm reboot
    Proxmox-->>API: VM Rebooting

    loop SSH Polling (up to 60 attempts)
        API->>VM: SSH Connection Test
        alt Connection Successful
            VM-->>API: ready
        else Connection Failed
            API->>API: Wait 5 seconds
        end
    end

    API-->>Client: SSE: VM Installation Complete (VMID, IP, User)
    API-->>Client: event: close
```

## VM Lifecycle State Machine

```mermaid
stateDiagram-v2
    [*] --> Creating: POST /api/vm/install
    Creating --> Created: qm create + importdisk Complete
    Created --> Running: qm start
    Running --> Stopped: qm stop / shutdown
    Running --> Rebooting: POST /api/vm/:id/reboot
    Rebooting --> Running: SSH Ready Again
    Running --> Migrating: POST /api/vm/:id/set/node
    Migrating --> Running: Migration Complete
    Stopped --> Running: POST /api/vm/:id/start
    Stopped --> Destroyed: POST /api/vm/:id/destroy
    Running --> Destroyed: POST /api/vm/:id/destroy
    Destroyed --> [*]
```

---

©️ 2025 [邱敬幃 Pardn Chiu](https://www.linkedin.com/in/pardnchiu)