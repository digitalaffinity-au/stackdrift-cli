# Features

The CLI must:

  - be installable via a link directly to either an .sh file for linux  or a ps1 file for windows that will install the required binary.
  - Allow the user to log in via a link they click which directs them to the stackdrift website to login and obtain an API token.
  - Allow the user to execute the binary from any directory, which it will then ask them if they want to add this directory to an existing project or create a new project
  - It will then auto-detect the technologies that StackDrift currently supports and then list the detected technologies and dependencies. The user can then toggle on and off the ones they want to add
  - Will write a .stackdrift config file locally on what tech/dependencies are included
  - Detect previous runs of the CLI so we don't ask them to choose a project next time round
  - Add/Remove technology/dependencies
  - Be configurable so in development it points to http://192.168.1.47/ but the release on github points to stackdrift.net
  - Provide a check command that reports the project CVE status and exits non-zero when a tracked technology or dependency has a known CVE, so it can fail a CI pipeline.
  - Support a non-interactive flag on scan that accepts all detected items without prompting, for scripts and first runs.
 
There should also be a dist folder that has the compiled go binaries for linux, windows, and macOS, on both amd64 and arm64 where the platform has both.

Have a README.md file that instructs users how to set up StackDrift. No emojis, emdashes or metaphores. Plain English and simple sentances.
