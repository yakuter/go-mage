//go:build mage
// +build mage

package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var (
	name               = "go-mage"
	version            = "3.16.0"
	linterVersion      = "v1.52.2"
	versionInfoVersion = "v1.4.0"
	gocovVersion       = "v1.1.0"
	mingw_xcompiler    = "x86_64-w64-mingw32-gcc"
	mingw_xcompiler_32 = "i686-w64-mingw32-gcc"
	targetDir          = "./build"
	mainDir            = "./cmd/go-mage"
	assets             = map[string]string{
		"manifest":    "build/windows/assets/app.manifest",
		"icon'":       "build/windows/assets/app.ico",
		"versioninfo": "cmd/go-mage/versioninfo.json",
	}
	cleanupFiles = []string{
		"*cover.out",
		"*.log",
		"./cmd/go-mage/versioninfo.json",
		"./cmd/go-mage/*.syso",
		"./build/go-mage_*",
	}
)

type Tool struct {
	Name    string
	Address string
	Version string
	IsRepo  bool
}

var toolInfo = []Tool{
	{Name: "goversioninfo", Address: "github.com/josephspurrier/goversioninfo/cmd/goversioninfo", Version: versionInfoVersion, IsRepo: true},
	{Name: "golangci-lint", Address: "https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh", Version: linterVersion, IsRepo: false},
	{Name: "gocov", Address: "github.com/axw/gocov/gocov", Version: gocovVersion, IsRepo: true},
	{Name: "govulncheck", Address: "golang.org/x/vuln/cmd/govulncheck", Version: "latest", IsRepo: true},
}

type Builder struct {
	goos          string
	arch          string
	extra_tags    string
	extra_flags   []string
	extra_ldflags string
	tools         map[string]string
}

// Generate runs go generate.
func (b Builder) Generate() error {
	fmt.Println("Running Generate")
	fmt.Printf("Environment: %q \n", b.env())
	if err := b.installTools(); err != nil {
		return fmt.Errorf("failed to install tools error: %v", err)
	}
	if err := sh.RunWithV(b.env(), mg.GoCmd(), "generate", "./..."); err != nil {
		return fmt.Errorf("failed to run go generate error: %v", err)
	}
	if err := b.ensureSyso(); err != nil {
		return fmt.Errorf("failed to ensure syso error: %v", err)
	}
	fmt.Println("Generate completed successfully")
	return nil
}

// Linter runs the linter.
func (b Builder) Linter() error {
	fmt.Println("Running linter")
	fmt.Printf("Environment: %q \n", b.env())
	if err := b.installTools(); err != nil {
		return fmt.Errorf("failed to install tools error: %v", err)
	}
	if err := sh.RunWithV(b.env(), b.tools["golangci-lint"], "run", "--timeout", "15m", "./..."); err != nil {
		return fmt.Errorf("failed to run linter error: %v", err)
	}
	fmt.Println("Linter completed successfully")
	return nil
}

// VulnCheck runs the vulnerability check.
func (b Builder) Vulncheck() error {
	fmt.Println("Running vulnerability check")
	fmt.Printf("Environment: %q \n", b.env())
	if err := b.installTools(); err != nil {
		return fmt.Errorf("failed to install tools error: %v", err)
	}
	if err := sh.RunWithV(b.env(), b.tools["govulncheck"], "./..."); err != nil {
		return fmt.Errorf("failed to run linter error: %v", err)
	}
	fmt.Println("Vulnerability check completed successfully")
	return nil
}

// Test runs the tests.
func (b Builder) Test() error {
	fmt.Println("Running tests")
	fmt.Printf("Environment: %q \n", b.env())
	if err := b.installTools(); err != nil {
		return fmt.Errorf("failed to install tools error: %v", err)
	}

	if err := b.Generate(); err != nil {
		return fmt.Errorf("failed to generate error: %v", err)
	}

	coverFile := fmt.Sprintf("%s-cover.out", b.goos)

	args := []string{
		"test",
		"-v",
		"-count=1",
		"-failfast",
		"-tags", b.extra_ldflags + b.tags(),
		"-coverpkg=./...",
		"-coverprofile=" + coverFile,
	}
	if goarch() != "386" && os.Getenv("ENABLE_TEST_RACE") == "1" {
		args = append(args, "-race")
	}
	args = append(args, b.extra_flags...)
	args = append(args, "./...")

	fmt.Printf("Testing with envs: %q \n", b.env())
	fmt.Printf("Testing with args: %q \n", args)
	if err := sh.RunWithV(b.env(), mg.GoCmd(), args...); err != nil {
		return fmt.Errorf("failed to run tests error: %v", err)
	}

	if err := b.convertCoverage(coverFile); err != nil {
		return err
	}

	fmt.Println("Tests completed successfully")
	return nil
}

