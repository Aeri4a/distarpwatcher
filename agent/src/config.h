#ifndef CONFIG_H
#define CONFIG_H

typedef struct {
    char agent_id[64];
    char server_address[128];
    char interface[64];
} app_config_t;

extern app_config_t global_config;

// Load config from file, overriding defaults. Returns 0 on success, -1 on failure.
int load_config(const char *filename);

#endif // CONFIG_H