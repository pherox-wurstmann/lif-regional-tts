package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

func main() {
	logDirectory := os.Args[1]

	var lastReadLogDir string
	var lastReadLogFile string
	var lastReadLogLine int

	regionalRe := regexp.MustCompile(`You have gained <spush><color:.*>([0-9]+)<spop> <spush><color:.*>([a-zA-Z\ ]+)<spop> of <spush><color:.*>(.*)<spop>`)

	for {

		var latestLogDir, latestLogFile string
		var latestLogDirModified, latestLogFileModified time.Time = time.Unix(0, 0), time.Unix(0, 0)

		logDirs, err := ioutil.ReadDir(logDirectory)
		if err != nil {
			log.Fatal("cannot read log directory", err)
		}

		for _, logDir := range logDirs {
			if logDir.ModTime().After(latestLogDirModified) {
				latestLogDir = logDir.Name()
				latestLogDirModified = logDir.ModTime()
			}
		}

		currentLogDir := strings.Join([]string{logDirectory, latestLogDir}, string(os.PathSeparator))
		if currentLogDir != lastReadLogDir {
			lastReadLogLine = -1
		}
		lastReadLogDir = currentLogDir
		logFiles, err := ioutil.ReadDir(lastReadLogDir)
		if err != nil {
			log.Fatal("cannot read log directory", err)
		}

		for _, logFile := range logFiles {
			if logFile.ModTime().After(latestLogFileModified) {
				latestLogFile = logFile.Name()
				latestLogFileModified = logFile.ModTime()
			}
		}

		currentLogFile := strings.Join([]string{lastReadLogDir, latestLogFile}, string(os.PathSeparator))
		if currentLogFile != lastReadLogFile {
			lastReadLogLine = -1
		}
		lastReadLogFile = currentLogFile
		logFileContents, err := ioutil.ReadFile(lastReadLogFile)
		if err != nil {
			log.Fatal("cannot read log file", err)
		}

		logLines := strings.Split(string(logFileContents), "\r\n")
		for index, logLine := range logLines {
			if index <= lastReadLogLine {
				continue
			}
			// do not tts on first run
			if lastReadLogLine != -1 {
				matches := regionalRe.FindStringSubmatch(logLine)
				if matches != nil {
					log.Printf("%q\n", matches)
					tts := string(matches[1]) + " " + string(matches[2]) + " from " + string(matches[3])
					cmd := "PowerShell -Command \"Add-Type -AssemblyName System.Speech; (New-Object System.Speech.Synthesis.SpeechSynthesizer).Speak('" + tts + "');\""
					batFileName := logDirectory + string(os.PathSeparator) + ".." + string(os.PathSeparator) + "tts.bat"
					ioutil.WriteFile(batFileName, []byte(cmd), 0777)
					defer os.Remove(batFileName)
					if _, err := exec.Command(batFileName).CombinedOutput(); err != nil {
						log.Fatal("tts failed: ", err)
					}
				}
			}
		}

		lastReadLogLine = len(logLines)
		time.Sleep(3 * time.Second)
	}
}
