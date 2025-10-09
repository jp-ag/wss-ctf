package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"os/exec"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// Config represents the main configuration file (config.json)
// that defines the order of challenges.
type Config struct {
	Challenges []string `json:"challenges"`
}

// Challenge represents the metadata for a single challenge (challenge.json).
type Challenge struct {
    Name     string   `json:"name"`
    Flag     string   `json:"flag"`     // Single flag (backward compatible)
    Flags    []string `json:"flags"`    // Multiple flags (new)
    Hints    []string `json:"hints"`
    Ports    []int    `json:"ports"`    // Mude para Ports e tipo []int
    Preface  string   `json:"preface"`
    Postface string   `json:"postface"`
}


func main() {
	// Define and parse the --build command-line flag.
	build := flag.Bool("build", false, "Force rebuild of all challenge images")
	clean := flag.Bool("clean", false, "Remove all challenge images and containers")
	debug := flag.Bool("debug", false, "Show verbose output including Docker operations")
	flag.Parse()

	// Set up signal handler to catch Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Launch goroutine to handle interrupt signals
	go func() {
		for range sigChan {
			fmt.Println("\n⚠️  Ctrl+C disabled. Please type 'quit' or 'exit' to shut down properly.")
		}
	}()

	// Create a new Docker client.
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Error: Could not create Docker client. Is Docker running? Details: %v", err)
	}

	// Check if Docker is running by pinging the daemon.
	_, err = cli.Ping(ctx)
	if err != nil {
		log.Fatalf("Error: Could not connect to Docker daemon. Please make sure Docker is running. Details: %v", err)
	}

	// Handle --clean flag
	if *clean {
		fmt.Println("Cleaning up all challenge images and containers...")
		cleanAll(ctx, cli)
		fmt.Println("All challenge resources have been removed.")
		return
	}

	fmt.Println("#######################################")
	fmt.Println("## Welcome to the Challenge Platform ##")
	fmt.Println("#######################################")

	// Load the main configuration file.
	configFile, err := os.ReadFile("/wss-ctf/challenges/config.json")
	if err != nil {
		log.Fatalf("Error: Could not read /wss-ctf/challenges/config.json. Details: %v", err)
	}

	var config Config
	if err := json.Unmarshal(configFile, &config); err != nil {
		log.Fatalf("Error: Could not parse /wss-ctf/challenges/config.json. Details: %v", err)
	}

	// Start with first challenge directly
	// Run first challenge (01-first-chal)
	if len(config.Challenges) > 0 {
		firstChallenge := config.Challenges[0]
		result := runChallenge(ctx, cli, firstChallenge, *build, *debug, false)

		// After first challenge, check if we should continue to second
		if result == "continue" && len(config.Challenges) > 1 {
			// Start second challenge silently
			secondChallenge := config.Challenges[1]
			runChallenge(ctx, cli, secondChallenge, *build, *debug, true)
		}
	}

	fmt.Println("\nChallenge session ended. Goodbye!")
}

func fileExists(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}

// runChallenge acts as a router, detecting the challenge type and calling the appropriate handler.
func runChallenge(ctx context.Context, cli *client.Client, dirName string, forceBuild bool, debug bool, silent bool) string {
    challengePath := filepath.Join("/wss-ctf/challenges", dirName)
    composePath := filepath.Join(challengePath, "docker-compose.yml")
    dockerfilePath := filepath.Join(challengePath, "Dockerfile")

    if fileExists(composePath) {
        // This is a Docker Compose-based challenge
        return runComposeChallenge(ctx, cli, challengePath, debug, silent)
    } else if fileExists(dockerfilePath) {
        // This is a Dockerfile-based challenge
        return runDockerfileChallenge(ctx, cli, dirName, forceBuild, debug, silent)
    } else {
        log.Printf("Error: No Dockerfile or docker-compose.yml found for challenge '%s'", dirName)
        return "menu"
    }
}

