#ifndef CAPTURE_H
#define CAPTURE_H

#include <pcap.h>

pcap_t* init_capture(const char* device);
void start_capture_loop(pcap_t *handle);

#endif /* CAPTURE_H */
