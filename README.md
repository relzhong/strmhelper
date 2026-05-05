# StrmHelper

StrmHelper is a powerful, lightweight tool designed to synchronize media from OpenList/AList servers to your local library as `.strm` files. It bridges the gap between cloud storage and local media players (like Jellyfin, Emby, or Plex), allowing you to stream your cloud library with minimal local disk usage.

## ✨ Key Features

- **Efficient Scanning**: Leverages directory modification times to skip unchanged folders, significantly reducing API overhead.
- **Smart Protection**: Prevents mass deletion of local files if the server becomes temporarily unavailable.
- **Sync Back Delete**: Optionally synchronize local file deletions back to the remote server, keeping your library tidy.
- **Grace Period**: Safety mechanism to prevent accidental deletions during transient network or mount failures.
- **Multi-lingual UI**: Modern, clean administrative interface with support for English and Chinese.
- **BDMV Support**: Intelligent detection and processing of Blu-ray folder structures.
- **Docker Ready**: Cross-platform support (amd64, arm64) via GHCR.

## 🚀 Quick Start with Docker

The easiest way to run StrmHelper is using Docker.

```bash
docker run -d \
  --name strmhelper \
  -p 8080:8080 \
  -v ./config:/app/config \
  -v /path/to/your/media:/media \
  -e TZ=Asia/Shanghai \
  ghcr.io/relzhong/strmhelper:latest
```

### Docker Compose

```yaml
version: '3'
services:
  strmhelper:
    image: ghcr.io/relzhong/strmhelper:latest
    container_name: strmhelper
    ports:
      - "8080:8080"
    volumes:
      - ./config:/app/config
      - /your/local/media/path:/media
    environment:
      - TZ=Asia/Shanghai
    restart: unless-stopped
```

## ⚙️ Configuration

1. **Access the Dashboard**: Open `http://localhost:8080` in your browser.
2. **Login**: Default credentials (if not configured in `config.yaml`) are usually `admin` / `admin`.
3. **Add a Task**:
   - **Source Directory**: The path on your OpenList/AList server (e.g., `/115/movies`).
   - **Target Directory**: The local path where `.strm` files will be generated (must be accessible to the container).
   - **Sync Back Delete**: Enable this if you want local deletions to trigger remote deletions.
   - **Directory Time Check**: Enable this for much faster scans on supported storage providers.

## 🛡️ Safety Features

### Smart Protection
If the number of files detected for deletion exceeds your configured **Threshold**, the scan will be aborted. This protects your library from being wiped if the server returns an empty list due to a glitch.

### Grace Scans
When a file is detected as missing, it enters a "Pending" state. It will only be deleted (or synced back) after it remains missing for the specified number of **Grace Scans**.

## 🛠️ Development

If you wish to build from source:

```bash
# Clone the repository
git clone https://github.com/relzhong/strmhelper.git
cd strmhelper

# Build the binary
make build

# Run locally
./bin/strmhelper
```

## 📄 License

AGPL-3.0 License. See `LICENSE` for more details.
