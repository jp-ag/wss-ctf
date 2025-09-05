# WSS CTF Challenge Platform
Use the **WSS CTF Challenge Platform** to create Docker-based Capture The Flag (CTF) challenges using Linux. With this platform, you can create interactive menu systems for selecting and running cybersecurity challenges.

## Prerequisites
You are using **Linux** or running a Linux-like terminal.
### For building the application: 
- You have installed  **Go** 1.19 or a higher version.
### For running challenge containers
- You have installed **Docker**. 
- **Docker daemon** is active.


## Installation

1. Open your **Linux** terminal.
2. Clone this repository [Git clone](https://github.com/jp-ag/wss-ctf.git 'Git clone')
3. Open the directory:
```cd wss-ctf
```
4. Build the application:
   ```bash
   go build -o start-challenges .
   ```


## Use

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

### Command Line Flags - Optional

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

## Challenge Structure - Non optional 
Já vem com dois desafios de template, você pode edita-los ou deleta-los

### Directory Layout
The layout comes pre-configured as shown below. You can edit the layout according to your challenges.
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
The *config.json* defines the order of your challenges. You can edit the order according to your challenges.
Pre-configured
```json
{
  "challenges": ["01-first-chal", "02-second-chal"]
}
```

### challenge.json
The *challenge.json* defines name, flags, and hints of your challenges. You can edit the metadata according to your challenges.
Pre-configured metadata:
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
- `name` - Display name of the challenge.
- `flag` - The correct flag to complete the challenge.
- `hints` - *Optional* Array of progressive hints.
- `port` - Host port to map the challenge containers port to.
- `preface` - *Optional* Text shown when starting the challenge.
- `postface` - *Optional* Text shown after successful completion.

## Docker Image Management

- For faster subsequent runs, images are cached after first build.
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

## How to play
Use the commands listed below to play a challenge. 
### Main Menu
- **To start a challenge**: Type the challenge's number listed on the menu to start.
- **To exit the platform**: Type `quit` or `exit` to exit.

### During a Challenge
- **To submit a flag**: Type the contents from the flag's *.txt* file and press Enter.
- **To get a hint**: Type `hint` to reveal hints.
- **To return to the Main Menu**: Type `menu`. Returning to **Main Menu** ends the challenge. 

### Security Features
Security features are automatically activated to prevent issues.
- **Ctrl+C Protection**: Ctrl+C is disabled to prevent accidental termination.
- **Graceful cleanup**: All containers are properly stopped and cleaned up once you leave the challenge. 

