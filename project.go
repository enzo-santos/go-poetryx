package poetryx

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
	"github.com/xaverkapeller/go-gitignore"
)

type PyprojectTomlBuildSystemTable struct {
	Requires     []string `toml:"requires"`
	BuildBackend string   `toml:"build-backend"`
}

type PyprojectTomlToolPoetryTable struct {
	Name         string            `toml:"name"`
	Version      string            `toml:"version"`
	Description  string            `toml:"description"`
	Authors      []string          `toml:"authors"`
	ReadmePath   string            `toml:"readme"`
	Dependencies map[string]string `toml:"dependencies"`
	Scripts      map[string]string `toml:"scripts"`
}

type PyprojectTomlToolTable struct {
	Poetry PyprojectTomlToolPoetryTable `toml:"poetry"`
}

type PyprojectTomlDocument struct {
	Tool        PyprojectTomlToolTable        `toml:"tool"`
	BuildSystem PyprojectTomlBuildSystemTable `toml:"build-system"`
}

const poetryConfigDefaultFileName = "pyproject.toml"
const gitignoreDefaultFileName = ".gitignore"
const initPythonScriptDefaultFileName = "__init__.py"

type PoetryProject struct {
	Name string
	Path string
}

func (p PoetryProject) WritePoetryConfig(document PyprojectTomlDocument) error {
	filePath := filepath.Join(p.Path, poetryConfigDefaultFileName)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error while creating file: %w", err)
	}
	defer file.Close()
	if err := toml.NewEncoder(file).Encode(document); err != nil {
		return fmt.Errorf("error while writing to file: %w", err)
	}
	return nil
}

func (p PoetryProject) ReadPoetryConfig() (PyprojectTomlDocument, error) {
	filePath := filepath.Join(p.Path, poetryConfigDefaultFileName)
	file, err := os.Open(filePath)
	if err != nil {
		return PyprojectTomlDocument{}, fmt.Errorf("error while opening file: %w", err)
	}
	defer file.Close()
	var document PyprojectTomlDocument
	if err := toml.NewDecoder(file).Decode(&document); err != nil {
		return PyprojectTomlDocument{}, fmt.Errorf("error while decoding file: %w", err)
	}
	return document, nil
}

func (p PoetryProject) InitializeDefaultMainPythonFile() error {
	filePath := filepath.Join(p.Path, p.Name, initPythonScriptDefaultFileName)

	inputFile, err := os.Open(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Do nothing if the file does not exist, we will create it later
		} else {
			return fmt.Errorf("error while opening file: %w", err)
		}
	} else if err == nil {
		defer inputFile.Close()
		reader := bufio.NewReader(inputFile)
		if _, err := reader.ReadByte(); err == nil {
			// Do nothing if the file already has contents
			return nil
		} else if !errors.Is(err, io.EOF) {
			return fmt.Errorf("error while reading first byte from file: %w", err)
		}
		if err := inputFile.Close(); err != nil {
			return fmt.Errorf("error while closing opened file: %w", err)
		}
	}

	outputFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error while creating file: %w", err)
	}
	defer outputFile.Close()
	writer := bufio.NewWriter(outputFile)
	defer writer.Flush()
	if _, err := writer.WriteString(
		"" +
			"def main() -> None:\n" +
			"    pass\n" +
			"\n" +
			"if __name__ == \"__main__\":\n" +
			"    main()\n",
	); err != nil {
		return fmt.Errorf("error while writing to file: %w", err)
	}
	return nil
}

func (p PoetryProject) AddScript(name string, location string) error {
	document, err := p.ReadPoetryConfig()
	if err != nil {
		return fmt.Errorf("error while reading Poetry config file: %w", err)
	}
	currentLocation, ok := document.Tool.Poetry.Scripts[name]
	if ok && location == currentLocation {
		return nil
	}
	scripts := make(map[string]string, len(document.Tool.Poetry.Scripts))
	for key, value := range document.Tool.Poetry.Scripts {
		scripts[key] = value
	}
	scripts[name] = location
	return p.WritePoetryConfig(PyprojectTomlDocument{
		Tool: PyprojectTomlToolTable{
			Poetry: PyprojectTomlToolPoetryTable{
				Name:         document.Tool.Poetry.Name,
				Version:      document.Tool.Poetry.Version,
				Description:  document.Tool.Poetry.Description,
				Authors:      document.Tool.Poetry.Authors,
				ReadmePath:   document.Tool.Poetry.ReadmePath,
				Dependencies: document.Tool.Poetry.Dependencies,
				Scripts:      scripts,
			},
		},
		BuildSystem: document.BuildSystem,
	})
}

func (p PoetryProject) AddDirectory(name string) error {
	directoryPath := filepath.Join(p.Path, name)
	if err := os.MkdirAll(directoryPath, 0777); err != nil {
		return fmt.Errorf("error while creating directory: %w", err)
	}
	return nil
}

func (p PoetryProject) AddIgnoredPath(path string) error {
	filePath := filepath.Join(p.Path, gitignoreDefaultFileName)
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("error while creating file: %w", err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("error while closing file: %w", err)
		}
	}
	contents, err := gitignore.NewFromFile(filePath)
	if err != nil {
		return fmt.Errorf("error while parsing file: %w", err)
	}
	match := contents.Match(filePath)
	if match != nil && match.Ignore() {
		return nil
	}
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return fmt.Errorf("error while opening file for adding path: %w", err)
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	defer writer.Flush()
	if _, err := writer.WriteString(fmt.Sprintf("%s/\n", path)); err != nil {
		return fmt.Errorf("error while adding path to file: %w", err)
	}
	return nil
}