// runComposeChallenge handles challenges defined by a docker-compose.yml file.
func runComposeChallenge(ctx context.Context, cli *client.Client, challengePath string, debug bool, silent bool) string {
    // Load challenge metadata, which is common for all challenge types
    challengeFile, err := os.ReadFile(filepath.Join(challengePath, "challenge.json"))
    if err != nil {
        log.Printf("Error: Could not read challenge.json in %s. Details: %v", challengePath, err)
        return "menu"
    }
    var challenge Challenge
    if err := json.Unmarshal(challengeFile, &challenge); err != nil {
        log.Printf("Error: Could not parse challenge.json in %s. Details: %v", challengePath, err)
        return "menu"
    }

    if !silent {
        fmt.Printf("\n--- Starting Challenge: %s ---\n", challenge.Name)
        fmt.Println("Detected docker-compose.yml, starting environment...")
    }

    // --- Docker Compose Logic ---
    // We execute docker-compose as an external command
    cmdUp := exec.Command("docker", "compose", "up", "-d")
    cmdUp.Dir = challengePath // Run the command in the challenge's directory
    if debug && !silent {
        cmdUp.Stdout = os.Stdout
        cmdUp.Stderr = os.Stderr
    }
    if err := cmdUp.Run(); err != nil {
        log.Printf("Error starting docker-compose for challenge '%s': %v", challenge.Name, err)
        // Attempt to clean up even if startup failed
        cmdDown := exec.Command("docker", "compose", "down")
        cmdDown.Dir = challengePath
        cmdDown.Run()
        return "menu"
    }
    // --- End Docker Compose Logic ---

    if !silent {
        fmt.Printf("\n✅ Challenge '%s' is now running!\n", challenge.Name)
        fmt.Println("   You can interact with it at:")
        for _, port := range challenge.Ports {
            // Adiciona uma descrição simples para as portas conhecidas
            if port == 9001 {
                fmt.Printf("   - Web Console: http://127.0.0.1:%d\n", port)
            } else if port == 9000 {
                fmt.Printf("   - API Endpoint: http://127.0.0.1:%d\n", port)
            } else {
                fmt.Printf("   - http://127.0.0.1:%d\n", port)
            }
        }
        fmt.Println()

        if challenge.Preface != "" {
            fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
            fmt.Println(challenge.Preface)
            fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
        }
    }

    // The user interaction loop
    reader := bufio.NewReader(os.Stdin)
    hintIndex := 0
    var finalResult string

    // Multi-flag support
    foundFlags := make(map[string]bool)
    hasMultipleFlags := len(challenge.Flags) > 0
    totalFlags := len(challenge.Flags)

    interactionLoop:
    for {
        fmt.Print("Enter flag > ")
        input, _ := reader.ReadString('\n')
        input = strings.TrimSpace(input)

        // Check for special commands first
        switch strings.ToLower(input) {
        case "hint":
             if len(challenge.Hints) == 0 {
                fmt.Println("No hints available for this challenge.")
            } else if hintIndex < len(challenge.Hints) {
                fmt.Printf("Hint %d/%d: %s\n", hintIndex+1, len(challenge.Hints), challenge.Hints[hintIndex])
                hintIndex++
            } else {
                fmt.Println("No more hints available.")
            }
            continue
        case "quit", "exit":
            finalResult = "quit"
            break interactionLoop
        }

        // Flag validation
        if hasMultipleFlags {
            // Multi-flag mode
            flagFound := false
            for _, validFlag := range challenge.Flags {
                if strings.EqualFold(input, validFlag) {
                    if foundFlags[validFlag] {
                        fmt.Println("Flag already found")
                    } else {
                        foundFlags[validFlag] = true
                        fmt.Println("\n✅ Correct! Flag found.")

                        // Check if all flags found
                        if len(foundFlags) == totalFlags {
                            if challenge.Postface != "" {
                                fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
                                fmt.Println(challenge.Postface)
                                fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
                            }
                            finalResult = "complete"
                            break interactionLoop
                        }
                    }
                    flagFound = true
                    break
                }
            }
            if !flagFound {
                fmt.Println("Incorrect flag. Try again. (Type 'hint' for a hint, or 'quit' to exit)")
            }
        } else {
            // Single flag mode (backward compatible)
            if strings.EqualFold(input, challenge.Flag) {
                fmt.Println("\n✅ Correct! Well done.")
                if challenge.Postface != "" {
                    fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
                    fmt.Println(challenge.Postface)
                    fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
                }
                fmt.Println("\nType 'next' to go straight to the next challenge, or press Enter to return to the menu...")
                inputNext, _ := reader.ReadString('\n')
                if strings.EqualFold(strings.TrimSpace(inputNext), "next") {
                    finalResult = "next"
                } else {
                    finalResult = "menu"
                }
                break interactionLoop
            } else {
                fmt.Println("Incorrect flag. Try again. (Type 'hint' for a hint, or 'quit' to exit)")
            }
        }
    }

    // Cleanup for Docker Compose
    fmt.Println("\nShutting down the current challenge environment...")
    cmdDown := exec.Command("docker", "compose", "down")
    cmdDown.Dir = challengePath
    if err := cmdDown.Run(); err != nil {
        log.Printf("Warning: could not run 'docker compose down': %v", err)
    }
    
    return finalResult
}

