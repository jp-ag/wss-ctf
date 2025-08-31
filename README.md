# WSS CTF Challenge Platform

A Docker-based Capture The Flag (CTF) challenge platform that manages and runs security challenges sequentially.

## Prerequisites

- **Go** (1.19 or higher) - For building the application
- **Docker** - For running challenge containers
- Docker daemon must be running

## Installation

1. Clone the repository
2. Build the application:
   ```bash
   go build -o start-challenges .
   ```

## Usage

### Running Challenges
```bash
./start-challenges
```
This will:
- Connect to Docker daemon
- Load challenges from `challenges/config.json`
- Build Docker images for each challenge (if not already built)
- Run challenges sequentially
- Accept flag submissions and provide hints

### Command Line Flags

- `--build` - Force rebuild of all challenge Docker images
  ```bash
  ./start-challenges --build
  ```

- `--clean` - Remove all challenge containers and images
  ```bash
  ./start-challenges --clean
  ```

## Challenge Structure

### Directory Layout
```
wss-ctf/
├── main.go
├── start-challenges (compiled binary)
└── challenges/
    ├── config.json
    └── [challenge-directories]/
        ├── challenge.json
        ├── Dockerfile
        └── [challenge files]
```

### config.json
Defines the order of challenges:
```json
{
  "challenges": ["01-first-chal", "02-second-chal"]
}
```

### challenge.json
Metadata for each challenge:
```json
{
  "name": "Challenge Name",
  "flag": "FLAG{example}",
  "hint": "Helpful hint text",
  "port": 8080
}
```

## Challenge Interaction

- **Submit flag**: Type the flag and press Enter
- **Get hint**: Type `hint`
- **Quit challenge**: Type `quit` or `exit`

## Docker Image Management

- Images are cached after first build for faster subsequent runs
- Use `--build` flag to force rebuild when challenge files change
- Use `--clean` flag to remove all challenge resources

## Requirements for Challenge Creation

Each challenge directory must contain:
1. `Dockerfile` - Instructions to build the challenge container
2. `challenge.json` - Challenge metadata (name, flag, hint, port)
3. Challenge application that runs on port 80 inside the container

## Notes

- Challenges run sequentially - you must complete or quit current challenge to proceed
- Each challenge container is mapped from internal port 80 to the specified host port
- Containers and images are automatically cleaned up after completion