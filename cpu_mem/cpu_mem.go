package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var lower_cpu float64 = 10
var lower_mem float64 = 20
var stressPID int

func getCPUPercentage() float64 {
	out, err := exec.Command("sh", "-c", "top -bn1 | grep 'Cpu(s)' | sed 's/.*, *\\([0-9.]*\\)%* id.*/\\1/'").Output()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cpuPercentageStr := strings.TrimSpace(string(out))
	cpuPercentage, err := strconv.ParseFloat(cpuPercentageStr, 64)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return 100 - cpuPercentage
}

func getMemPercentage() float64 {
	out, err := exec.Command("sh", "-c", "free | grep Mem | awk '{print $3/$2 * 100.0}'").Output()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	memPercentageStr := strings.TrimSpace(string(out))
	memPercentage, err := strconv.ParseFloat(memPercentageStr, 64)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return memPercentage
}

func stopStressProcess() {
	if stressPID != 0 {
		cmd := exec.Command("kill", "-9", strconv.Itoa(stressPID))
		err := cmd.Run()
		if err != nil {
			fmt.Println("Error stopping stress process:", err)
		}
	}
}

func increasing(cpu_rate float64, mem_rate float64) {
	stopStressProcess() // 终止之前的stress进程

	cmd := exec.Command("stress", "--cpu", "2") // 默认模拟2个CPU的高负载

	err := cmd.Start()
	if err != nil {
		fmt.Println("Error starting stress command:", err)
		return
	}

	stressPID = cmd.Process.Pid
	fmt.Printf("Started generating high CPU load with stress command for CPU rate %.2f%% and Mem rate %.2f%%\n", cpu_rate, mem_rate)

	for {
		cpu := getCPUPercentage()
		mem := getMemPercentage()

		fmt.Printf("Current CPU usage: %.2f%%\n", cpu)
		fmt.Printf("Current Memory usage: %.2f%%\n", mem)

		if cpu > cpu_rate && mem > mem_rate {
			stopStressProcess() // 取消高负载模拟
			fmt.Println("Stopped generating high CPU load")
			break
		}

		time.Sleep(5 * time.Second) // Check every 5 seconds
	}
}

func main() {
	increasing(lower_cpu, lower_mem)
}

