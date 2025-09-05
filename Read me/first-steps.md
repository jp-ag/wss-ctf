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
