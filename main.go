package main
import (
    "github.com/kardianos/osext"
    "strings"
    "fmt"
    "os"
    "path/filepath"
    "runtime"
    "os/exec"
    "io/ioutil"
    "syscall"
)

func fail(message string, args ...interface{}) {
    fmt.Fprintf(os.Stderr, message + "\n\n", args...)
    os.Exit(1)
}

func usageAndFail() {
    fail("Usage: %s <install|uninstall> <targetExecutable...>", os.Args[0])
}

func getCwdOrFail() string {
    result, err := os.Getwd();
    if err != nil {
        fail("Could not get current working directory. Got: %s", err.Error())
    }
    return result
}

func createTargetExecutableFor(sourceExecutable string) string {
    sourceExecutableDir := filepath.Dir(sourceExecutable)
    sourceExecutableFile := filepath.Base(sourceExecutable)

    targetExecutableDir := sourceExecutableDir
    targetExecutableFile := "." + sourceExecutableFile
    targetExecutable := filepath.Join(targetExecutableDir, targetExecutableFile)

    return targetExecutable
}

func createSourceExecutableFor(targetExecutable string) (string, bool) {
    targetExecutableDir := filepath.Dir(targetExecutable)
    targetExecutableFile := filepath.Base(targetExecutable)

    sourceExecutableDir := targetExecutableDir
    if !strings.HasPrefix(targetExecutableFile, ".") || len(targetExecutableFile) < 1 {
        return "", false
    }
    sourceExecutableFile := targetExecutableFile[1:]
    sourceExecutable := filepath.Join(sourceExecutableDir, sourceExecutableFile)

    return sourceExecutable, true
}

func isExecutable(filePath string) bool {
    fileInfo, err := os.Lstat(filePath)
    if err != nil {
        return false
    }
    if runtime.GOOS == "windows" {
        ext := strings.ToLower(filepath.Ext(filePath))
        if ext == ".bat" || ext == ".cmd" || ext == ".exe" {
            return true
        }
        return false
    }
    return hasPermission(0555, fileInfo.Mode())
}

func hasPermission(required os.FileMode, value os.FileMode) bool {
    if (value & required) == required {
        return true
    }
    return false
}

func makeAbsoluteOrExit(what string) string {
    result, err := filepath.Abs(what)
    if err != nil {
        fail("It is not possible to make '%s' absolute. Got: %s", what, err.Error())
    }
    return result
}

func absoluteLauncherFor(sourceExecutable string) string {
    return makeAbsoluteOrExit(sourceExecutable)
}

func launcherTargetFor(launcher string, source string) string {
    launcherDirectory := filepath.Dir(launcher)
    launcherFile := filepath.Base(launcher)
    launcherTargetDirectory, err := filepath.Rel(filepath.Dir(source), launcherDirectory)
    if err != nil {
        launcherTargetDirectory = launcherDirectory
    }
    launcherTarget := filepath.Join(launcherTargetDirectory, launcherFile)
    return launcherTarget;
}

func getAbsoluteSymlinkOf(file string) (string, bool) {
    absoluteFile, err := filepath.Abs(file)
    if err != nil {
        return "", false
    }
    fileInfo, err := os.Lstat(absoluteFile)
    if err != nil {
        fail("Cannot get absolute symlink. %v", err)
    }
    if (fileInfo.Mode() & os.ModeSymlink) != 0 {
        target, err := os.Readlink(absoluteFile)
        if err != nil {
            fail("Cannot get absolute symlink. %v", err)
        }
        oldCwd := getCwdOrFail()
        if err := os.Chdir(filepath.Dir(file)); err != nil {
            fail("Cannot get absolute symlink. %v", err)
        }
        absolute, err := filepath.Abs(target)
        if err := os.Chdir(oldCwd); err != nil {
            fail("Cannot get absolute symlink. %v", err)
        }
        return absolute, true
    }
    return "", false
}

func install(candidate string, absoluteLauncher string) {
    launcherTarget := launcherTargetFor(absoluteLauncher, candidate)
    candidateTarget := createTargetExecutableFor(candidate)

    err := os.Rename(candidate, candidateTarget)
    if err != nil {
        fail("Cannot rename candidate to its new name. Got: %v", err)
    }

    oldCwd := getCwdOrFail()
    os.Chdir(filepath.Dir(launcherTarget))
    err = os.Symlink(launcherTarget, candidate)
    if err != nil {
        fail("Could not place launcher on place of candiate. Got: %v", err)
    }
    os.Chdir(oldCwd)

    fmt.Fprintf(os.Stdout, "Launcher installed on %s.\n", candidate)
}

func isLauncher(what string, absoluteLauncher string) bool {
    if linkTarget, ok := getAbsoluteSymlinkOf(what); ok {
        if (linkTarget == absoluteLauncher) {
            return true
        }
    }
    return false
}

func checkAndInstall(candidate string, absoluteLauncher string) {
    if isLauncher(candidate, absoluteLauncher) {
        return
    }
    possibleSourceCandidate, ok := createSourceExecutableFor(candidate)
    if ok && isLauncher(possibleSourceCandidate, absoluteLauncher) {
        return
    }
    install(candidate, absoluteLauncher)
}

