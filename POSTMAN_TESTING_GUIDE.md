# Postman Testing Guide for Complaint Handlers

## Overview
This guide provides step-by-step instructions to test the new `GetComplaints` and `UpdateComplaint` handlers using Postman.

## Prerequisites
1. Start the application: `go run ./cmd/app/main.go`
2. The server should run on `http://localhost:YOUR_PORT` (check your config)
3. Have Postman installed

## Setup in Postman

### 1. Create Environment Variables
To make testing easier, create environment variables in Postman:

1. Click **Environments** in the left sidebar
2. Click **Create New Environment** → Name it "Complaints API"
3. Add the following variables:

| Variable | Value |
|----------|-------|
| `base_url` | `http://localhost:8080` (adjust port as needed) |
| `student_token` | (Will be populated after login) |
| `admin_token` | (Will be populated after login) |
| `complaint_id` | (Will be populated after creating a complaint) |

---

## Step-by-Step Testing

### Step 1: Register a Student Account

**Request:**
- **Method**: POST
- **URL**: `{{base_url}}/api/auth/register`
- **Headers**: `Content-Type: application/json`
- **Body** (JSON):
```json
{
  "email": "student@example.com",
  "name": "John Student",
  "username": "student123",
  "password": "StudentPass123!",
  "role": "student"
}
```

**Expected Response**: 201 Created
```json
{
  "id": "student-uuid",
  "email": "student@example.com",
  "name": "John Student",
  "username": "student123",
  "role": "student",
  "createdAt": "2026-02-09T..."
}
```

---

### Step 2: Register an Admin Account

**Request:**
- **Method**: POST
- **URL**: `{{base_url}}/api/auth/register`
- **Headers**: `Content-Type: application/json`
- **Body** (JSON):
```json
{
  "email": "admin@example.com",
  "name": "Jane Admin",
  "username": "admin123",
  "password": "AdminPass123!",
  "role": "admin"
}
```

**Expected Response**: 201 Created

---

### Step 3: Login as Student

**Request:**
- **Method**: POST
- **URL**: `{{base_url}}/api/auth/login`
- **Headers**: `Content-Type: application/json`
- **Body** (JSON):
```json
{
  "email": "student@example.com",
  "password": "StudentPass123!"
}
```

**Expected Response**: 200 OK
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Action**: 
- Copy the token value
- Go to Postman Environment variables
- Paste it into `student_token` variable

---

### Step 4: Login as Admin

**Request:**
- **Method**: POST
- **URL**: `{{base_url}}/api/auth/login`
- **Headers**: `Content-Type: application/json`
- **Body** (JSON):
```json
{
  "email": "admin@example.com",
  "password": "AdminPass123!"
}
```

**Expected Response**: 200 OK
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Action**: 
- Copy the token value
- Go to Postman Environment variables
- Paste it into `admin_token` variable

---

### Step 5: Create a Complaint (as Student)

**Request:**
- **Method**: POST
- **URL**: `{{base_url}}/api/complaints`
- **Headers**: 
  - `Content-Type: application/json`
  - `Authorization: Bearer {{student_token}}`
- **Body** (JSON):
```json
{
  "description": "The library is too noisy during study hours"
}
```

**Expected Response**: 201 Created
```json
{
  "id": "complaint-uuid-123",
  "userId": "student-uuid",
  "description": "The library is too noisy during study hours",
  "status": "pending",
  "createdAt": "2026-02-09T10:30:00Z"
}
```

**Action**: 
- Copy the `id` value
- Go to Postman Environment variables
- Paste it into `complaint_id` variable

---

### Step 6: Get Student's Own Complaints (as Student)

**Request:**
- **Method**: GET
- **URL**: `{{base_url}}/api/complaints`
- **Headers**: 
  - `Authorization: Bearer {{student_token}}`

**Expected Response**: 200 OK
```json
[
  {
    "id": "complaint-uuid-123",
    "userId": "student-uuid",
    "description": "The library is too noisy during study hours",
    "status": "pending",
    "createdAt": "2026-02-09T10:30:00Z"
  }
]
```

**Note**: Students can only see their own complaints

---

### Step 7: Get Student's Complaints Filtered by Status (as Student)

**Request:**
- **Method**: GET
- **URL**: `{{base_url}}/api/complaints?status=pending`
- **Headers**: 
  - `Authorization: Bearer {{student_token}}`

**Expected Response**: 200 OK
```json
[
  {
    "id": "complaint-uuid-123",
    "userId": "student-uuid",
    "description": "The library is too noisy during study hours",
    "status": "pending",
    "createdAt": "2026-02-09T10:30:00Z"
  }
]
```

**Test Other Status Values**: Try `status=approved` or `status=rejected`

---

### Step 8: Get Single Complaint by ID (as Admin)

**Request:**
- **Method**: GET
- **URL**: `{{base_url}}/api/complaints?id={{complaint_id}}`
- **Headers**: 
  - `Authorization: Bearer {{admin_token}}`

**Expected Response**: 200 OK
```json
[
  {
    "id": "complaint-uuid-123",
    "userId": "student-uuid",
    "description": "The library is too noisy during study hours",
    "status": "pending",
    "createdAt": "2026-02-09T10:30:00Z"
  }
]
```

**Note**: Only admins can retrieve complaints by ID

---

### Step 9: Student Cannot Access Admin Features (Negative Test)

**Request:**
- **Method**: GET
- **URL**: `{{base_url}}/api/complaints?id={{complaint_id}}`
- **Headers**: 
  - `Authorization: Bearer {{student_token}}`

**Expected Response**: 200 OK (but empty array)
```json
[]
```

