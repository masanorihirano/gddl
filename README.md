# gddl: Google Drive data DL tool for Izumi Lab.

## Requirements
No requirement if you use binary on release page.

## Setup
### pixz installation
If you use folder upload mode, pixz install is highly recommended.

For Mac:
```bash
brew install pixz
```
For linux
```bash
sudo apt install pixz
```
or build using https://github.com/vasi/pixz

### go install
If you use go binary via `go install` command, golang installation is required.
Then, you can get gddl via:
```bash
go install github.com/masanorihirano/gddl@lates
```


## usage
Interactive mode:
```bash
gddl
```

Show help:
```bash
gddl help
```
Show version:
```bash
gddl version
```
Show all repositories:
```bash
gddl show
```
Show folders in a repository:
```bash
gddl show [repository]
```
Show download candidates in a folder:
```bash
gddl show [repository] [folder]
```
Download:
```bash
gddl download [repository] [folder] [file] [path(optional)]
```
Upload:
```bash
gddl upload [repository] [folder] [file/folder]
```
## Note
If you use this program for the first time, it requires authorization of Google API.
Please follow the leads, and log in with account under the control of socsim.org with access right to the team drive.

The maximum number of files in one directory (appears and be able to be downloaded in this system) is limited to 1,000.
Please don't place more than 1,000 files in one directory on Google team drive.

## Author
Masanori HIRANO (https://mhirano.jp/; b2018mhirano@socsim.org)
