package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/masanorihirano/gddl/gddl"
	"os"
	"strconv"
	"time"
)

func showUsage() {
	fmt.Printf("Usage (only arguments):\n" +
		"\tShow all repositories:\n\t\tshow\n" +
		"\tShow folders in a repository:\n\t\tshow [repository]\n" +
		"\tShow download candidates in a folder:\n\t\tshow [repository] [folder]\n" +
		"\tDownload:\n\t\tdownload [repository] [folder] [file] [path(optional)]\n" +
		"\tUpload:\n\t\tupload [repository] [folder] [file/folder]\n")
}

func posString(slice []string, element string) int {
	for index, elem := range slice {
		if elem == element {
			return index
		}
	}
	return -1
}

func containsString(slice []string, element string) bool {
	return !(posString(slice, element) == -1)
}

func selectMenu() {
	var mode int
	fmt.Println("Please choose what you want:")
	fmt.Println("0. show usage")
	fmt.Println("1. data download")
	fmt.Println("2. data upload")
	fmt.Print("type 0/1/2: ")
	for {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		i, err := strconv.Atoi(scanner.Text())
		if err == nil {
			if i >= 0 && i < 3 {
				mode = i
				break
			}
		}
		fmt.Print("type correctly from 0/1/2: ")
	}
	fmt.Println("your choice:", mode)
	fmt.Println()
	for {
		if mode == 0 {
			showUsage()
			break
		}
		var repository string
		repositories := gddl.ListRepository()
		fmt.Println("Please choose repository:")
		for j, key := range repositories {
			fmt.Println(fmt.Sprintf("%d: %s", j, key))
		}
		fmt.Print(fmt.Sprintf("please type your choice (0-%d): ", len(repositories)-1))
		for {
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			i, err := strconv.Atoi(scanner.Text())
			if err == nil {
				if i >= 0 && i < len(repositories) {
					for j, key := range repositories {
						if i == j {
							repository = key
							break
						}
					}
					break
				}
			}
			fmt.Print(fmt.Sprintf("please type your choice correctly from 0-%d: ", len(repositories)-1))
		}
		fmt.Println(fmt.Sprintf("your choice: %s", repository))
		fmt.Println()
		var directory string
		directories, err := gddl.ListDirectory(repository)
		if err != nil {
			panic(err)
		}
		fmt.Println("Please choose directory:")
		for j, key := range directories {
			fmt.Println(fmt.Sprintf("%d: %s", j, key))
		}
		fmt.Print(fmt.Sprintf("please type your choice (0-%d): ", len(directories)-1))
		for {
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			i, err := strconv.Atoi(scanner.Text())
			if err == nil {
				if i >= 0 && i < len(directories) {
					for j, key := range directories {
						if i == j {
							directory = key
							break
						}
					}
					break
				}
			}
			fmt.Print(fmt.Sprintf("please type your choice correctly from 0-%d: ", len(directories)-1))
		}
		fmt.Println(fmt.Sprintf("your choice: %s", directory))
		fmt.Println()
		if mode == 1 {
			var fileName string
			fileList, err := gddl.ListFiles(repository, directory)
			fmt.Println("Please choose file/folder:")
			for j, key := range fileList {
				fmt.Println(fmt.Sprintf("%d: %s", j, key))
			}
			fmt.Print(fmt.Sprintf("please type your choice (0-%d): ", len(fileList)-1))
			for {
				scanner := bufio.NewScanner(os.Stdin)
				scanner.Scan()
				i, err := strconv.Atoi(scanner.Text())
				if err == nil {
					if i >= 0 && i < len(fileList) {
						for j, key := range fileList {
							if i == j {
								fileName = key
								break
							}
						}
						break
					}
				}
				fmt.Print(fmt.Sprintf("please type your choice correctly from 0-%d: ", len(fileList)-1))
			}
			fmt.Println(fmt.Sprintf("your choice: %s", fileName))
			fmt.Println()
			dir, err := os.Getwd()
			if err != nil {
				panic(err)
			}
			err = gddl.DownloadAndSave(dir, repository, directory, fileName, false, true)
			if err != nil {
				panic(err)
			}
			time.Sleep(time.Second * 1)
			fmt.Println("Download was finished.")
			fmt.Print("Do you want to download the next file? [y/n]: ")
			var goNext bool
			for {
				scanner := bufio.NewScanner(os.Stdin)
				scanner.Scan()
				yesResponses := []string{"y", "Y", "yes", "Yes", "YES"}
				noResponses := []string{"n", "N", "no", "No", "NO"}
				response := scanner.Text()
				if containsString(yesResponses, response) {
					goNext = true
					break
				} else if containsString(noResponses, response) {
					goNext = false
					break
				}
				fmt.Print("please type 'y' or 'n': ")
			}
			if goNext {
				continue
			} else {
				break
			}
		} else if mode == 2 {
			fmt.Println("Currently this mode is disabled.")
			break
		}
	}
	fmt.Println("This is the end of this program. To close this, please type anything")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
}

func main() {
	flag.Parse()
	if flag.Arg(0) == "" {
		selectMenu()
		return
	}
	if flag.Arg(0) == "show" {
		if flag.Arg(1) == "" {
			printStr := ""
			for _, key := range gddl.ListRepository() {
				if printStr == "" {
					printStr = key
				} else {
					printStr += "\n" + key
				}
			}
			fmt.Print(printStr)
		} else {
			includes := false
			for _, key := range gddl.ListRepository() {
				if key == flag.Arg(1) {
					includes = true
					break
				}
			}
			if !includes {
				fmt.Printf("Cannot found repository: %s", flag.Arg(1))
				return
			}
			if flag.Arg(2) == "" {
				printStr := ""
				directories, err := gddl.ListDirectory(flag.Arg(1))
				if err != nil {
					panic(err)
				}
				for _, key := range directories {
					if printStr == "" {
						printStr = key
					} else {
						printStr += "\n" + key
					}
				}
				fmt.Print(printStr)
			} else {
				if flag.Arg(3) == "" {
					includes := false
					directories, err := gddl.ListDirectory(flag.Arg(1))
					if err != nil {
						panic(err)
					}
					for _, key := range directories {
						if key == flag.Arg(2) {
							includes = true
							break
						}
					}
					if !includes {
						fmt.Printf("Cannot found directory: %s", flag.Arg(2))
						return
					}
					printStr := ""
					fileList, err := gddl.ListFiles(flag.Arg(1), flag.Arg(2))
					if err != nil {
						panic(err)
					}
					for _, key := range fileList {
						if printStr == "" {
							printStr = key
						} else {
							printStr += "\n" + key
						}
					}
					fmt.Print(printStr)
				} else {
					showUsage()
					fmt.Print("Did you mean 'download'?")
					return
				}
			}
		}
		return
	} else if flag.Arg(0) == "download" {
		if len(flag.Args()) > 4 {
			showUsage()
			return
		} else {
			var dir string
			var err error
			if flag.Arg(4) != "" {
				dir = flag.Arg(4)
			} else {
				dir, err = os.Getwd()
				if err != nil {
					panic(err)
				}
			}
			err = gddl.DownloadAndSave(dir, flag.Arg(1), flag.Arg(2), flag.Arg(3), false, true)
			if err != nil {
				panic(err)
			}
		}
	}
}
