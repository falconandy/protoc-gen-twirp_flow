# Twirp Flow Plugin

A protoc plugin for generating a twirp client suitable for specific needs.

Based on [Twirp Typescript Plugin](https://github.com/larrymyers/protoc-gen-twirp_typescript)

This plugin supports:

* A minimal standalone client that supports JSON transport only.

## Setup

The protobuf v3 compiler is required. You can get the latest precompiled binary for your system here:

[https://github.com/google/protobuf/releases]

## Usage
    
All generated files will be placed relative to the specified output directory for the plugin.  
This is different behavior than the twirp Go plugin, which places the files relative to the input proto files.

This decision is intentional, since only client code destination is likely somewhere different
than the server code.

### Generating Code for Twirp v6

By default code is generated that supports the twirp v5 spec. If you want to use the the v6 prerelease specify the 
version using protoc params.

    protoc --twirp_flow_out=version=v6:<path-to-project> <path-to-proto-file>
    
The relevant change here is the routing path, which now starts with the proto package instead of "twirp/".
