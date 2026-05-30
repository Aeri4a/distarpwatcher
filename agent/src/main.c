#include "signals.h"
#include "capture.h"
#include "grpc_client.h"
#include "config.h"
#include "log.h"
#include <stdio.h>

int main(int argc, char *argv[]) {
    if (argc > 1) {
        if (load_config(argv[1]) != 0) {
            LOG_ERR("Could not load specified config file '%s'", argv[1]);
            return 1;
        }
    } else {
        if (load_config("/etc/distarpwatcher/agent.conf") != 0) {
            if (load_config("agent.conf") != 0) { // fallback to local
                LOG_WARN("No config file found. Using default settings.");
            }
        }
    }

    LOG_INFO("Starting ARP Watcher Agent");
    LOG_INFO("Agent ID: %s", global_config.agent_id);
    LOG_INFO("Server Address: %s", global_config.server_address);
    LOG_INFO("Capture Interface: %s", global_config.interface);

    init_grpc_client(global_config.server_address);

    setup_signal_handlers();

    pcap_t *handle = init_capture(global_config.interface);
    if (!handle) {
        destroy_grpc_client();
        return 1;
    }

    set_pcap_handle(handle);
    start_capture_loop(handle);

    LOG_INFO("Cleaning up...");
    pcap_close(handle);
    destroy_grpc_client();
    LOG_INFO("Exiting gracefully.");

    return 0;
}
