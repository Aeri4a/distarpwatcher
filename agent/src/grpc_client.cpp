#include "grpc_client.h"
#include <iostream>
#include <memory>
#include <string>
#include <chrono>

#include <grpcpp/grpcpp.h>
#include "pb/collector.grpc.pb.h"

using grpc::Channel;
using grpc::ClientContext;
using grpc::ClientWriter;
using grpc::Status;
using distarpwatcher::ARPCollector;
using distarpwatcher::ARPEvent;
using distarpwatcher::ARPEventResponse;

class ARPClient {
public:
    ARPClient(std::shared_ptr<Channel> channel)
        : stub_(ARPCollector::NewStub(channel)) {
        stream_ = stub_->ARPStream(&context_, &response_);
    }

    ~ARPClient() {
        if (stream_) {
            stream_->WritesDone();
            Status status = stream_->Finish();
            if (!status.ok()) {
                std::cerr << "gRPC Stream finished with error: " 
                          << status.error_code() << ": " << status.error_message() << std::endl;
            } else {
                std::cout << "gRPC Stream finished successfully. Events received by server: " 
                          << response_.events_received() << std::endl;
            }
        }
    }

    void SendEvent(const ARPEvent& event) {
        if (stream_) {
            if (!stream_->Write(event)) {
                std::cerr << "Failed to write event to gRPC stream." << std::endl;
            }
        }
    }

private:
    std::unique_ptr<ARPCollector::Stub> stub_;
    ClientContext context_;
    ARPEventResponse response_;
    std::unique_ptr<ClientWriter<ARPEvent>> stream_;
};

static std::unique_ptr<ARPClient> g_client = nullptr;

extern "C" {

void init_grpc_client(const char* target) {
    if (!g_client) {
        std::string target_str(target);
        g_client = std::make_unique<ARPClient>(grpc::CreateChannel(
            target_str, grpc::InsecureChannelCredentials()));
        std::cout << "Initialized gRPC client connected to " << target_str << std::endl;
    }
}

void send_arp_event(
    const char* agent_id,
    uint32_t opcode,
    const uint8_t* target_ip,
    const uint8_t* target_mac,
    const uint8_t* sender_ip,
    const uint8_t* sender_mac
) {
    if (g_client) {
        ARPEvent event;
        event.set_agent_id(agent_id);
        
        // Current timestamp in milliseconds
        auto now = std::chrono::system_clock::now();
        auto duration = now.time_since_epoch();
        event.set_timestamp(std::chrono::duration_cast<std::chrono::milliseconds>(duration).count());
        
        event.set_opcode(opcode);
        event.set_target_ip(target_ip, 4);
        event.set_target_mac(target_mac, 6);
        event.set_sender_id(sender_ip, 4); // `sender_id` in proto corresponds to sender_ip based on size 4
        event.set_sender_mac(sender_mac, 6);

        g_client->SendEvent(event);
    }
}

void destroy_grpc_client() {
    g_client.reset();
}

}
