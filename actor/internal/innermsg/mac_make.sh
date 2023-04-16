#!/bin/bash
protoc --gogofaster_out=plugins=grpc:. msg.proto
