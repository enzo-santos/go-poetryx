package poetryx

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var ErrExecutableNotFound = errors.New("could not find Poetry executable")

type Driver struct {
	PoetryExecutableFilePath string
}

func NewDriverFromEnvironment() (Driver, error) {
	var cmdExecutableText string
	switch osName := runtime.GOOS; osName {
	case "linux":
	case "ios":
		cmdExecutableText = "which"
	case "windows":
		cmdExecutableText = "where"
	default:
		return Driver{}, fmt.Errorf("unsupported OS: %q", osName)
	}
	cmd := exec.Command(cmdExecutableText, "poetry")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return Driver{}, fmt.Errorf("error while retrieving stdout from command: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return Driver{}, fmt.Errorf("error while launching command: %w", err)
	}
	reader := bufio.NewReader(stdout)
	line, err := reader.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) {
			return Driver{}, ErrExecutableNotFound
		}
		return Driver{}, fmt.Errorf("error while reading line from stdout: %w", err)
	}
	path := strings.TrimSpace(line)
	return Driver{
		PoetryExecutableFilePath: path,
	}, nil
}

func NewDriver(path string) (Driver, error) {
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Driver{}, ErrExecutableNotFound
		}
		return Driver{}, fmt.Errorf("error while checking if file exists: %w", err)
	}
	return Driver{
		PoetryExecutableFilePath: path,
	}, nil
}

func (d Driver) CreateNewProject(path string, name string) (PoetryProject, error) {
	projectPath := filepath.Join(path, name)
	if _, err := os.Stat(projectPath); err == nil {
		return PoetryProject{}, fmt.Errorf("directory already exists: %s", projectPath)
	}
	cmd := exec.Command(
		d.PoetryExecutableFilePath,
		"new",
		name,
		fmt.Sprintf("--directory=%q", path),
	)
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		return PoetryProject{}, fmt.Errorf("error while running `poetry new`: %w", err)
	}
	return PoetryProject{
		Name: name,
		Path: projectPath,
	}, nil
}

func (d Driver) Install(path string) error {
	if err := exec.Command(
		d.PoetryExecutableFilePath,
		"install",
		fmt.Sprintf("--directory=%s", path),
	).Run(); err != nil {
		return fmt.Errorf("error while installing Poetry project %q: %w", path, err)
	}
	return nil
}

func (d Driver) InstallProject(project PoetryProject) error {
	return d.Install(project.Path)
}