// runDockerfileChallenge handles challenges defined by a Dockerfile.
func runDockerfileChallenge(ctx context.Context, cli *client.Client, dirName string, forceBuild bool, debug bool, silent bool) string {
    // This function contains the exact same logic as your original runChallenge function
    challengePath := filepath.Join("/wss-ctf/challenges", dirName)
    challengeFile, err := os.ReadFile(filepath.Join(challengePath, "challenge.json"))
    if err != nil {
        log.Printf("Error: Could not read challenge.json in %s. Skipping. Details: %v", challengePath, err)
        return "menu"
    }
    var challenge Challenge
    if err := json.Unmarshal(challengeFile, &challenge); err != nil {
        log.Printf("Error: Could not parse challenge.json in %s. Skipping. Details: %v", challengePath, err)
        return "menu"
    }

    if !silent {
        fmt.Printf("\n--- Starting Challenge: %s ---\n", challenge.Name)
    }

    imageTag := "challenge-" + strings.ToLower(dirName) + ":latest"
    containerName := "challenge-container-" + strings.ToLower(dirName)
    cleanup(ctx, cli, containerName, "", false, debug)
    exists, err := imageExists(ctx, cli, imageTag)
    if err != nil {
        log.Printf("Warning: Could not check if image '%s' exists: %v. Attempting to build.", imageTag, err)
    }
    if forceBuild || !exists {
        if forceBuild && debug && !silent {
            fmt.Print("Build forced by user with --build flag")
        }
        err = buildImage(ctx, cli, challengePath, imageTag, debug && !silent)
        if err != nil {
            log.Printf("Error: Failed to build Docker image for challenge %s. Details: %v", dirName, err)
            return "fail"
        }
    } else {
        if debug && !silent {
            fmt.Printf("Using existing image '%s'. Use --build to force a rebuild.\n", imageTag)
        }
    }
    if len(challenge.Ports) == 0 {
    log.Printf("Error: No ports defined in challenge.json for '%s'", challenge.Name)
    return "fail"
    }
    _, err = runContainer(ctx, cli, imageTag, containerName, challenge.Ports[0], debug && !silent)
    if err != nil {
        log.Printf("Error: Failed to run Docker container for challenge %s. Details: %v", dirName, err)
        cleanup(ctx, cli, containerName, imageTag, true, debug)
        return "fail"
    }

    if !silent {
        fmt.Printf("\n✅ Challenge '%s' is now running!\n", challenge.Name)
        fmt.Println("   You can interact with it at:")
        for _, port := range challenge.Ports {
            // Adiciona uma descrição simples para as portas conhecidas
            if port == 9001 {
                fmt.Printf("   - Web Console: http://127.0.0.1:%d\n", port)
            } else if port == 9000 {
                fmt.Printf("   - API Endpoint: http://127.0.0.1:%d\n", port)
            } else {
                fmt.Printf("   - http://127.0.0.1:%d\n", port)
            }
        }
        if challenge.Preface != "" {
            fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
            fmt.Println(challenge.Preface)
            fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
        }
    }
    reader := bufio.NewReader(os.Stdin)
    hintIndex := 0
    var finalResult string

    // Check if this is the first challenge (01-first-chal)
    isFirstChallenge := strings.Contains(dirName, "01-first-chal")

    // Multi-flag support
    foundFlags := make(map[string]bool)
    hasMultipleFlags := len(challenge.Flags) > 0
    totalFlags := len(challenge.Flags)

    interactionLoop:
    for {
        fmt.Print("Enter flag > ")
        input, _ := reader.ReadString('\n')
        input = strings.TrimSpace(input)

        // Check for special commands first
        switch strings.ToLower(input) {
        case "hint":
            if len(challenge.Hints) == 0 {
                fmt.Println("No hints available for this challenge.")
            } else if hintIndex < len(challenge.Hints) {
                fmt.Printf("Hint %d/%d: %s\n", hintIndex+1, len(challenge.Hints), challenge.Hints[hintIndex])
                hintIndex++
            } else {
                fmt.Println("No more hints available.")
            }
            continue
        case "quit", "exit":
            finalResult = "quit"
            break interactionLoop
        }

        // Flag validation
        if hasMultipleFlags {
            // Multi-flag mode
            flagFound := false
            for _, validFlag := range challenge.Flags {
                if strings.EqualFold(input, validFlag) {
                    if foundFlags[validFlag] {
                        fmt.Println("Flag already found")
                    } else {
                        foundFlags[validFlag] = true
                        fmt.Println("\n✅ Correct! Flag found.")

                        // Check if all flags found
                        if len(foundFlags) == totalFlags {
                            if challenge.Postface != "" {
                                fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
                                fmt.Println(challenge.Postface)
                                fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
                            }
                            finalResult = "complete"
                            break interactionLoop
                        }
                    }
                    flagFound = true
                    break
                }
            }
            if !flagFound {
                fmt.Println("Incorrect flag. Try again. (Type 'hint' for a hint, or 'quit' to exit)")
            }
        } else {
            // Single flag mode (backward compatible)
            if strings.EqualFold(input, challenge.Flag) {
                fmt.Println("\n✅ Correct! Well done.")
                if challenge.Postface != "" {
                    fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
                    fmt.Println(challenge.Postface)
                    fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
                }

                if isFirstChallenge {
                    // For first challenge, don't break - return "continue" to keep it running
                    finalResult = "continue"
                    break interactionLoop
                } else {
                    // For other challenges, ask if they want to continue
                    fmt.Println("\nType 'next' to go straight to the next challenge, or press Enter to return to the menu...")
                    inputNext, _ := reader.ReadString('\n')
                    if strings.EqualFold(strings.TrimSpace(inputNext), "next") {
                        finalResult = "next"
                    } else {
                        finalResult = "menu"
                    }
                    break interactionLoop
                }
            } else {
                fmt.Println("Incorrect flag. Try again. (Type 'hint' for a hint, or 'quit' to exit)")
            }
        }
    }

    // Don't cleanup first challenge if returning "continue"
    if !(isFirstChallenge && finalResult == "continue") {
        fmt.Println("\nShutting down the current challenge...")
        cleanup(ctx, cli, containerName, imageTag, false, debug)
    }
    return finalResult
}

