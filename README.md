# Generic Shop Sample 🛍️
## Project Overview

A Comprehensive Go-based E-Commerce Platform

## 🚀 Project Status
```
Disclaimer: This project is a learning playground and not yet production-ready. It's designed for educational purposes and
experimentation features
```

## Core Functionalities

### 👤 User Management
 - Authentication
 - Profile management
 - Role-based access control

### 📦 Product Catalog
 - CRUD operations
 - Category hierarchies
 - Flexible product management

### 🛒 Ordering System
 - Complete order workflow

### 💳 Payment Integration
 - Zarin-Pal payment gateway

### 📧 Communication Services
 - Email verifier

### 💬 Community Interaction
 - Product commenting system
 - User reviews

### 🛠 Technology Stack
 - Language    Go (Golang)	1.21+
 - Web framework   Gin
 - CLI management  Cobra
 - Row sql and database    postgresSQL
 - Caching system  redis

 - Payment gateway Zarin-Pal   v4

## 🚀 Quick Start

Prerequisites
```bash
Go 1.21+
Docker (optional)
Zarin-Pal Merchant Account
```
Local Development Setup
1. Clone the Repository
```
git clone https://github.com/yourusername/generic-shop-sample.git
cd generic-shop-sample
```

2. Configuration
```bash
# Copy environment template
cp .env.example .env

# Edit .env file with your specific configurations
# Set database, payment gateway, and email service credentials
nano .env
```

3. Environment Variables
```bash

# Set precise path to your .env file
export DOTENV_PATH=~/exact/path/to/.env
```

4. Build & Run
Local Execution
```bash
# Build the application
go build .

# Start the server
./generic-shop-sample serve

# Create admin user
./generic-shop-sample new-admin -u admin -p securePassword
```

Docker Deployment
```bash
# Build and start services
docker compose build
docker compose up -d

# View logs
docker compose logs -f
```

## 🔐 Security Features
 - secure password hashing base
 - JWT Authentication
 - Role-Based Access Control

## 🧪 Testing
Run Tests
```bash
## Unit Tests
go test ./...

# Integration Tests
go test -tags=integration ./...
```

## Contact
 - email: daniel-k84@outlook.com
 - [linkedin](https://www.linkedin.com/in/daniel-karami-786524346)