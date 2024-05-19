# go-poetryx

A extension to `poetry` implemented in Go.

## Usage

### `init` command

Same as `poetry new <project-name>`, but

- adds _assets/_ and _build/_ folders, with their entries to _.gitignore_
- overwrites the default _\_\_init\_\_.py_ to include a `main` function
- adds a script linked to that `main` function, allowing to run `poetry run main` without further modifications
- runs `poetry install` to make Poetry recognize that script

#### Syntax

```shell
poetryx init --name <project-name> [--poetry-path <poetry-path>] [--directory <directory-path>] 
```

where

- `<project-name>` is the name of the project, as you would pass to `poetry new <project-name>`.
- `<poetry-path>` is the path of the Poetry executable. If not provided, it'll infer it from the PATH environment
  variable.
- `<directory-path>` is the path of the directory to create the project, as you would pass
  to `poetry new ... --directory <directory-path>`. If not provided, it defaults to the current working directory.