// Checks if a Docker image with the given tag exists locally.
func imageExists(ctx context.Context, cli *client.Client, imageTag string) (bool, error) {
	// Use a filter to ask the daemon directly, which is more efficient
	// than listing all images and searching locally.
	filterArgs := filters.NewArgs()
	filterArgs.Add("reference", imageTag)

	images, err := cli.ImageList(ctx, image.ListOptions{Filters: filterArgs})
	if err != nil {
		return false, err
	}

	return len(images) > 0, nil
}

// buildImage creates a Docker image from a Dockerfile in the given path.
func buildImage(ctx context.Context, cli *client.Client, buildContextPath, tag string, debug bool) error {
	if debug {
		fmt.Printf("Building image '%s'...\n", tag)
	}

	// Create a tar archive of the build context directory.
	// The Docker daemon requires the build context as a tar stream.
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	err := filepath.Walk(buildContextPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}
		// The header name needs to be a relative path within the tar archive.
		header.Name, _ = filepath.Rel(buildContextPath, path)
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(tw, file)
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to create tar archive: %w", err)
	}

	buildOptions := types.ImageBuildOptions{
		Tags:       []string{tag},
		Remove:     true, // Remove intermediate containers after a successful build
		Dockerfile: "Dockerfile",
	}

	// Call the Docker SDK to build the image.
	buildResponse, err := cli.ImageBuild(ctx, buf, buildOptions)
	if err != nil {
		return fmt.Errorf("image build request failed: %w", err)
	}
	defer buildResponse.Body.Close()

	// Stream the build output to the console.
	_, err = io.Copy(os.Stdout, buildResponse.Body)
	if err != nil {
		return fmt.Errorf("failed to read build response: %w", err)
	}

	if debug {
		fmt.Printf("Image '%s' built successfully.\n", tag)
	}
	return nil
}

