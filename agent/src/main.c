#include "signals.h"
#include "capture.h"
#include <stdio.h>

int main(int argc, char *argv[]) {
    const char *device = "any";
    if (argc > 1) {
        device = argv[1];
    }

    printf("Initializing ARP capture on interface: %s\n", device);
    
    setup_signal_handlers();

    pcap_t *handle = init_capture(device);
    if (!handle) {
        return 1;
    }

    start_capture_loop(handle);

    printf("Cleaning up...\n");
    pcap_close(handle);
    printf("Exiting gracefully.\n");

    return 0;
}
