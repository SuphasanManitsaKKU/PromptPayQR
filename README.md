# Online Payment System

## Tech Stack
- **Frontend**: Next.js
- **Backend**: Go
- **Database**: MySQL (using GORM as ORM)

## Description
This project is an online payment system that allows users to:
1. Generate QR codes for PromptPay transfers.
2. Verify payments via slips using the EasySlip API.

## Outcome
The system enhances user convenience by providing quick and seamless payment generation and verification.

---

## Important Note
🚨 **Important**: The EasySlip API is no longer available due to the quota being exhausted.  
- You can still generate QR codes for PromptPay transfers as usual.  
- However, payment verification via slips is currently not functional because the EasySlip API quota has been used up.

---

## Setup

### Run the Project
To start the system, use the following command:
```bash
docker-compose up --build -d