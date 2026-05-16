#ifndef SIGNALS_H
#define SIGNALS_H

#include <signal.h>
#include <pcap.h>

extern volatile sig_atomic_t keep_running;

void setup_signal_handlers(void);
void set_pcap_handle(pcap_t *handle);

#endif /* SIGNALS_H */
