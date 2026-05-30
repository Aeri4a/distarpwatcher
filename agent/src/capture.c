#include "capture.h"
#include "signals.h"
#include "grpc_client.h"
#include "config.h"
#include "log.h"
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
        LOG_ERR("Couldn't open device %s: %s", device, errbuf);
        return nullptr;
    }

    if (pcap_compile(handle, &fp, filter_exp, 0, net) == -1) {
        LOG_ERR("Couldn't parse filter %s: %s", filter_exp, pcap_geterr(handle));
        return nullptr;
    }

    if (pcap_setfilter(handle, &fp) == -1) {
        LOG_ERR("Couldn't install filter %s: %s", filter_exp, pcap_geterr(handle));
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
        LOG_WARN("Unsupported datalink type %d, assuming Ethernet.", link_type);
        header_offset = 14;
    }

    LOG_INFO("Starting capture loop...");
    while (keep_running) {
        res = pcap_next_ex(handle, &header, &packet);
        if (res == 0) continue; /* Timeout */
        if (res == -1) {
            LOG_ERR("Error reading the packets: %s", pcap_geterr(handle));
            break;
        }
        if (res == -2) break; /* pcap_breakloop */

        if (header->caplen < header_offset + sizeof(struct ether_arp)) {
            continue; /* Packet too short */
        }

        arp = (struct ether_arp*)(packet + header_offset);
        
        LOG_DEBUG("ARP [Opcode: %d] Sender: %d.%d.%d.%d (%02X:%02X:%02X:%02X:%02X:%02X) -> Target: %d.%d.%d.%d", 
               ntohs(arp->ea_hdr.ar_op),
               arp->arp_spa[0], arp->arp_spa[1], arp->arp_spa[2], arp->arp_spa[3],
               arp->arp_sha[0], arp->arp_sha[1], arp->arp_sha[2], arp->arp_sha[3], arp->arp_sha[4], arp->arp_sha[5],
               arp->arp_tpa[0], arp->arp_tpa[1], arp->arp_tpa[2], arp->arp_tpa[3]);

        send_arp_event(
            global_config.agent_id,
            ntohs(arp->ea_hdr.ar_op),
            arp->arp_tpa,
            arp->arp_tha,
            arp->arp_spa,
            arp->arp_sha
        );
    }
    LOG_INFO("Capture loop stopped.");
}
