-- Seed sample ARP events into the arp_events table
-- Includes:
-- 1. IP Conflict: Two agents report different MACs for the same IP (192.168.1.100)
-- 2. Segment Conflict: Two different agents report the same MAC address (BB:BB:BB:BB:BB:BB)

INSERT INTO arp_events (agent_id, captured_at, opcode, target_ip, target_mac, sender_ip, sender_mac) VALUES
-- Normal traffic (Agent 1)
('agent-1', NOW() - INTERVAL '12 minutes', 1, '192.168.1.1', '00:00:00:00:00:00', '192.168.1.50', 'AA:BB:CC:DD:EE:01'),
('agent-1', NOW() - INTERVAL '11 minutes', 2, '192.168.1.50', 'AA:BB:CC:DD:EE:01', '192.168.1.1', '00:11:22:33:44:55'),

-- Normal traffic (Agent 2)
('agent-2', NOW() - INTERVAL '10 minutes', 1, '192.168.1.1', '00:00:00:00:00:00', '192.168.1.70', '11:22:33:AA:BB:CC'),
('agent-2', NOW() - INTERVAL '9 minutes', 2, '192.168.1.70', '11:22:33:AA:BB:CC', '192.168.1.1', '00:11:22:33:44:55'),

-- IP CONFLICT: Two different agents report the same IP (192.168.1.100) but different MACs
-- (Classic ARP Poisoning / Hijacking attempt)
('agent-1', NOW() - INTERVAL '8 minutes', 2, '192.168.1.1', '00:11:22:33:44:55', '192.168.1.100', 'DE:AD:BE:EF:00:01'),
('agent-2', NOW() - INTERVAL '7 minutes', 2, '192.168.1.1', '00:11:22:33:44:55', '192.168.1.100', 'CA:FE:BA:BE:00:02'),

-- SEGMENT CONFLICT: Two different agents report the SAME MAC (BB:BB:BB:BB:BB:BB)
-- This detects if a device is physically visible in two different segments (Lateral movement or MAC Flapping)
('agent-1', NOW() - INTERVAL '6 minutes', 1, '192.168.1.1', '00:00:00:00:00:00', '192.168.1.200', 'BB:BB:BB:BB:BB:BB'),
('agent-2', NOW() - INTERVAL '5 minutes', 1, '192.168.1.1', '00:00:00:00:00:00', '192.168.1.200', 'BB:BB:BB:BB:BB:BB'),

-- More traffic
('agent-1', NOW() - INTERVAL '4 minutes', 1, '192.168.1.1', '00:00:00:00:00:00', '192.168.1.80', 'AA:BB:CC:DD:EE:03'),
('agent-2', NOW() - INTERVAL '3 minutes', 1, '192.168.1.1', '00:00:00:00:00:00', '192.168.1.90', '11:22:33:AA:BB:DD'),
('agent-1', NOW() - INTERVAL '2 minute', 2, '192.168.1.80', 'AA:BB:CC:DD:EE:03', '192.168.1.1', '00:11:22:33:44:55'),
('agent-2', NOW() - INTERVAL '1 minute', 2, '192.168.1.90', '11:22:33:AA:BB:DD', '192.168.1.1', '00:11:22:33:44:55');
