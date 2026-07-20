# StackDrift CLI

A command line tool for StackDrift. It scans a project directory, detects the
technologies and dependency manifests that StackDrift supports, and adds them to
one of your StackDrift projects. StackDrift then tracks versions, end of life
dates, and security advisories for you.

## Install

### Linux

```
curl -fsSL https://raw.githubusercontent.com/digitalaffinity-au/stackdrift-cli/main/scripts/install.sh | bash
```

This downloads the binary to `~/.local/bin/stackdrift`. If that directory is not
on your PATH, the script tells you how to add it.

The Linux script also works on macOS. It picks the right binary for Intel or
Apple Silicon automatically.

### Windows

Open PowerShell and run:

```
irm https://raw.githubusercontent.com/digitalaffinity-au/stackdrift-cli/main/scripts/install.ps1 | iex
```

This downloads the binary to `%LOCALAPPDATA%\StackDrift` and adds it to your
PATH.

## Sign in

```
stackdrift login
```

This prints a link and a short code. Open the link in your browser, sign in to
StackDrift, and confirm the code matches. The CLI waits until you approve, then
saves a token so you do not need to sign in again. The token is stored in your
user config directory, not in your project.

To sign out:

```
stackdrift logout
```

## Track a directory

From inside a project directory:

```
stackdrift scan
```

The first time, it asks whether to add the directory to an existing project or
create a new one. It then lists the technologies and dependency manifests it
found. Use the numbers to toggle items on or off, then press Enter. The CLI adds
the selected items to your project and writes a `.stackdrift` file that records
what is tracked.

Commit `.stackdrift` to your repository. On later runs the CLI reads it and does
not ask you to pick a project again.

To accept everything without prompts, for example on a first automated run:

```
stackdrift scan --yes
```

This needs the project to be chosen once interactively first, or a `.stackdrift`
file already present.

## Check for CVEs in CI

```
stackdrift check
```

This prints the CVE status of the project and exits with a non-zero code if any
tracked technology or dependency has a known CVE. Use it in a pipeline to fail a
build when a new advisory appears.

## What it detects

Technologies:

- .NET Full Framework and .NET Core SDK, from `.csproj` target frameworks
- .NET Core Runtime, from a Dockerfile base image
- Laravel, from `composer.json`
- The host operating system, from `/etc/os-release`
- The Linux kernel version
- Operating systems named in a Dockerfile `FROM` line

Dependency manifests:

- npm: `package.json` and `package-lock.json`
- NuGet: `.csproj`, `packages.lock.json`, and `Directory.Packages.props`

Folders like `node_modules`, `bin`, `obj`, and `.git` are skipped.

## Other commands

```
stackdrift status    show the tracked technologies and dependencies
stackdrift check     report CVE status and exit non-zero if any are found
stackdrift remove    remove technologies or dependencies from the project
stackdrift whoami    show the signed in account
stackdrift version   print the CLI version
```

## Pointing at a different server

Release builds talk to the public StackDrift server. To use a different server,
set an environment variable:

```
STACKDRIFT_URL=http://192.168.1.47 stackdrift login
```

## Building from source

You need Go installed. To build release binaries for Linux and Windows into
`dist/`:

```
STACKDRIFT_BUILD_URL=https://stackdrift.net scripts/build.sh 0.1.0
```

To run the tests:

```
go test ./...
```
