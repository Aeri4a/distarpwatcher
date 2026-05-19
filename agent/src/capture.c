#include "capture.h"
#include "signals.h"
#include "grpc_client.h"
#include "config.h"
#include <stdio.h>
#include <stdlib.h>
#include <arpa/inet.h>
#include <netinet/if_ether.h>

pcap_t* init_capture(const char* device) {
    char errbuf[PCAP_ERRBUF_SIZE];
    pcap_t *handle;
    struct bpf_program fp;
    char filter_exp[] = "arp";
    bpf_u_int32 net = PCAP_NETMASK_UNKNOWN;

    handle = pcap_open_live(device, BUFSIZ, 1, 1000, errbuf);
    if (handle == NULL) {
        fprintf(stderr, "Couldn't open device %s: %s\n", device, errbuf);
        return nullptr;
    }

    if (pcap_compile(handle, &fp, filter_exp, 0, net) == -1) {
        fprintf(stderr, "Couldn't parse filter %s: %s\n", filter_exp, pcap_geterr(handle));
        return nullptr;
    }

    if (pcap_setfilter(handle, &fp) == -1) {
        fprintf(stderr, "Couldn't install filter %s: %s\n", filter_exp, pcap_geterr(handle));
        pcap_freecode(&fp);
        return nullptr;
    }

    pcap_freecode(&fp);
    return handle;
}

void start_capture_loop(pcap_t *handle) {
    struct pcap_pkthdr *header;
    const u_char *packet;
    const struct ether_arp *arp;
    int res;
    int link_type = pcap_datalink(handle);
    int header_offset = 0;

    if (link_type == DLT_EN10MB) {
        header_offset = 14; /* Ethernet header size */
    } else if (link_type == DLT_LINUX_SLL) {
        header_offset = 16; /* Linux Cooked Capture (SLL) header size */
    } else {
        fprintf(stderr, "Warning: Unsupported datalink type %d, assuming Ethernet.\n", link_type);
        header_offset = 14;
    }

    printf("Starting capture loop...\n");
    while (keep_running) {
        res = pcap_next_ex(handle, &header, &packet);
        if (res == 0) continue; /* Timeout */
        if (res == -1) {
            fprintf(stderr, "Error reading the packets: %s\n", pcap_geterr(handle));
            break;
        }
        if (res == -2) break; /* pcap_breakloop */

        if (header->caplen < header_offset + sizeof(struct ether_arp)) {
            continue; /* Packet too short */
        }

        arp = (struct ether_arp*)(packet + header_offset);
        
        printf("Captured ARP packet: length %d\n", header->len);
        printf("  Target MAC: %02X:%02X:%02X:%02X:%02X:%02X\n",
           arp->arp_tha[0], arp->arp_tha[1], arp->arp_tha[2],
           arp->arp_tha[3], arp->arp_tha[4], arp->arp_tha[5]);
        printf("  Target IP:  %d.%d.%d.%d\n",
               arp->arp_tpa[0], arp->arp_tpa[1], arp->arp_tpa[2], arp->arp_tpa[3]);
        printf("==========================================\n\n");

        send_arp_event(
            global_config.agent_id,
            ntohs(arp->ea_hdr.ar_op),
            arp->arp_tpa,
            arp->arp_tha,
            arp->arp_spa,
            arp->arp_sha
        );
    }
    printf("Capture loop stopped.\n");
}
