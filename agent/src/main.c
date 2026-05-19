#include "signals.h"
#include "capture.h"
#include "grpc_client.h"
#include <stdio.h>

int main(int argc, char *argv[]) {
    const char *device = "any";
    if (argc > 1) {
        device = argv[1];
    }

    printf("Initializing ARP capture on interface: %s\n", device);
    
    init_grpc_client("localhost:50051");

    setup_signal_handlers();

    pcap_t *handle = init_capture(device);
    if (!handle) {
        destroy_grpc_client();
        return 1;
    }

    set_pcap_handle(handle);
    start_capture_loop(handle);

    printf("Cleaning up...\n");
    pcap_close(handle);
    destroy_grpc_client();
    printf("Exiting gracefully.\n");

    return 0;
}
