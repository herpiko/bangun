package build

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	network "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/go-git/go-git/v5"
	"github.com/google/uuid"
	imagespec "github.com/opencontainers/image-spec/specs-go/v1"
)

type DebianBuildRequest struct {
	Distro string
	Arch   string
	GitURL string
}

type DebianBuildResult struct {
	Log string
}

type DebianBuildUsecase interface {
	Build(DebianBuildRequest) (DebianBuildResult, error)
}

type debianBuildUsecase struct {
}

func NewDebianBuildUsecase() DebianBuildUsecase {
	return &debianBuildUsecase{}
}

func (b *debianBuildUsecase) Build(req DebianBuildRequest) (DebianBuildResult, error) {

	id := uuid.New().String()
	jobID := strings.Replace(id, "-", "", -1)

	_, err := git.PlainClone(fmt.Sprintf("/tmp/bangun/builds/%s", jobID), false, &git.CloneOptions{
		URL:      req.GitURL,
		Progress: os.Stdout,
	})

	if err != nil {
		fmt.Println(err)
		return DebianBuildResult{}, err
	}

	dockerClient, err := client.NewEnvClient()
	if err != nil {
		log.Fatalf("Unable to create docker client: %s", err)
	}

	imagename := "herpiko/pbocker-" + req.Distro
	containername := fmt.Sprintf("bangun-build-%s", jobID)
	portopening := "8080"
	inputEnv := []string{fmt.Sprintf("LISTENINGPORT=%s", portopening)}
	resultPath := fmt.Sprintf("/media/homelab/src/bangun/results/%s/%s", jobID, req.Distro)
	err = runContainer(dockerClient, jobID, imagename, containername, portopening, inputEnv, resultPath)
	if err != nil {
		log.Println(err)
	}

	return DebianBuildResult{}, nil
}

func runContainer(
	client *client.Client,
	jobID string,
	imagename string,
	containername string,
	port string,
	inputEnv []string,
	resultPath string,
) error {
	// Configured hostConfig:
	// https://godoc.org/github.com/docker/docker/api/types/container#HostConfig
	hostConfig := &container.HostConfig{
		Privileged: true,
		RestartPolicy: container.RestartPolicy{
			Name: "no",
		},
		LogConfig: container.LogConfig{
			Type:   "json-file",
			Config: map[string]string{},
		},
	}

	// Define Network config (why isn't PORT in here...?:
	// https://godoc.org/github.com/docker/docker/api/types/network#NetworkingConfig
	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{},
	}
	gatewayConfig := &network.EndpointSettings{
		Gateway: "172.17.0.1",
	}
	networkConfig.EndpointsConfig["bridge"] = gatewayConfig

	// Configuration
	// https://godoc.org/github.com/docker/docker/api/types/container#Config
	config := &container.Config{
		Image:    imagename,
		Env:      inputEnv,
		Hostname: fmt.Sprintf("%s-hostnameexample", imagename),
	}

	// Mounting a volume:
	// https://docs.docker.com/engine/tutorials/dockervolumes/
	// https://godoc.org/github.com/docker/docker/api/types/container#Mount
	repoMount := mount.Mount{
		Type:     mount.TypeBind,
		Source:   "/tmp/bangun/builds/" + jobID,
		Target:   "/src",
		ReadOnly: false,
	}
	scriptMount := mount.Mount{
		Type:     mount.TypeBind,
		Source:   "/media/homelab/src/bangun/scripts",
		Target:   "/scripts",
		ReadOnly: false,
	}

	// create the directory if not exist
	err := os.MkdirAll(resultPath, os.ModePerm)
	if err != nil {
		log.Println(err)
		return err
	}
	resultMount := mount.Mount{
		Type:     mount.TypeBind,
		Source:   resultPath,
		Target:   "/result",
		ReadOnly: false,
	}
	hostConfig.Mounts = []mount.Mount{repoMount, scriptMount, resultMount}

	// Run script after container start
	// https://docs.docker.com/engine/reference/builder/#cmd
	config.Cmd = []string{"/bin/bash", "-c", "cd /src && /scripts/debian-build.sh"}

	// Creating the actual container. This is "nil,nil,nil" in every example.
	cont, err := client.ContainerCreate(
		context.Background(),
		config,
		hostConfig,
		networkConfig,
		&imagespec.Platform{},
		containername,
	)

	if err != nil {
		log.Println(err)
		return err
	}

	// Run the actual container
	client.ContainerStart(context.Background(), cont.ID, container.StartOptions{})
	log.Printf("Container %s is created", cont.ID)

	// get the log of the container
	logRC, err := client.ContainerLogs(context.Background(), cont.ID, container.LogsOptions{
		ShowStdout: true,
		Follow:     true,
	})
	if err != nil {
		log.Println(err)
		return err
	}

	// logRC is a ioReadCloser. Stream and print it to stdout
	// _, err = io.Copy(os.Stdout, logRC)
	// if err != nil {
	//	log.Println(err)
	//		return err
	//	}

	// Also stream it to a file
	logFile, err := os.Create(fmt.Sprintf(resultPath + "/log.txt"))
	if err != nil {
		log.Println(err)
		return err
	}
	defer logFile.Close()
	_, err = io.Copy(logFile, logRC)
	if err != nil {
		log.Println(err)
		return err
	}

	// Wait for container to be done and print the log
	resultC, errC := client.ContainerWait(context.Background(), containername, "")
	select {
	case err := <-errC:
		log.Println(err)
	case res := <-resultC:
		log.Println(res)
	}

	return nil
}