// runContainer creates and starts a container from a given image.
func runContainer(ctx context.Context, cli *client.Client, image, name string, hostPort int, debug bool) (string, error) {
	if debug {
		fmt.Printf("Starting container '%s' from image '%s'...\n", name, image)
	}

	// Configure port mapping.
	portStr := strconv.Itoa(hostPort)
	exposedPorts, portBindings, err := nat.ParsePortSpecs([]string{fmt.Sprintf("%s:80", portStr)})
	if err != nil {
		return "", fmt.Errorf("failed to parse port specs: %w", err)
	}

	// Create the container.
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        image,
		ExposedPorts: exposedPorts,
	}, &container.HostConfig{
		PortBindings: portBindings,
	}, nil, nil, name)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	// Start the container.
	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	if debug {
		fmt.Printf("Container started successfully (ID: %s).\n", resp.ID[:12])
	}
	return resp.ID, nil
}

// cleanup stops and removes a container, and optionally the associated image.
func cleanup(ctx context.Context, cli *client.Client, containerName, imageName string, removeImage bool, debug bool) {
	if debug {
		fmt.Printf("Cleaning up resources for %s...\n", containerName)
	}

	// Stop the container with a timeout.
	timeout := 3 // seconds
	if err := cli.ContainerStop(ctx, containerName, container.StopOptions{Timeout: &timeout}); err != nil {
		// Log the error but don't stop the cleanup process. It might already be stopped.
		// fmt.Printf("Warning: Could not stop container %s (it may not be running): %v\n", containerName, err)
	}

	// Remove the container.
	if err := cli.ContainerRemove(ctx, containerName, container.RemoveOptions{Force: true}); err != nil {
		// fmt.Printf("Warning: Could not remove container %s: %v\n", containerName, err)
	}

	// Optionally remove the image.
	if removeImage && imageName != "" {
		if _, err := cli.ImageRemove(ctx, imageName, image.RemoveOptions{Force: true}); err != nil {
			// fmt.Printf("Warning: Could not remove image %s: %v\n", imageName, err)
		}
	}
	if debug {
		fmt.Println("Cleanup complete.")
	}
}

// cleanAll removes all challenge containers and images
func cleanAll(ctx context.Context, cli *client.Client) {
	// Load the configuration to get all challenge directories
	configFile, err := os.ReadFile("/wss-ctf/challenges/config.json")
	if err != nil {
		log.Printf("Warning: Could not read /wss-ctf/challenges/config.json: %v", err)
		return
	}

	var config Config
	if err := json.Unmarshal(configFile, &config); err != nil {
		log.Printf("Warning: Could not parse /wss-ctf/challenges/config.json: %v", err)
		return
	}

	// Remove all challenge containers and images
	for _, challengeDir := range config.Challenges {
		imageTag := "challenge-" + strings.ToLower(challengeDir) + ":latest"
		containerName := "challenge-container-" + strings.ToLower(challengeDir)

		// Stop and remove container if it exists
		timeout := 3
		if err := cli.ContainerStop(ctx, containerName, container.StopOptions{Timeout: &timeout}); err != nil {
			// Container might not exist, continue
		}
		if err := cli.ContainerRemove(ctx, containerName, container.RemoveOptions{Force: true}); err != nil {
			// Container might not exist, continue
		}

		// Remove image
		if _, err := cli.ImageRemove(ctx, imageTag, image.RemoveOptions{Force: true}); err != nil {
			log.Printf("Note: Could not remove image %s (might not exist)\n", imageTag)
		} else {
			fmt.Printf("Removed image: %s\n", imageTag)
		}
	}
}
