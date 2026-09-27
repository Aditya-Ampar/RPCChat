# RPC Chat

A concurrent RPC-Based chat service written in Go.

This project is being developed incrementally with a focus on:

- GO concurrency and goroutines
- RPC-based client/server architecture
- SQL database integration
- Network Security
- Authentication and authorization
- Vulnerability assessment and penetration testing
- Secure software development
- Security Hardening
- Dependency and patch management
- Automated security testing


## Current Version

**v0.1.0 - Initial RPC Chat Skeleton**

The current version provides:

- Go RPC server
- Go RPC client
- TCP communication
- Concurrent client handling using goroutines
- Basic 'SendMessage' RPC Method

## Architecture

'''text
Client
   |
   | TCP
   v
Go RPC Server
   |
   |-- goroutine per connection
   |
   v
ChatService
   


## Current Limitation
- ~~Authentication~~
- Authorization
- ~~TLS~~
- Persistent Storage
- ~~Input Validation~~
- ~~Rate Limiting~~
- Security Monitoring
