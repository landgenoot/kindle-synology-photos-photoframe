package main

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"time"
)

// Suspend device and use real time clock alarm to wake it up.
// If our wake up time is more or less 24 hours away, we can put it to
// sleep immediately. Otherwise, we will wait another 30 seconds, which enables us
// to abort the process.
func suspendToRam(duration int) {
	if runtime.GOARCH != "arm" {
		return // Skip if not on Kindle
	}
	cmd1 := exec.Command("sh", "-c", "echo \"\" > /sys/class/rtc/rtc1/wakealarm")
	err1 := cmd1.Run()
	if err1 != nil {
		log.Fatal(err1)
	}
	cmd2 := exec.Command("sh", "-c", fmt.Sprintf("echo \"+%d\" > /sys/class/rtc/rtc1/wakealarm", duration))
	err2 := cmd2.Run()
	if err2 != nil {
		log.Fatal(err2)
	}

	// Check if we are waken up manually, give us time to abort the process
	if duration < 3600*24-60 {
		log.Println("Waiting 30 seconds before going back to sleep")
		time.Sleep(30 * time.Second)
	}

	log.Println("Suspending to RAM")

	cmd3 := exec.Command("sh", "-c", "echo \"mem\" > /sys/power/state")
	err3 := cmd3.Run()
	if err3 != nil {
		log.Fatal(err3)
	}
}

// Count seconds till next wake up time. Formatted as clock
// time in 24H format. E.g. 6, 30 means 6:30 AM.
func nextWakeup(now time.Time, hour int, minutes int) int {
	yyyy, mm, dd := now.Date()
	if now.Hour() > hour || now.Hour() == hour && now.Minute() >= minutes {
		dd++ // Jump to tomorrow, if wakeup time has already passed.
	}
	tomorrow := time.Date(yyyy, mm, dd, hour, minutes, 0, 0, now.Location())
	return int(tomorrow.Sub(now).Seconds())
}

func drawToScreen(imagePath string) {
	if runtime.GOARCH != "arm" {
		return // Skip if not on Kindle
	}
	cmd := exec.Command("/usr/sbin/eips", "-f", "-g", imagePath)
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func parseBatteryLevel(state string) (int, error) {
	re := regexp.MustCompile(`\d*%`)
	value := re.FindString(state)
	if value == "" {
		err := fmt.Errorf("Could not parse battery level %s", state)
		return -1, err
	}
	numericValue := value[:len(value)-1] // Ommit % manually, because golang Regex does not support lookarounds.
	i, _ := strconv.Atoi(numericValue)
	return i, nil
}
