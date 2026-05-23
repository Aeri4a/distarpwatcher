CREATE TABLE IF NOT EXISTS ip_mac_bindings (
    ip_address INET PRIMARY KEY,
    mac_address MACADDR NOT NULL,
    last_seen TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'TRUSTED',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ip_mac_bindings_mac_address ON ip_mac_bindings (mac_address);
CREATE INDEX IF NOT EXISTS idx_ip_mac_bindings_status ON ip_mac_bindings (status);