func installOnAllOf(this string, candidatePatterns []string) {
    absoluteLauncher := absoluteLauncherFor(this)
    for _, candidatesPattern := range candidatePatterns {
        candidates, err := filepath.Glob(candidatesPattern)
        if err != nil {
            fail("Could not install for '%s'. Got: %v", candidatesPattern, err)
        }
        for _, plainCandidate := range candidates {
            if isExecutable(plainCandidate) {
                checkAndInstall(makeAbsoluteOrExit(plainCandidate), absoluteLauncher)
            }
        }
    }
}

func uninstall(candidate string, absoluteLauncher string) {
    candidateTarget := createTargetExecutableFor(candidate)
    if err := os.Remove(candidate); err != nil {
        fail("Could not remove current launcher. Got: %v", err)
    }
    if err := os.Rename(candidateTarget, candidate); err != nil {
        fail("Could not move the candidate '%s' back to its old name '%s'. Got: %v", candidateTarget, candidate, err)
    }

    fmt.Fprintf(os.Stdout, "Launcher removed from %s.\n", candidate)
}

func checkAndUninstall(candidate string, absoluteLauncher string) {
    if isLauncher(candidate, absoluteLauncher) {
        uninstall(candidate, absoluteLauncher)
    }
}

func uninstallOnAllOf(this string, candidatePatterns []string) {
    absoluteLauncher := absoluteLauncherFor(this)
    for _, candidatesPattern := range candidatePatterns {
        candidates, err := filepath.Glob(candidatesPattern)
        if err != nil {
            fail("Could not install for '%s'. Got: %v", candidatesPattern, err)
        }
        for _, plainCandidate := range candidates {
            if isExecutable(plainCandidate) {
                checkAndUninstall(makeAbsoluteOrExit(plainCandidate), absoluteLauncher)
            }
        }
    }
}

func execute(targetExecutable string) {
    arguments := []string{targetExecutable}
    if (len(os.Args) > 1) {
        arguments = append(arguments, os.Args[1:]...)
    }

    command := exec.Cmd{
        Path: targetExecutable,
        Args: arguments,
        Stdin: os.Stdin,
        Stdout: os.Stdout,
        Stderr: os.Stderr,
    }

    var waitStatus syscall.WaitStatus
    if err := command.Run(); err != nil {
        if exitError, ok := err.(*exec.ExitError); ok {
            waitStatus = exitError.Sys().(syscall.WaitStatus)
            os.Exit(int(waitStatus.ExitStatus()))
        } else {
            fail("Could not start process. Got: %v", err)
        }
    } else {
        waitStatus = command.ProcessState.Sys().(syscall.WaitStatus)
        os.Exit(int(waitStatus.ExitStatus()))
    }
}

func isEnabled(value string) bool {
    trimmed := strings.TrimSpace(strings.ToLower(value))
    if trimmed == "yes" || trimmed == "y" || trimmed == "true" || trimmed == "1" || trimmed == "on" {
        return true
    }
    return false
}

func isExecutionAllowed() bool {
    if byteContent, err := ioutil.ReadFile("/etc/oracle-javase-license-accepted"); err == nil {
        if isEnabled(string(byteContent)) {
            return true
        }
    }
    if isEnabled(os.Getenv("ORACLE_JAVASE_LICENSE_ACCEPTED")) {
        return true
    }
    return false
}

func permitExecutionAndDisplayInstructions() {
    fmt.Fprint(os.Stderr, "You must accept the 'Oracle Binary Code License Agreement for Java SE' to use this software.\n" +
    "\n" +
    "To do this, follow the steps below:\n" +
    "1# Read http://www.oracle.com/technetwork/java/javase/terms/license/index.html in your web browser.\n" +
    "2# Accept the license by choosing one of the following possibilities:\n" +
    "\ta) Place a readable file named '/etc/oracle-javase-license-accepted' with content 'yes'.\n" +
    "\tOR\n" +
    "\tb) Provide an environment variable 'ORACLE_JAVASE_LICENSE_ACCEPTED' with content 'yes'.\n" +
    "\n" +
    "This program will exit now.\n")
    os.Exit(127)
}

func runAsInstallerMode(this string) {
    if len(os.Args) <= 2 {
        usageAndFail()
    }
    command := strings.ToLower(strings.TrimSpace(os.Args[1]))
    if command == "install" || command == "i" {
        installOnAllOf(this, os.Args[2:])
    } else if command == "uninstall" || command == "u" || command == "remove" || command == "r" {
        uninstallOnAllOf(this, os.Args[2:])
    } else {
        usageAndFail()
    }
}

func resolveThisExecutable() (string, error) {
    if runtime.GOOS == "windows" {
        return osext.Executable()
    } else {
        this := os.Args[0]
        if this[0] == '/' || this[0] == '.' {
            return makeAbsoluteOrExit(this), nil
        } else {
            execIsInPath, err := exec.LookPath(this)
            if err != nil {
               return "", err
            }
            return execIsInPath, nil
        }
    }
}

func main() {
    this, err := resolveThisExecutable()
    if err != nil {
        fail("Could not get name of myself. Got: %v", err)
    }
    if _, ok := getAbsoluteSymlinkOf(this); ok {
        targetExecutable := createTargetExecutableFor(this)
        if _, err := os.Stat(targetExecutable); err == nil {
            if isExecutionAllowed() {
                execute(targetExecutable)
            } else {
                permitExecutionAndDisplayInstructions()
            }
        } else {
            runAsInstallerMode(this)
        }
    } else {
        runAsInstallerMode(this)
    }
}
