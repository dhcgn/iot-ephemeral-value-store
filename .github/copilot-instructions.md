# Copilot Instructions for IoT Ephemeral Value Store

## Project Overview
This project is an IoT Ephemeral Value Store server implemented in Go. It provides a simple HTTP interface for storing and retrieving temporary IoT data.

## Key Components

### Main Packages
- `net/http`: For HTTP server functionality
- `github.com/gorilla/mux`: For routing HTTP requests
- `github.com/dgraph-io/badger/v3`: For data storage
- `golang.org/x/time/rate`: For rate limiting

### Custom Packages
- `domain`: Handles key generation and validation
- `httphandler`: Contains HTTP request handlers
- `middleware`: Implements middleware functions
- `stats`: Manages server statistics
- `storage`: Handles data storage operations

## Code Style and Conventions
- Use Go standard formatting (gofmt)
- Follow Go best practices for error handling
- Use meaningful variable and function names
- Implement unit tests for all packages

## Key Functionalities
- Generate unique key pairs for data upload and retrieval
- Store data temporarily with configurable duration
- Retrieve data in JSON and plain text formats
- Support patch operations for combining multiple uploads
- Implement rate limiting and request size limiting
- Provide statistics on server usage

## API Endpoints
- `/kp`: Generate key pair
- `/u/{uploadKey}`: Upload data
- `/d/{downloadKey}/json`: Download data as JSON
- `/d/{downloadKey}/plain/{param}`: Download specific data field
- `/d/{downloadKey}/plain-from-base64url/{param}`: Download specific data field, server-sided decoded from json
- `/patch/{uploadKey}/{param}`: Patch data
- `/delete/{uploadKey}`: Delete data

## Configuration
- Configurable data retention period
- Adjustable rate limiting parameters
- Customizable maximum request size

## Deployment
- Supports Docker deployment
- Provides systemd service installation script

When working on this project, ensure that new features and modifications align with the existing architecture and maintain the simplicity of the API for IoT devices.