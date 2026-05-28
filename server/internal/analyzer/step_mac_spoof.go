package analyzer

import (
	"context"
	"fmt"
	"net"
	"server/internal/database"
	"time"
)

type MACSpoofDetectorStep struct {
	db database.Database
}

func (msd *MACSpoofDetectorStep) Process(ctx context.Context, report *AnalysisReport) error {
	senderMAC := net.HardwareAddr(report.Event.SenderMac).String()

	agents, err := msd.db.GetAgentsForMAC(ctx, senderMAC, 10*time.Minute)
	if err != nil {
		return err
	}

	if len(agents) > 1 {
		agentList := ""
		for i, a := range agents {
			if i > 0 {
				agentList += ", "
			}
			agentList += a
		}

		report.AddFinding(fmt.Sprintf(
			"Segment Conflict (MAC Spoofing/Flapping): MAC [%s] is being reported simultaneously by multiple agents: [%s]",
			senderMAC, agentList,
		))
	}

	return nil
}
