package scan

import (
	"fmt"
	"github.com/marco-lancini/goscan/core/model"
	"github.com/marco-lancini/goscan/core/utils"
)

// ---------------------------------------------------------------------------------------
// DISPATCHER
// ---------------------------------------------------------------------------------------
func ScanSweep(kind string, target string) {
	// Dispatch scan
	switch kind {
	case "PING":
		utils.Config.Log.LogInfo("Starting Ping Sweep")
		folder, file, nmapArgs := "sweep", "ping", utils.Const_NMAP_SWEEP
		execSweep(file, target, folder, file, nmapArgs)

	default:
		utils.Config.Log.LogError("Invalid type of scan")
		return
	}
}

func execSweep(name, target, folder, file, nmapArgs string) {
	targets := model.GetAllTargets(utils.Config.DB)
	for _, h := range targets {
		// Scan only if:
		//   - target is ALL
		//   - or if host is the selected one
		if target == "ALL" || target == h.Address {
			temp := h
			file = fmt.Sprintf("%s_%s", file, h.Address)
			go workerSweep(name, &temp, folder, file, nmapArgs)
		}
	}
}

// ---------------------------------------------------------------------------------------
// WORKER
// ---------------------------------------------------------------------------------------
func workerSweep(name string, h *model.Target, folder string, file string, nmapArgs string) {
	// Instantiate new NmapScan
	s := NewScan(name, h.Address, folder, file, nmapArgs)
	ScansList = append(ScansList, s)

	// Run the scan
	s.RunNmap()

	// Parse nmap's output
	res := s.ParseOutput()
	if res != nil {
		// Identify live hosts
		for _, host := range res.Hosts {
			status := host.Status.State
			if status == "up" {
				// Save as host
				addr := host.Addresses[0].Addr
				model.AddHost(utils.Config.DB, addr, "up", "new")
			}
		}
	}

	// Update status of target
	model.Mutex.Lock()
	h.Step = "sweeped"
	utils.Config.DB.Save(&h)
	model.Mutex.Unlock()

	// utils.Config.Log.LogInfo("Ping sweep completed!")
}
