#include "grpc_client.h"
#include "config.h"
#include "log.h"
#include <iostream>
#include <memory>
#include <string>
#include <chrono>
#include <fstream>
#include <sstream>

#include <grpcpp/grpcpp.h>
#include "pb/collector.grpc.pb.h"

using grpc::Channel;
using grpc::ClientContext;
using grpc::ClientWriter;
using grpc::Status;
using distarpwatcher::ARPCollector;
using distarpwatcher::ARPEvent;
using distarpwatcher::ARPEventResponse;

static std::string get_file_contents(const char *filename) {
    std::ifstream in(filename, std::ios::in | std::ios::binary);
    if (in) {
        std::ostringstream contents;
        contents << in.rdbuf();
        in.close();
        return contents.str();
    }
    LOG_WARN("Could not read certificate file: %s", filename);
    return "";
}

class ARPClient {
public:
    ARPClient(std::shared_ptr<Channel> channel)
        : channel_(channel), stub_(ARPCollector::NewStub(channel)) {
        connectStream();
    }

    ~ARPClient() {
        closeStream();
    }

    void SendEvent(const ARPEvent& event) {
        if (!stream_) {
            LOG_INFO("gRPC stream not active. Attempting to reconnect...");
            connectStream();
        }

        if (stream_) {
            if (!stream_->Write(event)) {
                stream_->WritesDone();
                Status status = stream_->Finish();

                LOG_ERR("gRPC Write failed. Reason: [%d] %s", status.error_code(), status.error_message().c_str());

                if (!status.error_details().empty()) {
                    LOG_ERR("gRPC Error Details: %s", status.error_details().c_str());
                }

                stream_ = nullptr;
            }
        } else {
            LOG_ERR("gRPC Reconnection failed. Dropping event.");
        }
    }

private:
    void connectStream() {
        context_ = std::make_unique<ClientContext>();
        stream_ = stub_->ARPStream(context_.get(), &response_);
    }

    void closeStream() {
        if (stream_) {
            stream_->WritesDone();
            Status status = stream_->Finish();
            if (!status.ok()) {
                LOG_ERR("gRPC Stream closed with error code %d: %s", status.error_code(), status.error_message().c_str());
            } else {
                LOG_INFO("gRPC Stream finished successfully. Events received by server: %d", response_.events_received());
            }
            stream_ = nullptr;
        }
    }

    std::shared_ptr<Channel> channel_;
    std::unique_ptr<ARPCollector::Stub> stub_;
    std::unique_ptr<ClientContext> context_;
    ARPEventResponse response_;
    std::unique_ptr<ClientWriter<ARPEvent>> stream_;
};
static std::unique_ptr<ARPClient> g_client = nullptr;

extern "C" {

void init_grpc_client(const char* target) {
    if (!g_client) {
        std::string target_str(target);
        
        grpc::SslCredentialsOptions ssl_opts;
        ssl_opts.pem_root_certs = get_file_contents(global_config.ca_cert);
        ssl_opts.pem_private_key = get_file_contents(global_config.client_key);
        ssl_opts.pem_cert_chain = get_file_contents(global_config.client_cert);

        auto channel_creds = grpc::SslCredentials(ssl_opts);

        g_client = std::make_unique<ARPClient>(grpc::CreateChannel(
            target_str, channel_creds));
        
        LOG_INFO("Initialized gRPC client (mTLS) connected to %s", target);
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
