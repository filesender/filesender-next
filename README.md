# FileSender Next

FileSender Next is a fresh rewrite of the widely-used FileSender application, now implemented in Go.

## Why a Rewrite?

During the last FileSender meeting at the TNC24 conference, it was concluded that working towards a more secure FileSender codebase should be the priority for the roadmap. After more than 12 years of development, the original FileSender 2.x codebase has grown in complexity.

This rewrite aims to:

- Improve security
- Simplify the architecture
- Enable new features, including end-to-end encryption for large files

For more information, see the [official update from FileSender](https://filesender.org/filesender-online-infoshare-update-on-release-3-0-and-security-approach/).

## Current Status

This repository represents the initial stage of the rewrite effort, aiming to deliver an MVP that captures the core FileSender functionality with improved security and simplicity.

## Setup

### Development Setup (Local)

1. **Install Dependencies**:
    - [Go](https://golang.org/dl/) (1.24 or newer)
    - [Make](https://www.gnu.org/software/make/)
        - **Windows users** can install Make via [Gow](https://github.com/bmatzelle/gow), [Chocolatey](https://chocolatey.org/packages/make), or WSL.

2. **Clone the Repository**:

    ```sh
    git clone https://codeberg.org/filesender/filesender-next.git
    cd filesender-next
    ```

3. **Run the Application**:

    ```sh
    make run-dev
    ```

    This uses a dummy authentication method and stores data locally in `./data`.

4. **Other Useful Commands**:
    - `make test` Run all tests with coverage
    - `make lint` Run linter
    - `make fmt` Format code
    - `make clean` Remove built files
    - `make install` Install binary to `/usr/local/bin/filesender`
    - `make hotreload` Run with hot reloading (requires `watchexec`)

### Docker Setup

If you prefer containerized deployment:

1. **Install Dependencies**:
    - [Docker](https://www.docker.com/) or [Podman](https://podman.io/)

2. **Build the Docker Image**:

    ```sh
    docker build -t filesender:latest .
    ```

3. **Run the Container**:

    ```sh
    docker run -p 8080:8080 filesender:latest
    ```

    By default, the container:

    - Listens on port `8080`
    - Uses a dummy authentication method
    - Stores state in `/app/data`

#### Environment Variables

You can configure behavior by passing environment variables when running the container:

- `FILESENDER_AUTH_METHOD` Sets the authentication method (default: `dummy`)
- `STATE_DIRECTORY` Directory for storing internal state (default: `/app/data`)
- `MAX_UPLOAD_SIZE` Maximum file upload size in bytes (default: `2147483648`, 2GB)

Example with custom configuration:
```sh
docker run -p 8080:8080 \
    -e FILESENDER_AUTH_METHOD=dummy \ 
    -e MAX_UPLOAD_SIZE=4294967296 \ 
    -e STATE_DIRECTORY=/app/data \ 
    filesender:latest
```

## Repositories

- Primary development occurs on [Codeberg](https://codeberg.org/filesender/filesender-next).
- Mirrored repository on [GitHub](https://github.com/filesender/filesender-next).
