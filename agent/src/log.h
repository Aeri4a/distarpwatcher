#ifndef LOG_H
#define LOG_H

#include <stdio.h>

// Systemd/Journald
#define SD_EMERG   "<0>"
#define SD_ALERT   "<1>"
#define SD_CRIT    "<2>"
#define SD_ERR     "<3>"
#define SD_WARNING "<4>"
#define SD_NOTICE  "<5>"
#define SD_INFO    "<6>"
#define SD_DEBUG   "<7>"

#define LOG_INFO(fmt, ...)    fprintf(stdout, SD_INFO "[INFO] " fmt "\n", ##__VA_ARGS__)
#define LOG_WARN(fmt, ...)    fprintf(stderr, SD_WARNING "[WARN] " fmt "\n", ##__VA_ARGS__)
#define LOG_ERR(fmt, ...)     fprintf(stderr, SD_ERR "[ERROR] " fmt "\n", ##__VA_ARGS__)
#define LOG_DEBUG(fmt, ...)   fprintf(stdout, SD_DEBUG "[DEBUG] " fmt "\n", ##__VA_ARGS__)

#endif // LOG_H
