#include "config.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <ctype.h>

app_config_t global_config = {
    .agent_id = "agent-default",
    .server_address = "localhost:50051",
    .interface = "any"
};

static void trim(char *str) {
    if (!str || *str == '\0') return;
    char *start = str;
    char *end = str + strlen(str) - 1;
    while(isspace((unsigned char)*start)) start++;
    while(end > start && isspace((unsigned char)*end)) end--;
    end[1] = '\0';
    memmove(str, start, end - start + 2);
}

int load_config(const char *filename) {
    FILE *file = fopen(filename, "r");
    if (!file) {
        return -1;
    }

    char line[256];
    while (fgets(line, sizeof(line), file)) {
        trim(line);
        // Skip comments and empty lines
        if (line[0] == '#' || line[0] == ';' || line[0] == '\0' || line[0] == '[') {
            continue;
        }

        char *delim = strchr(line, '=');
        if (delim) {
            *delim = '\0';
            char *key = line;
            char *value = delim + 1;
            trim(key);
            trim(value);

            if (strcmp(key, "agent_id") == 0) {
                strncpy(global_config.agent_id, value, sizeof(global_config.agent_id) - 1);
                global_config.agent_id[sizeof(global_config.agent_id) - 1] = '\0';
            } else if (strcmp(key, "server_address") == 0) {
                strncpy(global_config.server_address, value, sizeof(global_config.server_address) - 1);
                global_config.server_address[sizeof(global_config.server_address) - 1] = '\0';
            } else if (strcmp(key, "interface") == 0) {
                strncpy(global_config.interface, value, sizeof(global_config.interface) - 1);
                global_config.interface[sizeof(global_config.interface) - 1] = '\0';
            }
        }
    }
    fclose(file);
    return 0;
}