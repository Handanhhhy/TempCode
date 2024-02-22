package main

import (
	"flag"
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var (
	cpuThreshold = flag.Float64("cpu", 20.0, "CPU Threshold in percentage")
	memThreshold = flag.Float64("mem", 50.0, "Memory Threshold in percentage")
	timeout      = flag.Int("timeout", 30, "stress default exit time")
	interval     = flag.Float64("interval", 10, "buffer interval")
	monTime      = flag.Int("monTime", 5, "moniter time")
)

func getCPUPercentage() float64 {
	out, err := exec.Command("sh", "-c", "top -bn1 | grep 'Cpu(s)' | sed 's/.*, *\\([0-9.]*\\)%* id.*/\\1/'").Output()
	if err != nil {
		fmt.Println(err)
		return 0.0
	}

	cpuPercentageStr := strings.TrimSpace(string(out))
	cpuPercentage, err := strconv.ParseFloat(cpuPercentageStr, 64)
	if err != nil {
		fmt.Println(err)
		return 0.0
	}

	return 100 - cpuPercentage
}

func getMemPercentage() float64 {
	out, err := exec.Command("sh", "-c", "free | grep Mem | awk '{print $3/$2 * 100.0}'").Output()
	if err != nil {
		fmt.Println(err)
		return 0.0
	}

	memPercentageStr := strings.TrimSpace(string(out))
	memPercentage, err := strconv.ParseFloat(memPercentageStr, 64)
	if err != nil {
		fmt.Println(err)
		return 0.0
	}

	return memPercentage
}

func killProcessesByCommand(command string) {
	out, err := exec.Command("ps", "aux").Output()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, command) {
			fields := strings.Fields(line)
			pid := fields[1]
			exec.Command("kill", "-9", pid).Run()
			fmt.Printf("Killed %s process with PID %s\n", command, pid)
			break
		}
	}
}

func getTotalMemory() int {
	out, err := exec.Command("sh", "-c", "free -g | grep Mem | awk '{print $2}'").Output()
	if err != nil {
		fmt.Println(err)
		return 0
	}

	totalMemoryStr := strings.TrimSpace(string(out))
	totalMemory, err := strconv.Atoi(totalMemoryStr)
	if err != nil {
		fmt.Println(err)
		return 0
	}

	return totalMemory
}

func main() {
	flag.Parse()
	fmt.Printf("CPU Threshold : %.2f%%\n", *cpuThreshold)
	fmt.Printf("Memory Threshold : %.2f%%\n", *memThreshold)
	totalMemory := getTotalMemory()
	for {
		cpu := getCPUPercentage()
		mem := getMemPercentage()
		fmt.Printf("[%s]\n", time.Now().Format("2006-01-02 15:04:05"))
		fmt.Printf("Current CPU usage: %.2f%%\n", cpu)
		fmt.Printf("Current Memory usage: %.2f%%\n", mem)

		if cpu < *cpuThreshold {
			fmt.Println("CPU usage is below threshold! Sending alert...")
			stressCommand := fmt.Sprintf("stress --cpu 1 --timeout %d &", *timeout)
			exec.Command("sh", "-c", stressCommand).Run()
		}

		if cpu > *cpuThreshold+*interval {
			killProcessesByCommand("stress --cpu")
		}

		if mem < *memThreshold {
			fmt.Println("Memory usage is below threshold! Sending alert...")
			memoryToAddPercentage := float64(*memThreshold) - mem
			memoryToAdd := int(math.Ceil(float64(totalMemory) * memoryToAddPercentage / 100))
			fmt.Printf("To reach the memory threshold of %.2f%%, you need to add %dGB of memory.\n", *memThreshold, memoryToAdd)
			if memoryToAdd > 0 {
				stressCommand := fmt.Sprintf("stress --vm 1 --vm-bytes %dG --timeout %d &", memoryToAdd, *timeout)
				exec.Command("sh", "-c", stressCommand).Run()
			}
		}

		if mem > *memThreshold+*interval {
			killProcessesByCommand("stress --vm")
		}
		time.Sleep(time.Duration(*monTime) * time.Second)
	}
}
