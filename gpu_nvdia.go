package gpu_nvidia

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//### Following the formating for telegraf plugins

func (_ *gpu_nvidia) Description() string {
	return "nvidia-smi return statistics"
}

const sampleConfig = `
  ## number gpu for stats
  # gpu = 1
  ## binary location for nvidia-smi
  # binPath = /usr/bin/nvidia-smi
`

func (_ *gpu_nvidia) SampleConfig() string {
	return sampleConfig
}

type gpu_nvidia struct {
	//"nvidia-smi full path"
	binPath string //`"/usr/bin/nvidia-smi"`
	//"display some things"
	verbose bool //`false`
	//select GPU to query , default 99 for all
	gpuId int //`99`
}

func getResult(bin string, metric string, verbose bool, gpuId int) string {
	query := fmt.Sprintf("--query-gpu=%s", metric)
	gpu := " "
	if gpuId != 99 {
		// if its not specified , take them all
		gpu = fmt.Sprintf("--id=%d", gpuId)
	}
	opts := []string{"--format=noheader,nounits,csv", query, gpu}

	if verbose {
		log.Print("Going to run ")
		log.Print(bin)
		log.Println(" with ")
		log.Println(opts)
	}
	ret, err := exec.Command(bin, opts...).CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s", err, ret)
		return ""
	}
	return string(ret)
}

//func main() {
func (g *gpu_nvidia) Gather(acc telegraf.Accumulator) error {
	if _, err := os.Stat(g.binPath); os.IsNotExist(err) {
		//fmt.Fprintf(os.Stderr, "Bin path does not exists: %s", g.binPath)
		acc.AddError(err)
		//continue
		return nil // exit
	}

	metrics := "fan.speed,memory.total,memory.used,memory.free,pstate,temperature.gpu,name,uuid,compute_mode,utilization.gpu,utilization.memory,power.draw"
	//nvidia-smi --format=csv  --query-gpu=fan.speed,memory.total,memory.used,memory.free,pstate,temperature.gpu,name,uuid,compute_mode,utilization.gpu,utilization.memory,power.draw

	results := getResult(g.binPath, metrics, g.verbose, g.gpuId)

	if results == "" {
		return nil // exit
	}

	for _, m := range strings.Split(results, "\n") {
		if m != "" {

			splitResults := strings.Split(m, ",")

			// log.Printf("nvidiasmi,uuid=%s ", strings.TrimSpace(splitResults[7])) // it should be available ... if no, you have some problems
			//
			// log.Printf("gpu_name=\"%s\",", strings.TrimSpace(splitResults[6]))
			// log.Printf("gpu_compute_mode=\"%s\",", strings.TrimSpace(splitResults[8]))
			//
			// log.Printf("fan_speed=%s,", strings.TrimSpace(splitResults[0])) // it's a % 0-100
			//
			// log.Printf("memory_total=%s,", strings.TrimSpace(splitResults[1])) // they
			// log.Printf("memory_used=%s,", strings.TrimSpace(splitResults[2]))  // are
			// log.Printf("memory_free=%s,", strings.TrimSpace(splitResults[3]))  // MiB
			//
			// log.Printf("pstate=%s,", strings.TrimSpace(strings.Replace(splitResults[4], "P", "", -1))) // strip the P
			// log.Printf("temperature=%s,", strings.TrimSpace(splitResults[5]))                          // in degrees Celcius
			// log.Printf("utilization.gpu=%s,", strings.TrimSpace(splitResults[9]))                      // in Percentage
			// log.Printf("utilization.memory=%s,", strings.TrimSpace(splitResults[10]))                  // in Percentage
			// log.Printf("power.draw=%s", strings.TrimSpace(splitResults[11]))                           // in Watt
			// log.Printf("\n")
			//[sachin_rase@x ~]$  /usr/bin/nvidia-smi --format=noheader,nounits,csv --query-gpu=fan.speed,memory.total,memory.used,memory.free,pstate,temperature.gpu,name,uuid,compute_mode,utilization.gpu,utilization.memory,power.draw
			//fan.speed [%], memory.total [MiB], memory.used [MiB], memory.free [MiB], pstate, temperature.gpu, name, uuid, compute_mode, utilization.gpu [%], utilization.memory [%], power.draw [W]
			// [Not Supported], 16308, 0, 16308, P0, 31, Tesla P100-PCIE-16GB, GPU-4128918d-a887-1670-9668-7128be62a911, Default, 0, 0, 24.77
			// [Not Supported], 16308, 0, 16308, P0, 27, Tesla P100-PCIE-16GB, GPU-70900a83-d2eb-ae66-5f18-b1c6b038fe47, Default, 0, 0, 24.56
			// [Not Supported], 16308, 0, 16308, P0, 31, Tesla P100-PCIE-16GB, GPU-31157ca4-d8c9-afd5-d673-aa430310ba9b, Default, 0, 0, 24.54
			// [Not Supported], 16308, 0, 16308, P0, 30, Tesla P100-PCIE-16GB, GPU-28395af2-6be9-5221-8f05-3010cd1371ac, Default, 0, 0, 25.73

			uuid := strings.TrimSpace(splitResults[7]) // it should be available ... if no, you have some problems
			gpu_name := strings.TrimSpace(splitResults[6])
			gpu_compute_mode := strings.TrimSpace(splitResults[8])
			fan_speed := strings.TrimSpace(splitResults[0])                            // it's a % 0-100
			memory_total := strings.TrimSpace(splitResults[1])                         // they
			memory_used := strings.TrimSpace(splitResults[2])                          // are
			memory_free := strings.TrimSpace(splitResults[3])                          // MiB
			pstate := strings.TrimSpace(strings.Replace(splitResults[4], "P", "", -1)) // strip the P
			temperature := strings.TrimSpace(splitResults[5])                          // in degrees Celcius
			utilization_gpu := strings.TrimSpace(splitResults[9])                      // in Percentage
			utilization_memory := strings.TrimSpace(splitResults[10])                  // in Percentage
			power_draw := strings.TrimSpace(splitResults[11])

			tags := map[string]string{
				"gpu_name":     gpu_name,
				"compute_mode": gpu_compute_mode,
				"uuid":         uuid,
			}
			fields := map[string]interface{}{
				"fan.speed":          fan_speed,
				"memory.total":       memory_total,
				"memory.used":        memory_used,
				"memory.free":        memory_free,
				"pstate":             pstate,
				"temperature.gpu":    temperature,
				"utilization.gpu":    utilization_gpu,
				"utilization.memory": utilization_memory,
				"power.draw":         power_draw,
			}
			acc.AddFields("nvidia_smi", fields, tags)
		}
	}
	return nil
}

func init() {
	inputs.Add("gpu_nvidia", func() telegraf.Input {
		return &gpu_nvidia{
			binPath: "/usr/bin/nvidia-smi",
			verbose: false,
			gpuId:   99,
		}
	})
}
