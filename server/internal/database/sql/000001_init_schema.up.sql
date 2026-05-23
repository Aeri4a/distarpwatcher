CREATE TABLE IF NOT EXISTS arp_events (
    id SERIAL PRIMARY KEY,
    agent_id VARCHAR(128) NOT NULL,
    captured_at TIMESTAMP WITH TIME ZONE NOT NULL,
    opcode INTEGER NOT NULL,
    target_ip INET,
    target_mac MACADDR,
    sender_ip INET,
    sender_mac MACADDR,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_arp_events_agent_id ON arp_events (agent_id);
CREATE INDEX IF NOT EXISTS idx_arp_events_captured_at ON arp_events (captured_at);
CREATE INDEX IF NOT EXISTS idx_arp_events_sender_ip ON arp_events (sender_ip);
