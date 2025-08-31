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
	Flag     string   `json:"flag"`
	Hints    []string `json:"hints"`
	Port     int      `json:"port"`
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
	fmt.Println("\nInitializing...")

	fmt.Println("Docker client connected successfully.")

	// Load the main configuration file.
	configFile, err := os.ReadFile("challenges/config.json")
	if err != nil {
		log.Fatalf("Error: Could not read challenges/config.json. Details: %v", err)
	}

	var config Config
	if err := json.Unmarshal(configFile, &config); err != nil {
		log.Fatalf("Error: Could not parse challenges/config.json. Details: %v", err)
	}

	// Main menu loop
	reader := bufio.NewReader(os.Stdin)
    currentIndex := -1 // -1means that the user will choose

    for {
        if currentIndex == -1 {
            fmt.Println("\n=== CHALLENGE MENU ===")
            fmt.Println("Pick your challenge:")

            // Display available challenges
            for i, challengeDir := range config.Challenges {
                challengePath := filepath.Join("challenges", challengeDir)
                challengeFile, err := os.ReadFile(filepath.Join(challengePath, "challenge.json"))
                if err != nil {
                    fmt.Printf("%d. %s (Error loading metadata)\n", i+1, challengeDir)
                    continue
                }
                var challenge Challenge
                if err := json.Unmarshal(challengeFile, &challenge); err != nil {
                    fmt.Printf("%d. %s (Error parsing metadata)\n", i+1, challengeDir)
                    continue
                }
                fmt.Printf("%d. %s\n", i+1, challenge.Name)
            }
            fmt.Println("\nType a number to select a challenge, or 'quit' to exit")
            fmt.Print("Your choice > ")

            input, _ := reader.ReadString('\n')
            input = strings.TrimSpace(input)

            if strings.EqualFold(input, "quit") || strings.EqualFold(input, "exit") {
                fmt.Println("\nExiting challenge platform. Goodbye!")
                return
            }

            choice, err := strconv.Atoi(input)
            if err != nil || choice < 1 || choice > len(config.Challenges) {
                fmt.Println("Invalid choice. Please enter a number between 1 and", len(config.Challenges))
                continue
            }
            currentIndex = choice - 1
        }

        challengeDir := config.Challenges[currentIndex]
        result := runChallenge(ctx, cli, challengeDir, *build, *debug)
        if result == "next" {
            // Go to the next challenge automatically
            if currentIndex+1 < len(config.Challenges) {
                fmt.Println("\nMoving on to the next challenge...")
                currentIndex++
                continue
            } else {
                fmt.Println("\nYou have completed all the challenges!")
                currentIndex = -1 // Back to menu
                continue
            }
        }
        // Back to menu
        currentIndex = -1
    }
}

// runChallenge handles the logic for a single challenge: build, run, interact, and cleanup.
func runChallenge(ctx context.Context, cli *client.Client, dirName string, forceBuild bool, debug bool) string {	challengePath := filepath.Join("challenges", dirName)

	// Load challenge metadata.
	challengeFile, err := os.ReadFile(filepath.Join(challengePath, "challenge.json"))
	if err != nil {
		log.Printf("Error: Could not read challenge.json in %s. Skipping. Details: %v", challengePath, err)
		return "menu" // Return true to continue to the next challenge
	}

	var challenge Challenge
	if err := json.Unmarshal(challengeFile, &challenge); err != nil {
		log.Printf("Error: Could not parse challenge.json in %s. Skipping. Details: %v", challengePath, err)
		return "menu"
	}

	fmt.Printf("\n--- Starting Challenge: %s ---\n", challenge.Name)

	// Define unique names for the image and container to avoid conflicts.
	imageTag := "challenge-" + strings.ToLower(dirName) + ":latest"
	containerName := "challenge-container-" + strings.ToLower(dirName)

	// Clean up any previous container with the same name.
	cleanup(ctx, cli, containerName, "", false, debug) // Don't remove image yet, just container

	// Check if the image exists locally.
	exists, err := imageExists(ctx, cli, imageTag)
	if err != nil {
		// If we can't check, it's better to try building
		log.Printf("Warning: Could not check if image '%s' exists: %v. Attempting to build.", imageTag, err)
	}

	if forceBuild || !exists {
		if forceBuild && debug {
			fmt.Print("Build forced by user with --build flag")
		}
		// Build the Docker image for the challenge.
		err = buildImage(ctx, cli, challengePath, imageTag, debug)
		if err != nil {
			log.Printf("Error: Failed to build Docker image for challenge %s. Details: %v", dirName, err)
			return "fail"
		}
	} else {
		if debug {
			fmt.Printf("Using existing image '%s'. Use --build to force a rebuild.\n", imageTag)
		}
	}

	// Run the container from the newly built image.
	_, err = runContainer(ctx, cli, imageTag, containerName, challenge.Port, debug)
	if err != nil {
		log.Printf("Error: Failed to run Docker container for challenge %s. Details: %v", dirName, err)
		cleanup(ctx, cli, containerName, imageTag, true, debug) // Cleanup container and image on failure
		return "fail"
	}

	fmt.Printf("\n✅ Challenge '%s' is now running!\n", challenge.Name)
	fmt.Printf("   You can interact with it at: http://127.0.0.1:%d\n\n", challenge.Port)

	// Display preface if available
	if challenge.Preface != "" {
		fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println(challenge.Preface)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	}

	// Start the user interaction loop.
	reader := bufio.NewReader(os.Stdin)
	hintIndex := 0
	for {
		fmt.Print("Enter flag > ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if strings.EqualFold(input, challenge.Flag) {
			fmt.Println("\n✅ Correct! Well done.")
            fmt.Println("\nShutting down the current challenge...")

            // Display postface if available
            if challenge.Postface != "" {
                fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
                fmt.Println(challenge.Postface)
                fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
            }

            cleanup(ctx, cli, containerName, imageTag, false, debug)
            fmt.Println("\nType 'next' to go straight to the next challenge, or press Enter to return to the menu...")
            inputNext, _ := reader.ReadString('\n')
            inputNext = strings.TrimSpace(inputNext)
            if strings.EqualFold(inputNext, "next") {
                return "next" // Signals the menu to advance to the next challenge
            }
            return "menu" // Nack to menu
		} else if strings.EqualFold(input, "hint") {
			if len(challenge.Hints) == 0 {
				fmt.Println("No hints available for this challenge.")
			} else if hintIndex < len(challenge.Hints) {
				fmt.Printf("Hint %d/%d: %s\n", hintIndex+1, len(challenge.Hints), challenge.Hints[hintIndex])
				hintIndex++
				if hintIndex < len(challenge.Hints) {
					fmt.Printf("(Type 'hint' again for another hint)\n")
				} else {
					fmt.Println("(No more hints available)")
				}
			} else {
				fmt.Println("No more hints available. You've seen all the hints.")
			}
		} else if strings.EqualFold(input, "menu") {
			fmt.Println("\nReturning to challenge menu...")
			cleanup(ctx, cli, containerName, imageTag, false, debug)
			return "menu" // Return to menu
		} else {
			fmt.Println("Incorrect flag. Try again. (Type 'hint' for a hint or 'menu' to return to menu)")
		}
	}
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
	configFile, err := os.ReadFile("challenges/config.json")
	if err != nil {
		log.Printf("Warning: Could not read challenges/config.json: %v", err)
		return
	}

	var config Config
	if err := json.Unmarshal(configFile, &config); err != nil {
		log.Printf("Warning: Could not parse challenges/config.json: %v", err)
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
