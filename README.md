> [!Note]
> This content is translated by LLM. Original text can be found [here](README.zh.md)

# Go Qemu - Proxmox VM API

> Go Qemu is a Proxmox VE virtual machine management API service developed in Go, **automatically downloads official images and completes end-to-end deployment**. Supports **one-click automated deployment** for Debian, Ubuntu, and RockyLinux.

[![pkg](https://pkg.go.dev/badge/github.com/pardnchiu/go-qemu.svg)](https://pkg.go.dev/github.com/pardnchiu/go-qemu)
[![version](https://img.shields.io/github/v/tag/pardnchiu/go-qemu?label=release)](https://github.com/pardnchiu/go-qemu/releases)
[![license](https://img.shields.io/github/license/pardnchiu/go-qemu)](LICENSE)<br>
[![readme](https://img.shields.io/badge/readme-EN-white)](README.md)
[![readme](https://img.shields.io/badge/readme-ZH-white)](README.zh.md)

## Key Features

### Complete Virtual Machine Lifecycle Management
- Support for multiple Linux distributions (**automatic official image download and deployment**)
- Real-time SSE streaming installation progress feedback
- Intelligent IP address allocation and management
- Complete SSH key configuration

### Multi-node Cluster Support
- Support for Proxmox VE cluster environments
- Remote node SSH operation support
- Unified API for multi-node requests

### Zero-configuration Deployment
- Automatically downloads the latest cloud images from official sources, no manual image preparation required
  - Debian: download from `cloud.debian.org`
  - RockyLinux: download from `dl.rockylinux.org`
  - Ubuntu: download from `cloud-images.ubuntu.com`

## Usage

### Health Check
```
GET /api/health
```

### Virtual Machine Management

#### Create Virtual Machine
> Supported OS
> - Debian: 11, 12, 13
> - RockyLinux: 8, 9, 10
> - Ubuntu: 20.04, 22.04, 24.04
```
POST /api/vm/install
```

```json
{
  "id": 101,                                          // Optional, VM ID, auto-assigned if not specified
  "name": "test-vm",                                  // VM name
  "node": "pve1",                                     // Optional, specify node name
  "os": "ubuntu",                                     // Required, supports: debian, ubuntu, rockylinux
  "version": "22.04",                                 // Required, OS version
  "cpu": 2,                                           // vCPU cores, default 2
  "ram": 2048,                                        // RAM size (MB), default 2048
  "disk": "20G",                                      // Disk size, default 16G
  "passwd": "password123",                            // SSH password, default "passwd"
  "pubkey": "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5..."     // Optional, SSH public key
}
```

**SSE Response**
```javascript
data: {"step":"preparation > checking VMID","status":"processing","message":"[*] using specified VMID: 101"}
data: {"step":"VM creation > creating VM","status":"success","message":"[+] VM created successfully (2.45s)"}
```

#### Start Virtual Machine
```
POST /api/vm/{id}/start
```

**SSE Response**
```javascript
data: {"step":"starting VM","status":"success","message":"[+] VM starting (1.23s)"}
data: {"step":"waiting for SSH","status":"success","message":"[+] VM is ready (15.67s)"}
```

#### Reboot Virtual Machine
```
POST /api/vm/{id}/reboot
```

**SSE Response**
```javascript
data: {"step":"rebooting VM","status":"success","message":"[+] VM rebooting (1.23s)"}
data: {"step":"waiting for SSH","status":"success","message":"[+] VM is ready (15.67s)"}
```

#### Virtual Machine Status
```
GET /api/vm/{id}/status
```

#### Shutdown Virtual Machine
```
POST /api/vm/{id}/shutdown
```

#### Stop Virtual Machine (Force shutdown)
```
POST /api/vm/{id}/stop
```

#### Remove Virtual Machine
```
POST /api/vm/{id}/destroy
```

## Environment Configuration

### Required Environment Variables
```bash
PORT=8080

# Main node name
MAIN_NODE=PVE1

# Network gateway
GATEWAY=10.0.0.1

# Allowed access IPs (comma-separated, 0.0.0.0 allows all)
ALLOW_IPS=0.0.0.0

# Node IP configuration
NODE_PVE1=10.0.0.2
NODE_PVE2=10.0.0.3
NODE_PVE3=10.0.0.4

# IP assignment range
ASSIGN_IP_START=100
ASSIGN_IP_END=254

# VM root password
# Default is 8 spaces (no password)
VM_ROOT_PASSWORD=
```

### Initialization Scripts

The system automatically executes corresponding initialization scripts:
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

## Usage Examples

### Create Virtual Machine
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

### Manage Virtual Machine
```bash
# Get status
curl http://localhost:8080/api/vm/101/status

# Start
curl -X POST http://localhost:8080/api/vm/101/start

# Stop
curl -X POST http://localhost:8080/api/vm/101/stop

# Reboot
curl -X POST http://localhost:8080/api/vm/101/reboot

# Destroy
curl -X POST http://localhost:8080/api/vm/101/destroy
```

## License

This source code project is licensed under the [MIT](LICENSE) license.

## Author

<img src="https://avatars.githubusercontent.com/u/25631760" align="left" width="96" height="96" style="margin-right: 0.5rem;">

<h4 style="padding-top: 0">邱敬幃 Pardn Chiu</h4>

<a href="mailto:dev@pardn.io" target="_blank">
  <img src="https://pardn.io/image/email.svg" width="48" height="48">
</a> <a href="https://linkedin.com/in/pardnchiu" target="_blank">
  <img src="https://pardn.io/image/linkedin.svg" width="48" height="48">
</a>

***

©️ 2025 [邱敬幃 Pardn Chiu](https://pardn.io)
