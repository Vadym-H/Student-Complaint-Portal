# Student Complaint Portal

A Go-based complaint management system using Azure Cosmos DB and Azure Service Bus.

## 🏗️ Architecture

- **Backend**: Go (Chi router)
- **Database**: Azure Cosmos DB
- **Message Queue**: Azure Service Bus
- **Infrastructure**: Terraform (IaC)
- **Authentication**: JWT-based auth

## 📋 Prerequisites

- Go 1.25.5 or higher
- Azure account with active subscription
- Terraform (for infrastructure deployment)

## 🚀 Setup Instructions

### 1. Clone the Repository

```bash
git clone https://github.com/Vadym-H/Student-Complaint-Portal.git
cd Student-Complaint-Portal
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Configure Environment Variables

Copy the example environment file and update with your values:

```bash
cp .env.example .env
```

Edit `.env` and fill in your actual Azure credentials:

- `COSMOS_ENDPOINT`: Your Azure Cosmos DB endpoint
- `COSMOS_KEY`: Your Azure Cosmos DB primary key
- `SERVICE_BUS_CONNECTION`: Your Azure Service Bus connection string
- `JWT_SECRET`: A secure random string (minimum 32 characters)

**⚠️ IMPORTANT: Never commit the `.env` file to Git!**

### 4. Deploy Infrastructure (Optional)

If you need to deploy Azure infrastructure:

```bash
cd terraform
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your settings
terraform init
terraform plan
terraform apply
```

### 5. Run the Application

```bash
go run cmd/app/main.go
```

The server will start on `http://localhost:8080` (or the port specified in your `.env` file).

## 📁 Project Structure

```
.
├── cmd/
│   └── app/
│       └── main.go           # Application entry point
├── internal/
│   ├── config/               # Configuration management
│   ├── handlers/             # HTTP handlers
│   ├── lib/logger/          # Logging utilities
│   ├── middleware/          # HTTP middleware
│   ├── models/              # Data models
│   └── services/            # Business logic
├── terraform/               # Infrastructure as Code
├── .env.example            # Environment variables template
├── go.mod                  # Go dependencies
└── README.md              # This file
```

## 🔒 Security Notes

The following files contain sensitive information and are **NEVER** committed to Git:

- `.env` - Contains all your secrets (API keys, connection strings)
- `terraform/*.tfstate` - Contains infrastructure state with sensitive data
- `terraform/*.tfvars` - Contains your actual Terraform variables

Always use the `.example` versions as templates.

## 🛠️ Development

### Running Tests

```bash
go test ./...
```

### Building for Production

```bash
go build -o bin/app cmd/app/main.go
```

## 📝 API Endpoints

- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - Login user
- `GET /api/complaints` - List complaints
- `POST /api/complaints` - Create complaint
- `GET /api/complaints/{id}` - Get complaint by ID
- `PUT /api/complaints/{id}` - Update complaint
- `DELETE /api/complaints/{id}` - Delete complaint

## 🤝 Contributing

1. Create a feature branch
2. Make your changes
3. Test thoroughly
4. Submit a pull request

## 📄 License

[Add your license here]

## 👤 Author

Vadym H. - [GitHub Profile](https://github.com/Vadym-H)

