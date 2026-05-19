#include "signals.h"
#include "capture.h"
#include "grpc_client.h"
#include "config.h"
#include <stdio.h>

int main(int argc, char *argv[]) {
    if (argc > 1) {
        if (load_config(argv[1]) != 0) {
            fprintf(stderr, "Error: Could not load specified config file '%s'\n", argv[1]);
            return 1;
        }
    } else {
        if (load_config("/etc/distarpwatcher/agent.conf") != 0) {
            if (load_config("agent.conf") != 0) { // fallback to local
                printf("Warning: No config file found. Using default settings.\n");
            }
        }
    }

    printf("Starting ARP Watcher Agent\n");
    printf("  Agent ID: %s\n", global_config.agent_id);
    printf("  Server Address: %s\n", global_config.server_address);
    printf("  Capture Interface: %s\n", global_config.interface);

    init_grpc_client(global_config.server_address);

    setup_signal_handlers();

    pcap_t *handle = init_capture(global_config.interface);
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
