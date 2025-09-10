# Creating New Challenges
Follow the steps below to create a Capture the Flag challenge in the **WSS CTF Challenge Platform**.

## Prerequisites
1. You have followed the steps described in [First Steps (README\first-steps.md)].


## Creating Challenges 
The platform comes with two template challenges by default. Edit the templates and create new challenges using the resources below.

### Directory layout
<details>
<summary>The layout defines the challenge structure and how it is shown to the player. You can edit the layout according to your challenges.</summary>

The default *directory layout* is shown below:
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
</details>

### Creating Metadata
<details>
<summary>The <i>challenge.json</i> defines names, flags, and hints of your challenges. You can edit the metadata according to your needs.</summary>

The default *challenge.json* file is shown below:
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
1. Add your metadata by editing the fields listed below.
**List of fields:**
- `name` - Defines the display name of the challenge.
- `flag` - Defines the flag used to complete the challenge.
- `hints` (*Optional*) - Defines the array of progressive hints. 
- `port` - Defines the host port to map the challenge containers port to.
- `preface` (*Optional*) - Defines the text shown at start of a challenge.
- `postface` (*Optional*) - Defines the text shown at the end of a challenge.

</details>

### Assigning a number 
<details>
<summary>The <i>config.json</i> assigns a number to your challenge. You can edit the metadata according to your needs.</summary>

> **Important:** Use the same challenge name as used in the `name` field inside the *challenge.json*.

The default *config.json* file is shown below::
```json
{
  "challenges": ["01-first-chal", "02-second-chal"]
}
```
</details>

## Docker Image and Container Management
<details>
<summary>Commands for managing images and containers</summary>

### Command Line Flags
Use the commands below to manage images and containers.

####  `--build`

Use `--build` force a rebuild of all **Docker** images.
  ```bash
  ./start-challenges --build
  ```

#### `--clean` 
Use `--clean` to remove all challenge containers and images.
  ```bash
  ./start-challenges --clean
  ```
**Important information**
- All images are cached after the first build for faster subsequent runs.
- Containers are automatically cleaned up when returning to menu or completing challenges.
</details>

### Troubleshooting
<details>
<summary>Learn how to generate a detailed output.</summary>

Use `--debug` to generate complete output, including **Docker** operations.
#### `--debug` 
  ```bash
  ./start-challenges --debug
  ```
</details>