func (b Builder) convertCoverage(coverFile string) error {
	out, err := sh.OutputWith(b.env(), b.tools["gocov"], "convert", coverFile)
	if err != nil {
		return fmt.Errorf("failed to convert coverage: %w", err)
	}

	cmd := exec.Command(b.tools["gocov"], "report")
	cmd.Stdin = strings.NewReader(out)
	cmdOutput, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to run report: %w", err)
	}
	fmt.Printf("%s\n", cmdOutput)
	return nil
}

// Build builds the binary for the given platform.
func (b Builder) Build() error {
	fmt.Println("Building binary")
	fmt.Printf("Environment: %q \n", b.env())

	printGoenv()

	if err := b.installTools(); err != nil {
		return fmt.Errorf("failed to install tools error: %v", err)
	}

	fmt.Printf("Creating output directory '%s'\n", targetDir)
	if err := os.MkdirAll(targetDir, 0700); err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to create output: %v", err)
	}

	targetFile := filepath.Join(targetDir, b.targetFilename())
	fmt.Printf("Removing existing binary '%s'\n", targetFile)
	if err := sh.Rm(targetFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing binary: %v", err)
	}

	if err := b.ensureSyso(); err != nil {
		return fmt.Errorf("failed to ensure syso: %v", err)
	}

	args := []string{
		"build",
		"-trimpath",
		"-tags", b.tags(),
		"-ldflags=" + b.extra_ldflags + b.ldflags(),
		"-o", targetFile,
	}
	args = append(args, b.extra_flags...)
	args = append(args, mainDir)

	fmt.Printf("Building with args: %q \n", args)
	if err := sh.RunWithV(b.env(), mg.GoCmd(), args...); err != nil {
		return fmt.Errorf("failed to build binary error: %v", err)
	}

	fmt.Println("Binary build completed successfully")
	return nil
}

// Clean cleans the files generated by the build.
func (b Builder) Clean() error {
	fmt.Println("Cleaning up")
	for _, pattern := range cleanupFiles {
		files, err := filepath.Glob(pattern)
		if err != nil {
			return fmt.Errorf("failed to glob pattern '%s' error: %v\n", pattern, err)
		}
		for _, file := range files {
			fmt.Printf("Removing file '%s'\n", file)
			if err := os.Remove(file); err != nil {
				return fmt.Errorf("failed to remove file '%s' error: %v\n", file, err)
			}
		}
	}
	fmt.Println("Cleaning up completed successfully")
	return nil
}

func (b *Builder) installTools() error {
	if b.tools == nil {
		b.tools = make(map[string]string)
	}
	for _, tool := range toolInfo {
		if err := b.installTool(tool); err != nil {
			// Return at first error
			return err
		}
	}
	return nil
}

func (b *Builder) installTool(tool Tool) error {
	if b.tools[tool.Name] != "" {
		return nil
	}

	gopath, err := gopath()
	if err != nil {
		return err
	}
	gobin, err := gobin(gopath)
	if err != nil {
		return err
	}

	var filename = tool.Name
	var filenameWithVersion = tool.Name + "-" + tool.Version
	if b.goos == "windows" {
		filename += ".exe"
		filenameWithVersion += ".exe"
	}

	path, err := b.locateBinPath(gobin, gopath, filenameWithVersion)
	if err == nil {
		fmt.Printf("Tool '%s' found at path '%s'\n", tool.Name, path)
		b.tools[tool.Name] = path
		return nil
	}

	fmt.Printf("Tool '%s' not found, installing\n", tool.Name)

	if tool.IsRepo {
		if err := sh.RunWithV(b.env(), mg.GoCmd(), "install", tool.Address+"@"+tool.Version); err != nil {
			return fmt.Errorf("failed to install tool '%s' error: %v\n", tool.Name, err)
		}
	} else {
		cmd := exec.Command("sh", "-s", "--", "-b", gobin, tool.Version)
		resp, err := http.Get(tool.Address)
		if err != nil {
			return fmt.Errorf("failed to install tool '%s' error: %v\n", tool.Name, err)
		}
		cmd.Stdin = resp.Body
		if err := cmd.Run(); err != nil {
			_ = resp.Body.Close()
			return fmt.Errorf("failed to install tool '%s' error: %v\n", tool.Name, err)
		}
		_ = resp.Body.Close()
	}

	path, err = b.locateBinPath(gobin, gopath, filename)
	if err != nil {
		return fmt.Errorf("failed to install tool '%s' error: %v\n", tool.Name, err)
	}

	pathWithVersion := filepath.Join(filepath.Dir(path), filenameWithVersion)
	if err := os.Rename(path, pathWithVersion); err != nil {
		return fmt.Errorf("failed to rename '%s' as '%s' error: %v\n", path, pathWithVersion, err)
	}

	b.tools[tool.Name] = pathWithVersion

	fmt.Printf("Tool '%s' installed successfully at path '%s'\n", tool.Name, pathWithVersion)
	return nil
}