**Note**: Students cannot use the `id` parameter; they can only see their own complaints

---

### Step 10: Update Complaint Status (as Admin)

**Request:**
- **Method**: PUT
- **URL**: `{{base_url}}/api/complaints/{{complaint_id}}`
- **Headers**: 
  - `Content-Type: application/json`
  - `Authorization: Bearer {{admin_token}}`
- **Body** (JSON):
```json
{
  "status": "approved"
}
```

**Expected Response**: 200 OK
```json
{
  "message": "Complaint status updated successfully",
  "complaintId": "complaint-uuid-123",
  "status": "approved"
}
```

---

### Step 11: Verify Status Changed (as Student)

**Request:**
- **Method**: GET
- **URL**: `{{base_url}}/api/complaints`
- **Headers**: 
  - `Authorization: Bearer {{student_token}}`

**Expected Response**: 200 OK
```json
[
  {
    "id": "complaint-uuid-123",
    "userId": "student-uuid",
    "description": "The library is too noisy during study hours",
    "status": "approved",
    "createdAt": "2026-02-09T10:30:00Z"
  }
]
```

**Note**: Notice the status changed from "pending" to "approved"

---

### Step 12: Update to "Rejected" Status (as Admin)

**Request:**
- **Method**: PUT
- **URL**: `{{base_url}}/api/complaints/{{complaint_id}}`
- **Headers**: 
  - `Content-Type: application/json`
  - `Authorization: Bearer {{admin_token}}`
- **Body** (JSON):
```json
{
  "status": "rejected"
}
```

**Expected Response**: 200 OK

---

## Error Test Cases

### Test 1: Missing Authorization Token
**Request:**
- **Method**: GET
- **URL**: `{{base_url}}/api/complaints`
- **Headers**: (No Authorization header)

**Expected Response**: 401 Unauthorized
```json
"Authentication required"
```

---

### Test 2: Invalid Token
**Request:**
- **Method**: GET
- **URL**: `{{base_url}}/api/complaints`
- **Headers**: 
  - `Authorization: Bearer invalid-token-123`

**Expected Response**: 401 Unauthorized
```json
"Invalid token"
```

---

### Test 3: Student Trying to Update Complaint (Negative Test)
**Request:**
- **Method**: PUT
- **URL**: `{{base_url}}/api/complaints/{{complaint_id}}`
- **Headers**: 
  - `Content-Type: application/json`
  - `Authorization: Bearer {{student_token}}`
- **Body** (JSON):
```json
{
  "status": "approved"
}
```

**Expected Response**: 403 Forbidden
```json
"Forbidden: admin access required"
```

---

### Test 4: Invalid Status Value
**Request:**
- **Method**: PUT
- **URL**: `{{base_url}}/api/complaints/{{complaint_id}}`
- **Headers**: 
  - `Content-Type: application/json`
  - `Authorization: Bearer {{admin_token}}`
- **Body** (JSON):
```json
{
  "status": "invalid-status"
}
```

**Expected Response**: 400 Bad Request
```json
"Invalid status value"
```

---

### Test 5: Empty Status Value
**Request:**
- **Method**: PUT
- **URL**: `{{base_url}}/api/complaints/{{complaint_id}}`
- **Headers**: 
  - `Content-Type: application/json`
  - `Authorization: Bearer {{admin_token}}`
- **Body** (JSON):
```json
{
  "status": ""
}
```

**Expected Response**: 400 Bad Request
```json
"Status cannot be empty"
```

---

### Test 6: Missing Complaint ID in URL
**Request:**
- **Method**: PUT
- **URL**: `{{base_url}}/api/complaints/`
- **Headers**: 
  - `Content-Type: application/json`
  - `Authorization: Bearer {{admin_token}}`
- **Body** (JSON):
```json
{
  "status": "approved"
}
```

**Expected Response**: 400 Bad Request
```json
"Complaint ID required"
```

---

### Test 7: Non-existent Complaint ID
**Request:**
- **Method**: GET
- **URL**: `{{base_url}}/api/complaints?id=non-existent-id-123`
- **Headers**: 
  - `Authorization: Bearer {{admin_token}}`

**Expected Response**: 404 Not Found
```json
"Complaint not found"
```

---

## Summary of Endpoints

| Method | Endpoint | Auth | Role | Description |
|--------|----------|------|------|-------------|
| POST | `/api/complaints` | ✅ Required | student/admin | Create a new complaint |
| GET | `/api/complaints` | ✅ Required | student/admin | Get complaints (student: own; admin: all/by-id) |
| PUT | `/api/complaints/{id}` | ✅ Required | admin only | Update complaint status |

---

## Expected Behavior Summary

### GetComplaints Handler
- **Student Role**: 
  - Returns only their own complaints
  - Can filter by status using `?status=pending`
  - Cannot use `?id=` parameter
  
- **Admin Role**: 
  - Can retrieve a specific complaint using `?id=complaint-uuid`
  - Can filter all complaints by status using `?status=pending`
  - Returns an array in both cases

### UpdateComplaint Handler
- **Admin-only access** - Enforced by middleware
- **Validates status**: Only "pending", "approved", "rejected" are allowed
- **Sends to Service Bus**: Complaint ID is queued for async processing
- **Logs changes**: Includes adminId, complaintId, newStatus in logs
- **Returns success JSON** with confirmation details

---

## Notes
- Replace `http://localhost:8080` with your actual server URL if different
- All tokens use Bearer scheme: `Authorization: Bearer {token}`
- Tokens expire after 24 hours (configurable in middleware)
- Service Bus messages are sent asynchronously - check your Service Bus queue for messages

