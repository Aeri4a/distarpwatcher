#include "signals.h"
#include <stdio.h>
#include <string.h>

volatile sig_atomic_t keep_running = 1;
static pcap_t *global_pcap_handle = NULL;

static void handle_signal(int sig) {
    (void)sig;
    keep_running = 0;
    if (global_pcap_handle != NULL) {
        pcap_breakloop(global_pcap_handle);
    }
}

void setup_signal_handlers(void) {
    struct sigaction sa;
    memset(&sa, 0, sizeof(sa));
    sa.sa_handler = handle_signal;
    
    sigaction(SIGINT, &sa, NULL);
    sigaction(SIGTERM, &sa, NULL);
}

void set_pcap_handle(pcap_t *handle) {
    global_pcap_handle = handle;
}