func (b *Builder) locateBinPath(gobin, gopath, filename string) (string, error) {
	patterns := []string{
		// /C/Users/username/go/bin/filename-version
		filepath.Join(gopath, "bin", filename),
		// /C/Users/username/go/bin/windows_amd64/filename-version
		filepath.Join(gopath, "bin", b.goos+"_"+b.arch, filename),
		// /C/Users/username/go/bin/filename
		filepath.Join(gobin, filename),
		// /C/Users/username/go/bin/windows_amd64/filename
		filepath.Join(gobin, b.goos+"_"+b.arch, filename),
	}
	var path string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return "", err
		}
		// break on first match
		if len(matches) > 0 {
			path = matches[0]
			break
		}
	}
	if path == "" {
		return "", fmt.Errorf("failed to locate '%s'", filename)
	}
	return path, nil
}

func (b *Builder) targetFilename() string {
	var extension string
	if b.goos == "windows" {
		extension = ".exe"
	}
	// tactical_windows_amd64.exe
	return fmt.Sprintf("%s_%s_%s%s", name, b.goos, b.arch, extension)
}

func (b *Builder) ensureSyso() error {
	if b.goos != "windows" {
		return nil
	}

	sysoPath := "cmd/tactical/resource_" + b.arch + ".syso"
	if err := sh.Rm(sysoPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing syso: %v", err)
	}

	err := sh.RunWithV(
		b.env(),
		b.tools["goversioninfo"],
		"-manifest="+assets["manifest"],
		"-icon="+assets["icon"],
		"-o", sysoPath,
		assets["versioninfo"],
	)
	if err != nil {
		return fmt.Errorf("failed to generate syso error: %v", err)
	}

	fmt.Printf("Syso generated at path '%s'\n", sysoPath)
	return nil
}

func (b *Builder) ldflags() string {
	commitID, _ := sh.Output("git", "rev-parse", "HEAD")

	flags := "-s -w"
	flags += " -X github.com/yakuter/go-mage/pkg/buildvars.Version=" + version
	flags += " -X github.com/yakuter/go-mage/pkg/buildvars.BuildTime=" + time.Now().Format(time.RFC3339)
	flags += " -X github.com/yakuter/go-mage/pkg/buildvars.CommitID=" + commitID
	flags += " -X github.com/yakuter/go-mage/pkg/buildvars.BuildMode=" + buildEnv()

	if b.goos == "linux" || b.goos == "windows" {
		flags += " -extldflags '-static'"
	}

	return flags
}

func (b *Builder) tags() string {
	return fmt.Sprintf("netgo osusergo %s", buildEnv())
}

func (b *Builder) isAdmin() bool {
	if b.goos == "windows" {
		if err := sh.Run("NET", "SESSION"); err == nil {
			println("Administrator privileges detected")
			return true
		}
	} else {
		if os.Getenv("SUDO_USER") != "" {
			println("Administrator privileges detected")
			return true
		}
	}
	return false
}

