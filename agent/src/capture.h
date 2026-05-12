#ifndef CAPTURE_H
#define CAPTURE_H

#include <pcap.h>

#pragma pack(push, 1) // no padding added by the compiler
struct custom_arp_header {
    uint16_t hw_type;       /* Hardware Type (e.g., Ethernet = 1) */
    uint16_t proto_type;    /* Protocol Type (e.g., IPv4 = 0x0800) */
    uint8_t  hw_len;        /* Hardware Address Length (e.g., MAC = 6) */
    uint8_t  proto_len;     /* Protocol Address Length (e.g., IPv4 = 4) */
    uint16_t opcode;        /* Operation Code (Request = 1, Reply = 2) */
    uint8_t  sender_mac[6]; /* Sender Hardware Address */
    uint8_t  sender_ip[4];  /* Sender IP Address */
    uint8_t  target_mac[6]; /* Target Hardware Address */
    uint8_t  target_ip[4];  /* Target IP Address */
};
#pragma pack(pop)

pcap_t* init_capture(const char* device);
void start_capture_loop(pcap_t *handle);

#endif /* CAPTURE_H */
