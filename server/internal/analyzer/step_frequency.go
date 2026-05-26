package analyzer

import (
	"context"
	"fmt"
	"net"
)

type FrequencyDetectorStep struct {
	history map[string][]uint64

	threshold int
	window    uint64
}

func (freqdet *FrequencyDetectorStep) Process(ctx context.Context, report *AnalysisReport) error {
	senderMac := net.HardwareAddr(report.Event.SenderMac).String()
	cutOff := report.Event.Timestamp - freqdet.window

	cleanOldHistory(freqdet.history, senderMac, cutOff)

	freqdet.history[senderMac] = append(freqdet.history[senderMac], report.Event.Timestamp)

	if len(freqdet.history[senderMac]) > freqdet.threshold {
		report.AddFinding(fmt.Sprintf("ARP Flood detected from MAC [%s]: %d packets in the last %dms",
			senderMac, len(freqdet.history[senderMac]), freqdet.window))
	}

	return nil
}

func cleanOldHistory(history map[string][]uint64, mac string, cutoff uint64) {
	timestamps := history[mac]

	keepIndex := 0
	for i, ts := range timestamps {
		if ts >= cutoff {
			keepIndex = i
			break
		}
	}

	if len(timestamps) > 0 && timestamps[len(timestamps)-1] < cutoff {
		history[mac] = nil
		return
	}

	history[mac] = timestamps[keepIndex:]
}
