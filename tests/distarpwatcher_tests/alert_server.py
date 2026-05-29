#!/usr/bin/env python3
"""
Simple Mock Webhook Receiver
Listens for POST requests and prints the JSON payload.
"""

import json
from http.server import HTTPServer, BaseHTTPRequestHandler
import argparse

class WebhookHandler(BaseHTTPRequestHandler):
    def do_POST(self):
        content_length = int(self.headers['Content-Length'])
        post_data = self.rfile.read(content_length)
        
        try:
            payload = json.loads(post_data.decode('utf-8'))
            print("\n" + "="*40)
            print("Received webhook alert")
            print(f"Agent ID: {payload.get('agent_id', 'N/A')}")
            print(f"Message:  {payload.get('message', 'N/A')}")
            print("="*40 + "\n")
            
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps({"status": "received"}).encode('utf-8'))
        except Exception as e:
            print(f"[!] Error parsing payload: {e}")
            self.send_response(400)
            self.end_headers()

def main():
    parser = argparse.ArgumentParser(description="Mock Webhook Server")
    parser.add_argument("-p", "--port", type=int, default=9000, help="Port to listen on (default: 9000)")
    args = parser.parse_args()

    server_address = ('', args.port)
    httpd = HTTPServer(server_address, WebhookHandler)
    print(f"[*] Mock Webhook Server listening on port {args.port}...")
    try:
        httpd.serve_forever()
    except KeyboardInterrupt:
        print("\n[*] Shutting down server.")
        httpd.server_close()

if __name__ == "__main__":
    main()
