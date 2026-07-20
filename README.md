# StackDrift CLI

A command line tool for StackDrift. It scans a project directory, detects the
technologies and dependency manifests that StackDrift supports, and adds them to
one of your StackDrift projects. StackDrift then tracks versions, end of life
dates, and security advisories for you.

## Install

### Linux/MacOS

```
curl -fsSL https://raw.githubusercontent.com/digitalaffinity-au/stackdrift-cli/main/scripts/install.sh | bash
```

This installs the binary into a directory that is already on your PATH, such as
`~/.local/bin` or `/usr/local/bin`, so you can run `stackdrift` from anywhere
without changing any environment variables. Set `STACKDRIFT_INSTALL_DIR` to
force a specific directory.

The Linux script also works on macOS. It picks the right binary for Intel or
Apple Silicon automatically.

### Windows

Open PowerShell and run:

```
irm https://raw.githubusercontent.com/digitalaffinity-au/stackdrift-cli/main/scripts/install.ps1 | iex
```

This installs the binary into `%LOCALAPPDATA%\Microsoft\WindowsApps`, which is
already on your PATH, so you can run `stackdrift` from anywhere without changing
any environment variables.

## Updating

To upgrade to the latest release:

```
stackdrift update
```

It downloads the newest binary for your platform and replaces the one you are
running. If you are already on the latest version it does nothing. Pass
`--force` to reinstall anyway. Install the CLI somewhere you can write to, such
as the default `~/.local/bin`, so the update can replace it in place without
extra permissions.

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

- npm: `package.json`
- NuGet: `.csproj`

Each project becomes its own dependency group. Lock and version files next to a
manifest are included automatically so versions are pinned: `package-lock.json`
for npm, and `packages.lock.json` plus `Directory.Packages.props` for NuGet. A
solution with four `.csproj` files produces four groups.

Folders like `node_modules`, `bin`, `obj`, and `.git` are skipped.

## Other commands

```
stackdrift status    show the tracked technologies and dependencies
stackdrift check     report CVE status and exit non-zero if any are found
stackdrift remove    remove technologies or dependencies from the project
stackdrift whoami    show the signed in account
stackdrift update    download and install the latest release
stackdrift version   print the CLI version
```

## Pointing at a different server

The CLI always talks to the public StackDrift server at https://stackdrift.net.
The only way to point it at a different server is the `STACKDRIFT_URL`
environment variable at runtime:

```
STACKDRIFT_URL=http://localhost:5000 stackdrift login
```

## Building from source

You need Go installed. To build release binaries for Linux, Windows, and macOS
(amd64 and arm64) into `dist/`:

```
scripts/build.sh 0.1.0
```

Every binary targets https://stackdrift.net. There is no build-time server
option; use the `STACKDRIFT_URL` environment variable to point at another
server at runtime.

To run the tests:

```
go test ./...
```
