package analyzer

import (
	"context"
	"fmt"
	"log"
	"net"
	"server/internal/database"
	"time"
)

const SEEN_EXPIRATION_TS = 180 * 24 * time.Hour // 6 months

type MACChangeDetectorStep struct {
	db database.Database
}

func (macdet *MACChangeDetectorStep) Process(ctx context.Context, report *AnalysisReport) error {
	senderIPnet := net.IP(report.Event.SenderId)

	if senderIPnet.IsUnspecified() { // 0.0.0.0
		return nil
	}

	currentBinding, err := macdet.db.GetIPMACBinding(ctx, senderIPnet.String())
	if err != nil {
		return err
	}

	senderMAC := net.HardwareAddr(report.Event.SenderMac).String()

	if currentBinding == nil {
		err := macdet.db.CreateIPMACBinding(ctx, senderIPnet.String(), senderMAC, report.Event.Timestamp)
		if err != nil {
			return err
		}

		log.Printf("New MAC Binding %s -> %s", senderIPnet.String(), senderMAC)
		return nil
	}

	if senderMAC == currentBinding.MACAddress {
		err := macdet.db.UpdateLastSeen(ctx, senderIPnet.String(), report.Event.Timestamp)
		if err != nil {
			return err
		}

		return nil
	}

	if time.Since(currentBinding.LastSeen) > SEEN_EXPIRATION_TS {
		log.Printf("Legitimate MAC change detected for IP %s (Last seen: %v). Updating state.",
			senderIPnet.String(), currentBinding.LastSeen)

		err := macdet.db.UpdateMAC(ctx, senderIPnet.String(), senderMAC, report.Event.Timestamp)
		if err != nil {
			return err
		}
		return nil
	}

	msg := fmt.Sprintf("IP %s was previously at [%s], but now reports at [%s]. Potential ARP Poisoning detected.",
		senderIPnet.String(), currentBinding.MACAddress, senderMAC)
	report.RegisterAttack(AttackARPPoisoning, msg)

	err = macdet.db.UpdateStatus(ctx, senderIPnet.String(), database.BIND_CONFLICT)
	if err != nil {
		return err
	}

	return nil
}