func (b *Builder) env() map[string]string {
	env := make(map[string]string)

	env["GOOS"] = b.goos
	env["GOARCH"] = b.arch
	env["CGO_ENABLED"] = os.Getenv("CGO_ENABLED")
	env["BUILD_ENV"] = buildEnv()

	if b.goos == "darwin" {
		env["CGO_CFLAGS"] = "-mmacosx-version-min=10.15"
		env["CGO_LDFLAGS"] = "-mmacosx-version-min=10.15"
		// MACOS_SDK_VERSION can be 10.15, 11.0, 14.5, 15.0 etc.
		env["SDKROOT"] = macosSDK(os.Getenv("MACOS_SDK_VERSION"))
	}

	if b.goos == "windows" {
		env["CGO_LDFLAGS"] = "-static"
	}

	// If we are cross compiling, set the right compiler.
	if b.goos == "windows" {
		if b.arch == "amd64" {
			if mingwxcompiler_exists() {
				env["CC"] = mingw_xcompiler
			}
		} else {
			if mingwxcompiler32_exists() {
				env["CC"] = mingw_xcompiler_32
			}
		}
	}

	return env
}

// Generate runs go generate.
func Generate() error {
	return Builder{goos: goos(), arch: goarch()}.Generate()
}

// Linter runs the linter.
func Linter() error {
	return Builder{goos: goos(), arch: goarch()}.Linter()
}

// VulnCheck runs the vulnerability check.
func Vulncheck() error {
	return Builder{goos: goos(), arch: goarch()}.Vulncheck()
}

// Test runs the tests.
func Test() error {
	return Builder{goos: goos(), arch: goarch()}.Test()
}

// Build builds the binary.
func Build() error {
	return Builder{goos: goos(), arch: goarch()}.Build()
}

// Clean cleans the build directory.
func Clean() error {
	return Builder{goos: goos(), arch: goarch()}.Clean()
}

func gopath() (string, error) {
	gopath, err := sh.Output(mg.GoCmd(), "env", "GOPATH")
	if err != nil {
		return "", fmt.Errorf("failed to get GOPATH error: %v", err)
	}
	if gopath == "" {
		u, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("failed to get user info error: %v", err)
		}
		gopath = filepath.Join(u.HomeDir, "go")
	}
	if err := os.MkdirAll(gopath, 0700); err != nil && !os.IsExist(err) {
		return "", fmt.Errorf("failed to create GOPATH '%s' error: %v", gopath, err)
	}

	return gopath, nil
}

func gobin(gopath string) (string, error) {
	// use GOBIN if set in the environment, otherwise fall back to first path
	// in GOPATH environment string
	gobin, err := sh.Output(mg.GoCmd(), "env", "GOBIN")
	if err != nil {
		return "", fmt.Errorf("failed to get GOBIN error: %v", err)
	}
	if gobin == "" {
		if gopath == "" {
			return "", fmt.Errorf("failed to get GOBIN, GOPATH is empty")
		}
		paths := strings.Split(gopath, string([]rune{os.PathListSeparator}))
		gobin = filepath.Join(paths[0], "bin")
	}
	if err := os.MkdirAll(gobin, 0700); err != nil && !os.IsExist(err) {
		return "", fmt.Errorf("failed to create GOBIN '%s' error: %v", gobin, err)
	}

	return gobin, nil
}

func buildEnv() string {
	env := os.Getenv("BUILD_ENV")
	switch env {
	case "dev":
		return "dev"
	case "prod":
		return "prod"
	default:
		return "prod"
	}
}

func macosSDK(env string) string {
	sdkPath := fmt.Sprintf("/Library/Developer/CommandLineTools/SDKs/MacOSX%s.sdk", env)
	_, err := os.Stat(sdkPath)
	if os.IsNotExist(err) {
		fmt.Printf("SDK path %s does not exist\n", sdkPath)
		return ""
	}
	return sdkPath
}

func goos() string {
	goos, err := sh.Output(mg.GoCmd(), "env", "GOOS")
	if err != nil {
		fmt.Printf("failed to get GOOS error: %v", err)
		goos = runtime.GOOS
	}
	return goos
}

func goarch() string {
	goarch, err := sh.Output(mg.GoCmd(), "env", "GOARCH")
	if err != nil {
		fmt.Printf("failed to get GOARCH error: %v", err)
		goarch = runtime.GOARCH
	}
	return goarch
}

func printGoenv() {
	goenv, err := sh.Output(mg.GoCmd(), "env")
	if err != nil {
		fmt.Printf("failed to get GOENV error: %v", err)
	}
	fmt.Println(goenv)
}

func mingwxcompiler_exists() bool {
	err := sh.Run(mingw_xcompiler, "--version")
	return err == nil
}

func mingwxcompiler32_exists() bool {
	err := sh.Run(mingw_xcompiler_32, "--version")
	return err == nil
}
