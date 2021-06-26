package main

import (
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const LISTEN_ADDRESS = ":9101"
const NVIDIA_SMI_PATH = "/usr/bin/nvidia-smi"

var testMode string

type NvidiaSmiLog struct {
	DriverVersion string `xml:"driver_version"`
	AttachedGPUs  string `xml:"attached_gpus"`
	GPUs          []struct {
		ProductName  string `xml:"product_name"`
		ProductBrand string `xml:"product_brand"`
		UUID         string `xml:"uuid"`
		FanSpeed     string `xml:"fan_speed"`
		PCI          struct {
			PCIBus string `xml:"pci_bus"`
		} `xml:"pci"`
		FbMemoryUsage struct {
			Total string `xml:"total"`
			Used  string `xml:"used"`
			Free  string `xml:"free"`
		} `xml:"fb_memory_usage"`
		Utilization struct {
			GPUUtil    string `xml:"gpu_util"`
			MemoryUtil string `xml:"memory_util"`
		} `xml:"utilization"`
		Temperature struct {
			GPUTemp              string `xml:"gpu_temp"`
			GPUTempMaxThreshold  string `xml:"gpu_temp_max_threshold"`
			GPUTempSlowThreshold string `xml:"gpu_temp_slow_threshold"`
		} `xml:"temperature"`
		PowerReadings struct {
			PowerDraw  string `xml:"power_draw"`
			PowerLimit string `xml:"power_limit"`
		} `xml:"power_readings"`
		Clocks struct {
			GraphicsClock string `xml:"graphics_clock"`
			SmClock       string `xml:"sm_clock"`
			MemClock      string `xml:"mem_clock"`
			VideoClock    string `xml:"video_clock"`
		} `xml:"clocks"`
		MaxClocks struct {
			GraphicsClock string `xml:"graphics_clock"`
			SmClock       string `xml:"sm_clock"`
			MemClock      string `xml:"mem_clock"`
			VideoClock    string `xml:"video_clock"`
		} `xml:"max_clocks"`
	} `xml:"gpu"`
}

func formatValue(key string, meta string, value string) string {
	result := key
	if meta != "" {
		result += "{" + meta + "}"
	}
	return result + " " + value + "\n"
}

func filterNumber(value string) string {
	r := regexp.MustCompile("[^0-9.]")
	s := r.ReplaceAllString(value, "")
	if s != "" {
		return s
	} else {
		return "0"
	}
}

func nvidiaVersion(version string) (string, string, string) {
	var major string
	var minor string
	var patch string
	t := strings.Split(version, ".")

	if len(t) <= 1 {
		major = "0"
		minor = "0"
		patch = "0"
	} else if len(t) <= 2 {
		major = t[0]
		minor = t[1]
		patch = "0"
	} else {
		major = t[0]
		minor = t[1]
		patch = t[2]
	}
	return major, minor, patch
}

func metrics(w http.ResponseWriter, r *http.Request) {
	log.Print("Serving /metrics")

	var cmd *exec.Cmd
	if testMode == "1" {
		dir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		cmd = exec.Command("/bin/cat", dir+"/test.xml")
	} else {
		cmd = exec.Command(NVIDIA_SMI_PATH, "-q", "-x")
	}

	// Execute system command
	stdout, err := cmd.Output()
	if err != nil {
		println(err.Error())
		return
	}

	// Parse XML
	var xmlData NvidiaSmiLog
	xml.Unmarshal(stdout, &xmlData)

	// Output

	major, minor, patch := nvidiaVersion(filterNumber(xmlData.DriverVersion))
	patch_int, _ := strconv.Atoi(patch)
	patch = strconv.Itoa(patch_int)

	io.WriteString(w, formatValue("nvidiasmi_driver_major_version", "", major))
	io.WriteString(w, formatValue("nvidiasmi_driver_minor_version", "", minor))
	io.WriteString(w, formatValue("nvidiasmi_driver_patch_version", "", patch))
	io.WriteString(w, formatValue("nvidiasmi_attached_gpus", "", filterNumber(xmlData.AttachedGPUs)))
	for _, GPU := range xmlData.GPUs {
		io.WriteString(w, formatValue("nvidiasmi_fan_speed", "uuid=\""+GPU.UUID+"\"", filterNumber(GPU.FanSpeed)))
		io.WriteString(w, formatValue("nvidiasmi_memory_usage_total", "uuid=\""+GPU.UUID+"\"", filterNumber(GPU.FbMemoryUsage.Total)))
		io.WriteString(w, formatValue("nvidiasmi_memory_usage_used", "uuid=\""+GPU.UUID+"\"", filterNumber(GPU.FbMemoryUsage.Used)))
		io.WriteString(w, formatValue("nvidiasmi_memory_usage_free", "uuid=\""+GPU.UUID+"\"", filterNumber(GPU.FbMemoryUsage.Free)))
		io.WriteString(w, formatValue("nvidiasmi_utilization_gpu", "uuid=\""+GPU.UUID+"\"", filterNumber(GPU.Utilization.GPUUtil)))
		io.WriteString(w, formatValue("nvidiasmi_utilization_memory", "uuid=\""+GPU.UUID+"\"", filterNumber(GPU.Utilization.MemoryUtil)))
		io.WriteString(w, formatValue("nvidiasmi_temp_gpu", "uuid=\""+GPU.UUID+"\"", filterNumber(GPU.Temperature.GPUTemp)))
		io.WriteString(w, formatValue("nvidiasmi_temp_gpu_max", "uuid=\""+GPU.UUID+"\"", filterNumber(GPU.Temperature.GPUTempMaxThreshold)))
		io.WriteString(w, formatValue("nvidiasmi_temp_gpu_slow", "uuid=\""+GPU.UUID+"\"", filterNumber(GPU.Temperature.GPUTempSlowThreshold)))
		io.WriteString(w, formatValue("nvidiasmi_power_draw", "uuid=\""+GPU.UUID+"\"", filterNumber(GPU.PowerReadings.PowerDraw)))
		io.WriteString(w, formatValue("nvidiasmi_power_limit", "uuid=\""+GPU.UUID+"\"", filterNumber(GPU.PowerReadings.PowerLimit)))
		io.WriteString(w, formatValue("nvidiasmi_clock_graphics", "uuid=\""+GPU.UUID+"\"", filterNumber(GPU.Clocks.GraphicsClock)))
		io.WriteString(w, formatValue("nvidiasmi_clock_graphics_max", "uuid=\""+GPU.UUID+"\"", filterNumber(GPU.MaxClocks.GraphicsClock)))
		io.WriteString(w, formatValue("nvidiasmi_clock_sm", "uuid=\""+GPU.UUID+"\"", filterNumber(GPU.Clocks.SmClock)))
		io.WriteString(w, formatValue("nvidiasmi_clock_sm_max", "uuid=\""+GPU.UUID+"\"", filterNumber(GPU.MaxClocks.SmClock)))
		io.WriteString(w, formatValue("nvidiasmi_clock_mem", "uuid=\""+GPU.UUID+"\"", filterNumber(GPU.Clocks.MemClock)))
		io.WriteString(w, formatValue("nvidiasmi_clock_mem_max", "uuid=\""+GPU.UUID+"\"", filterNumber(GPU.MaxClocks.MemClock)))
		io.WriteString(w, formatValue("nvidiasmi_clock_video", "uuid=\""+GPU.UUID+"\"", filterNumber(GPU.Clocks.VideoClock)))
		io.WriteString(w, formatValue("nvidiasmi_clock_video_max", "uuid=\""+GPU.UUID+"\"", filterNumber(GPU.MaxClocks.VideoClock)))
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	log.Print("Serving /index")
	html := `<!doctype html>
<html>
    <head>
        <meta charset="utf-8">
        <title>Nvidia SMI Exporter modified by Junyi</title>
    </head>
    <body>
        <h1>Nvidia SMI Exporter modified by Junyi</h1>
        <p><a href="/metrics">Metrics</a></p>
    </body>
</html>`
	io.WriteString(w, html)
}

func main() {
	testMode = os.Getenv("TEST_MODE")
	if testMode == "1" {
		log.Print("Test mode is enabled")
	}

	log.Print("Nvidia SMI exporter listening on " + LISTEN_ADDRESS)
	http.HandleFunc("/", index)
	http.HandleFunc("/metrics", metrics)
	http.ListenAndServe(LISTEN_ADDRESS, nil)
}
