# WSS CTF Challenge Platform

A Docker-based Capture The Flag (CTF) challenge platform with an interactive menu system for selecting and running security challenges.

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
- Display an interactive menu to select challenges
- Build Docker images as needed (if not already built)
- Run selected challenges and accept flag submissions

### Command Line Flags

- `--build` - Force rebuild of all challenge Docker images
  ```bash
  ./start-challenges --build
  ```

- `--clean` - Remove all challenge containers and images
  ```bash
  ./start-challenges --clean
  ```

- `--debug` - Show verbose output including Docker operations
  ```bash
  ./start-challenges --debug
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
  "hints": [
    "First hint - basic guidance",
    "Second hint - more specific",
    "Final hint - very specific"
  ],
  "port": 8080,
  "preface": "Optional introduction text shown before challenge starts",
  "postface": "Optional congratulations text shown after completion"
}
```

**Fields:**
- `name` - Display name of the challenge
- `flag` - The correct flag to complete the challenge
- `hints` - Array of progressive hints (optional)
- `port` - Host port to map the challenge container to
- `preface` - Text shown when starting the challenge (optional)
- `postface` - Text shown after successful completion (optional)

## Challenge Interaction

### Main Menu
- **Select challenge**: Type a number (1, 2, etc.) to start a challenge
- **Exit platform**: Type `quit` or `exit`

### During Challenge
- **Submit flag**: Type the flag and press Enter
- **Get hint**: Type `hint` (reveals hints progressively)
- **Return to menu**: Type `menu`

### Security Features
- **Ctrl+C Protection**: Ctrl+C is disabled to prevent accidental termination
- **Graceful cleanup**: All containers are properly stopped and cleaned up

## Docker Image Management

- Images are cached after first build for faster subsequent runs
- Use `--build` flag to force rebuild when challenge files change
- Use `--clean` flag to remove all challenge resources

## Requirements for Challenge Creation

Each challenge directory must contain:
1. `Dockerfile` - Instructions to build the challenge container
2. `challenge.json` - Challenge metadata (name, flag, hints, port, etc.)
3. Challenge application that runs on port 80 inside the container

## Notes

- Challenges can be run in any order through the interactive menu
- Each challenge container is mapped from internal port 80 to the specified host port
- Containers are automatically cleaned up when returning to menu or completing challenges
- Use `--debug` flag to see detailed Docker operations for troubleshooting