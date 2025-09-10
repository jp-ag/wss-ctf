# Creating New Challenges
Follow the steps below to create a Capture the Flag challenge in the **WSS CTF Challenge Platform**.
## Prerequisites
1. You have followed the steps described in [First Steps (README\first-steps.md)].

### Running Challenges
Run the command below to connect to *Docker daemon* and load the `challenges/config.json`.
```bash
./start-challenges
```
**Result**: The **Main Menu** is displayed and you can now select a challenge. Challenges can be run in any order.

## Creating Challenges 
The platform comes with two template challenges by default. You can edit the templates or create new challenges using the resources below.

### Directory layout
<details>
<summary>The layout defines the challenge structure and how it is shown to the user. You can edit the layout according to your challenges.</summary>

**Default directory layout**
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
<summary>The <i>challenge.json</i> defines names, flags, and hints of your challenges. You can edit the metadata according to your needs and preferences.</summary>

**List of fields:**
- `name` - Defines the display name of the challenge.
- `flag` - Defines the flag used to complete the challenge.
- `hints` (*Optional*) - Defines the array of progressive hints. 
- `port` - Defines the host port to map the challenge containers port to.
- `preface` (*Optional*) - Defines the text shown at start of a challenge.
- `postface` (*Optional*) - Defines the text shown at the end of a challenge.

**Default *challenge.json***:
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
</details>

### Assigning a number 
<details>
<summary>The <i>config.json</i> assigns a number to your challenge. You can edit the metadata according to your needs and preferences.</summary>

> **Important:** Use the same challenge name as used in the `name` field inside the *challenge.json*.

**Default *config.json***:
```json
{
  "challenges": ["01-first-chal", "02-second-chal"]
}
```
</details>

## Docker Image and Container Management
<details>
<summary> Learn more about the challenges containers and images, and how to edit them.</summary>

**Important information**
- All images are cached after the first build for faster subsequent runs.
- Containers are automatically cleaned up when returning to menu or completing challenges.
### Command line flags - Optional
You can use the commands listed below to edit your images and containers.

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
