#!/usr/bin/env python3
"""
ARP Flood Simulator
Requires Scapy: pip install scapy
Run with root privileges (sudo)

WARNING: ONLY RUN THIS ON NETWORKS YOU OWN OR HAVE PERMISSION TO TEST ON.
"""

import argparse
import time
from scapy.all import ARP, Ether, sendp, get_if_hwaddr, conf

def simulate_flood(interface, count=100, delay=0.01):
    """
    Sends a flood of ARP packets to trigger frequency detection rules.
    """
    print(f"[*] Starting ARP flood simulation on interface: {interface}")
    
    try:
        source_mac = get_if_hwaddr(interface)
    except Exception as e:
        print(f"[!] Could not determine MAC for {interface}: {e}")
        return

    print(f"[*] Goal: Send {count} ARP packets rapidly from MAC {source_mac}")

    ether = Ether(dst="ff:ff:ff:ff:ff:ff", src=source_mac)
    arp = ARP(op=1, pdst="192.168.1.254", hwdst="00:00:00:00:00:00", psrc="192.168.1.100", hwsrc=source_mac)
    
    packet = ether / arp

    print(f"[*] Packet ready. Sending {count} packets with {delay}s delay between each...\n")

    start_time = time.time()
    for i in range(count):
        sendp(packet, iface=interface, verbose=False)
        if (i + 1) % 20 == 0:
            print(f"[+] Sent {i+1}/{count} packets...")
        time.sleep(delay)
            
    end_time = time.time()
    duration = end_time - start_time
    print(f"\n[*] Flood complete. Sent {count} packets in {duration:.2f} seconds.")

def main():
    parser = argparse.ArgumentParser(description="ARP Flood Simulator")
    parser.add_argument("-i", "--interface", required=True, help="Network interface to send on (e.g., eth0, wlo1)")
    parser.add_argument("-c", "--count", type=int, default=100, help="Number of packets to send (default: 100)")
    parser.add_argument("-d", "--delay", type=float, default=0.01, help="Delay between packets in seconds (default: 0.01)")
    
    args = parser.parse_args()
    
    simulate_flood(
        interface=args.interface,
        count=args.count,
        delay=args.delay
    )

if __name__ == "__main__":
    main()
