# go-mage

go-mage is a project that uses magefile instead of a Makefile to build and manage Go projects. magefile is a build tool written in Go, working similarly to Makefile but leveraging the power and flexibility of the Go language.

## Purpose of the Project

The purpose of the go-mage project is to provide a cross-platform example of a magefile for building, testing, and managing other build operations for Go projects. Since magefile is written in Go, it offers a more familiar and flexible structure for Go developers.

## Features and Notes

### Cross-Platform Support
This project can be built on Windows, macOS, and Linux operating systems.

- **Windows-Specific Features:**
  - Allows building for Windows using MSYS2 MinGW. So you should install MSYS2 first or just delete those lines from Magefile.
  - Automatically generates .syso and version info files.
  - Before building on Windows, ensure you place the manifest (app.manifest) and icon (app.icon) files into the build/windows/assets directory.

- **macOS-Specific Features:**:
  - Allows building with a specific macOS SDK. Simply set the desired SDK version before running Mage by using an environment variable, for example: `export MACOS_SDK_VERSION=14.5`

### Security and Code Quality
  - Performs a security scan using vulnerability check.
  - Ensures code quality with golangci-lint.
  - Provides test coverage reports using gocov.
  - During the build process, the Version, BuildTime, CommitID, and BuildMode variables in the project are updated using the version and other information specified in the magefile.
  - The project utilizes the tools goversioninfo, golangci-lint, gocov, and govulncheck. If these tools are not already installed on the system, they are automatically installed.

## Prerequisites

The mage tool must be installed on your system for this project to work. You can install mage by running the following command:
```sh
go install github.com/magefile/mage@latest
```

After installation, you can execute the magefile using the mage command. For example:
```sh
mage build
```

Alternatively, you can run the magefile directly using Go without installing the mage tool:
```sh
go run mage.go build
```

As another option, you can precompile mage and run it on your server, eliminating dependency issues:
```sh
mage --compile ./runmage
./runmage
```

## Installation

After cloning the project, follow these steps to install the required dependencies:
```sh
git clone https://github.com/yakuter/go-mage.git
cd go-mage
go mod tidy
```

## Usage
To see the available mage commands in the project, simply run:
```sh
➜ mage
Targets:
  build        builds the binary.
  clean        cleans the build directory.
  generate     runs go generate.
  linter       runs the linter.
  test         runs the tests.
  vulncheck    runs the vulnerability check.
```

You can then select and run any command directly. For example, to detect vulnerabilities using Go's built-in tool, run:
```sh
mage vulncheck
```

## Contributing
If you’d like to contribute, please submit a pull request or open an issue.

## License
This project is licensed under the MIT License. See the LICENSE file for more details.