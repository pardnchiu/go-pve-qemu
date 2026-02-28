# go-pve-qemu — Technical Documentation

> [Chinese version](doc.zh.md)

## Table of Contents

- [Deployment](#deployment)
- [Configuration](#configuration)
- [OS Support](#os-support)
- [API Reference](#api-reference)
- [SSE Event Format](#sse-event-format)
- [Install Pipeline Stages](#install-pipeline-stages)

---

## Deployment

### Prerequisites

- Go ≥ 1.20 installed on the Proxmox main node
- `qm` and `pvesh` CLI available (standard on Proxmox VE)
- SSH access from main node to all remote nodes (key-based, no passphrase)
- Network gateway that places VMID == last IP octet (e.g. `192.168.0.X` where `X` = VMID)

### Build & Run

```bash
git clone https://github.com/pardnchiu/go-pve-qemu.git
cd go-pve-qemu
cp .env.example .env
# configure .env (see Configuration section)
go build -o go-pve-qemu ./cmd/api
./go-pve-qemu
```

The server prints all registered routes on startup and listens on `PORT`.

### Systemd Service (optional)

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

---

## Configuration

All configuration is via environment variables (loaded from `.env` on startup via `godotenv`).

### Required

| Variable | Description |
|----------|-------------|
| `PORT` | API server listen port |
| `GATEWAY` | Network gateway — drives IP allocation (e.g. `192.168.0.1`) |
| `MAIN_NODE` | Primary Proxmox node name (commands run locally on this node) |
| `NODE_<name>` | IP address per cluster node (e.g. `NODE_pve2=192.168.0.12`) |
| `ASSIGN_STORAGE` | Proxmox storage pool for disk import (e.g. `local-zfs`) |

### Optional

| Variable | Default | Description |
|----------|---------|-------------|
| `ASSIGN_IP_START` | `100` | IP range start — last octet (minimum `100`, VMID must be ≥ 100) |
| `ASSIGN_IP_END` | `254` | IP range end — last octet (maximum `254`) |
| `ALLOW_IPS` | `0.0.0.0` | Comma-separated list of allowed source IPs; `0.0.0.0` allows all |
| `VM_MAX_CPU` | unlimited | Maximum vCPU cores per VM (1–32) |
| `VM_MAX_RAM` | unlimited | Maximum RAM per VM in MB |
| `VM_MAX_DISK` | unlimited | Maximum disk size per VM in GB |
| `VM_BALLOON_MIN` | — | Minimum RAM for balloon device in MB |
| `VM_ROOT_PASSWORD` | — | Default root password injected via OS init script |

### Example `.env`

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

---

## OS Support

| OS | Supported Versions | Image Source |
|----|-------------------|--------------|
| Debian | 11, 12, 13 | `cloud.debian.org` |
| Ubuntu | 20.04, 22.04, 24.04 | `cloud-images.ubuntu.com` |
| RockyLinux | 8, 9, 10 | `dl.rockylinux.org` |

OS cloud images are downloaded on first use and cached locally. Subsequent installs of the same OS/version reuse the cached image.

OS init scripts are served statically at `/sh/<os>_<version>.sh` and fetched inside the VM during SSH initialization.

---

## API Reference

All endpoints are prefixed with `/api`.

| Method | Path | Response Type | Description |
|--------|------|---------------|-------------|
| GET | `/api/health` | `text/plain` | Health check — returns `ok` |
| POST | `/api/vm/install` | SSE | Create and fully provision a VM |
| GET | `/api/vm/list` | JSON | List all VMs with per-node cluster stats |
| GET | `/api/vm/:id/status` | `text/plain` | VM power status (`running` or `stopped`) |
| POST | `/api/vm/:id/start` | SSE | Start VM, stream until SSH ready |
| POST | `/api/vm/:id/stop` | `text/plain` | Force-stop VM immediately |
| POST | `/api/vm/:id/shutdown` | `text/plain` | Graceful ACPI shutdown |
| POST | `/api/vm/:id/reboot` | SSE | Reboot VM, stream until SSH ready |
| POST | `/api/vm/:id/destroy` | `text/plain` | Delete VM and purge disk |
| POST | `/api/vm/:id/set/cpu` | `text/plain` | Set vCPU count |
| POST | `/api/vm/:id/set/memory` | `text/plain` | Set RAM size |
| POST | `/api/vm/:id/set/disk` | `text/plain` | Expand disk |
| POST | `/api/vm/:id/set/node` | SSE | Migrate VM to another cluster node |

---

### `GET /api/health`

Returns `200 OK` with body `ok`. Used for liveness probes.

---

### `POST /api/vm/install`

Provisions a complete VM end-to-end. Streams SSE events throughout.

**Request body (JSON)**

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `os` | string | Yes | — | `debian`, `ubuntu`, or `rockylinux` |
| `version` | string | Yes | — | OS version (e.g. `22.04`, `12`, `9`) |
| `name` | string | No | auto | VM name |
| `id` | int | No | auto | VMID (100–254); auto-allocated if omitted |
| `node` | string | No | main node | Target Proxmox node |
| `cpu` | int | No | `2` | vCPU cores (clamped to `VM_MAX_CPU`) |
| `ram` | int | No | `2048` | RAM in MB, minimum 512 (clamped to `VM_MAX_RAM`) |
| `disk` | string | No | `16G` | Disk size (e.g. `20G`), minimum `16G` |
| `user` | string | No | — | SSH username |
| `passwd` | string | No | — | SSH password |
| `pubkey` | string | No | — | SSH public key to inject |

**Example**

```json
{
  "os": "ubuntu",
  "version": "22.04",
  "name": "web-01",
  "cpu": 2,
  "ram": 4096,
  "disk": "40G",
  "user": "deploy",
  "passwd": "s3cr3t",
  "pubkey": "ssh-ed25519 AAAA..."
}
```

**SSE stream** — see [SSE Event Format](#sse-event-format)

---

### `GET /api/vm/list`

Returns cluster-wide VM inventory and per-node resource utilization.

**Response (JSON)**

```json
{
  "success": true,
  "vms": [
    {
      "vmid": 101,
      "name": "web-01",
      "os": "ubuntu",
      "running": true,
      "node": "pve1",
      "cpu": 2,
      "disk": 40,
      "memory": 4096,
      "memory_used": 1024
    }
  ],
  "nodes": [
    {
      "node": "pve1",
      "max_cpu": 32,
      "max_memory": 131072,
      "cpu": 0.12,
      "memory": 0.31,
      "memory_used": 40960,
      "disk": 0.45,
      "running": true
    }
  ]
}
```

---

### `GET /api/vm/:id/status`

Returns `running` or `stopped` as plain text.

---

### `POST /api/vm/:id/start`

Starts the VM and streams SSE until SSH is responsive.

---

### `POST /api/vm/:id/stop`

Force-stops the VM (equivalent to `qm stop --skiplock`). Returns plain text result.

---

### `POST /api/vm/:id/shutdown`

Sends ACPI shutdown signal. Returns plain text result.

---

### `POST /api/vm/:id/reboot`

Reboots the VM and streams SSE until SSH is responsive again.

---

### `POST /api/vm/:id/destroy`

Stops and permanently deletes the VM including all disk data (`--purge`). Returns plain text result.

---

### `POST /api/vm/:id/set/cpu`

**Request body**

```json
{ "cpu": 4 }
```

Valid range: 1–32 (further capped by `VM_MAX_CPU`). Returns plain text result.

---

### `POST /api/vm/:id/set/memory`

**Request body**

```json
{ "memory": 8192 }
```

Value in MB. Minimum 512, maximum `VM_MAX_RAM`. Returns plain text result.

---

### `POST /api/vm/:id/set/disk`

**Request body**

```json
{ "disk": "10G" }
```

Expands disk by the specified amount. Disk shrinking is not supported. Returns plain text result.

---

### `POST /api/vm/:id/set/node`

Migrates the VM to another cluster node with local disk transfer. Streams SSE progress.

**Request body**

```json
{ "node": "pve2" }
```

---

## SSE Event Format

All SSE endpoints emit `data:` lines with JSON payloads.

```
data: {"step":"<stage>","status":"<status>","message":"<text>","vm_id":<id>,"ip":"<ip>"}
```

| Field | Values | Description |
|-------|--------|-------------|
| `step` | string | Pipeline stage name (e.g. `"preparation > checking VMID"`) |
| `status` | `processing` \| `success` \| `error` | Event severity |
| `message` | string | Human-readable detail with elapsed time |
| `vm_id` | int | VMID (present on final success event) |
| `ip` | string | Assigned IP (present on final success event) |

**Example stream (install)**

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

---

## Install Pipeline Stages

The `POST /api/vm/install` pipeline executes the following stages in order:

| Stage | Description |
|-------|-------------|
| `preparation > checking VMID` | Auto-allocate or validate specified VMID |
| `preparation > assigning IP` | Derive static IP from gateway prefix + VMID |
| `preparation > validating CPU and RAM` | Clamp resources to configured limits |
| `preparation > validating disk size` | Enforce minimum 16 G and maximum `VM_MAX_DISK` |
| `preparation > setting default config values` | Apply storage and CPU-type defaults |
| `preparation > checking storage pool` | Verify storage pool exists and is active |
| `OS preparation > getting OS image` | Resolve official cloud image URL |
| `OS preparation > validating OS image URL` | HTTP HEAD check on image URL |
| `OS preparation > downloading OS image` | Download if not cached locally |
| `SSH preparation > checking SSH key` | Verify or generate host SSH key pair |
| `VM creation > creating VM` | `qm create` with cloud-init hardware profile |
| `VM creation > importing disk image` | `qm importdisk` into storage pool |
| `VM initialization > initializing configuration` | Attach disk, set boot order, cloud-init drive |
| `VM initialization > migrating VM` | Migrate to target node (if `node` specified) |
| `VM initialization > waiting for ready` | Poll SSH until VM is accessible |
| `VM initialization > SSH initialization` | Execute OS-specific init script via SSH |
| `VM initialization > rebooting VM` | Reboot to apply all settings |
| `VM initialization > finalizing` | Confirm SSH ready post-reboot, emit VMID/IP |

---

©️ 2025 [邱敬幃 Pardn Chiu](https://linkedin.com/in/pardnchiu)
