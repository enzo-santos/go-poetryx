package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/akamensky/argparse"

	"github.com/enzo-santos/go-poetryx"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	parser := argparse.NewParser("init", "Initializes a Poetry project with additional configurations.")
	poetryExecutablePathReference := parser.String("", "poetry-path", &argparse.Options{
		Required: false,
		Help: "path to the Poetry executable. " +
			"If not provided, it tries to find it from PATH.",
	})
	projectNameReference := parser.String("", "name", &argparse.Options{
		Required: true,
		Help:     "name of the Poetry project to be created. Should not contain spaces.",
	})
	projectRootPathReference := parser.String("d", "directory", &argparse.Options{
		Required: false,
		Help: "path to the containing folder of the project to be created. " +
			"If not provided, it uses the current working directory.",
	})
	if err := parser.Parse(os.Args); err != nil {
		return fmt.Errorf("error while parsing arguments: %w", err)
	}
	var driver poetryx.Driver
	{
		poetryExecutablePath := *poetryExecutablePathReference
		if len(poetryExecutablePath) == 0 {
			d, err := poetryx.NewDriverFromEnvironment()
			if err != nil {
				if errors.Is(err, poetryx.ErrExecutableNotFound) {
					fmt.Println("Could not find Poetry from PATH. Try setting the `--poetry-path` argument.")
					fmt.Println("It should point to the Poetry executable, not its containing folder.")
					fmt.Println("Example: `poetryx init foo --poetry-path C:\\Users\\John\\Poetry\\bin\\poetry.exe`")
					return nil
				}
				return fmt.Errorf("error while detecting Poetry driver from environment: %w", err)
			}
			driver = d
		} else {
			d, err := poetryx.NewDriver(poetryExecutablePath)
			if err != nil {
				if errors.Is(err, poetryx.ErrExecutableNotFound) {
					fmt.Printf("Could not find Poetry on %q. Try checking it again.\n", poetryExecutablePath)
					fmt.Println("It should point to the Poetry executable, not its containing folder.")
					fmt.Println("Example: `poetryx init foo --poetry-path C:\\Users\\John\\Poetry\\bin\\poetry.exe`")
					return nil
				}
				return fmt.Errorf("error while detecting Poetry driver on %q: %w", poetryExecutablePath, err)
			}
			driver = d
		}
	}

	var projectRootPath string
	{
		path := *projectRootPathReference
		if len(path) == 0 {
			currentWorkingDirectoryPath, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("error while retrieving current working directory: %w", err)
			}
			projectRootPath = currentWorkingDirectoryPath
		} else {
			projectRootPath = path
		}
	}
	projectName := *projectNameReference
	project, err := driver.CreateNewProject(projectRootPath, projectName)
	if err != nil {
		return fmt.Errorf("error while creating project: %w", err)
	}
	log.Print("`poetry new` ran successfully")
	for _, directoryName := range []string{"assets", "build"} {
		if err := project.AddDirectory(directoryName); err != nil {
			return fmt.Errorf("error while adding directory %q to project: %w", directoryName, err)
		}
		log.Printf("created directory %s", directoryName)
		if err := project.AddIgnoredPath(directoryName); err != nil {
			return fmt.Errorf("error while adding directory %q to project's .gitignore: %w", directoryName, err)
		}
		log.Printf("added directory %s to .gitignore", directoryName)
	}
	if err := project.InitializeDefaultMainPythonFile(); err != nil {
		return fmt.Errorf("error while initializing `__init__.py`: %w", err)
	}
	log.Printf("initialized %s/__init__.py", projectName)
	if err := project.AddScript("main", fmt.Sprintf("%s:main", projectName)); err != nil {
		return fmt.Errorf("error while adding script: %w", err)
	}
	log.Print("main script added")
	if err := driver.InstallProject(project); err != nil {
		return fmt.Errorf("error while installing project: %w", err)
	}
	log.Print("`poetry install` ran successfully")
	return nil
}
