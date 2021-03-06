package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	print "github.com/fatih/color"
)

var (
	index      int = -1
	url        string
	panelPaths []string

	marks = []string{
		" type=\"password\" ",
		" name=\"username\" ",
	}
)

func main() {

	print.HiRed(`
	##################################### 
	#        Admin Panel Finder         #
	#-----------------------------------#
	#       github.com/dursunkatar      #
	#####################################`)

	url = os.Args[1]
	panelPath := os.Args[2]

	if !strings.HasSuffix(url, "/") {
		url = url + "/"
	}

	err := loadPanels(panelPath)

	if err != nil {
		fmt.Println("Error: ", err.Error())
		return
	}

	goCount := 10
	chFound := make(chan string, 2)
	chThisIsNot := make(chan bool, 1)
	chFinished := make(chan bool, 1)

	fmt.Println("")
	fmt.Println("Panel Url Count: ", len(panelPaths))
	fmt.Println("")
	fmt.Println("Started...")

	for i := 0; i < goCount; i++ {
		go doControl(chFound, chThisIsNot, chFinished)
	}

exitLOOP:
	for {
		select {
		case mark := <-chFound:
			fmt.Print("MARK  : ")
			print.Cyan(strings.Trim(mark, " "))
			fmt.Print("FOUND : ")
			print.Green(<-chFound)
			break exitLOOP
		case <-chThisIsNot:
			go doControl(chFound, chThisIsNot, chFinished)
		case <-chFinished:
			break exitLOOP
		}
	}
	
	fmt.Println("Finish")
}

func doControl(found chan string, thisIsNot, finish chan bool) {
	index++
	if index >= len(panelPaths) {
		finish <- true
		return
	}

	if ok, path, mark := connectUrl(); ok {
		found <- mark
		found <- path
	} else {
		thisIsNot <- true
	}
}

func connectUrl() (bool, string, string) {

	path := panelPaths[index]
	_url := url + path

	req, err := http.NewRequest("GET", _url, nil)

	if err != nil {
		return false, "", ""
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:68.0) Gecko/20100101 Firefox/68.0")

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return false, "", ""
	}
	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)
	result := string(body)

	if ok, mark := isThis(&result); ok {
		return true, path, mark
	}
	return false, "", ""
}

func isThis(source *string) (bool, string) {
	for _, mark := range marks {
		if strings.Contains(*source, mark) {
			return true, mark
		}
	}
	return false, ""
}

func loadPanels(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		panelUrl := strings.Trim(scanner.Text(), " /\r")
		if !panelPathContains(panelUrl) {
			panelPaths = append(panelPaths, panelUrl)
		}
	}

	file.Close()
	return nil
}

func panelPathContains(s string) bool {
	for _, path := range panelPaths {
		if s == path {
			return true
		}
	}
	return false
}
