#ifndef GRPC_CLIENT_H
#define GRPC_CLIENT_H

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

void init_grpc_client(const char* target);

void send_arp_event(
    const char* agent_id,
    uint32_t opcode,
    const uint8_t* target_ip,
    const uint8_t* target_mac,
    const uint8_t* sender_ip,
    const uint8_t* sender_mac
);

void destroy_grpc_client();

#ifdef __cplusplus
}
#endif

#endif // GRPC_CLIENT_H
