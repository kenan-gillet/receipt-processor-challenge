# Receipt Processor Service

This is a Go implementation of a Receipt Processor service that calculates points for receipts based on specific rules.

## Overview

The Receipt Processor is a RESTful API service that:
1. Accepts receipt data via a POST endpoint
2. Processes receipts to calculate reward points based on predefined rules
3. Allows retrieval of points for a processed receipt via a GET endpoint

## API Endpoints

### Process Receipt
- **URL**: `/receipts/process`
- **Method**: `POST`
- **Request Body**: Receipt JSON object
- **Response**: JSON object with ID of the processed receipt
- **Status Codes**: 
  - `200 OK`: Receipt processed successfully
  - `400 Bad Request`: Invalid receipt data

### Get Points
- **URL**: `/receipts/{id}/points`
- **Method**: `GET`
- **Response**: JSON object with points for the receipt
- **Status Codes**: 
  - `200 OK`: Points retrieved successfully
  - `404 Not Found`: No receipt found for the given ID

## Data Models

### Receipt
```json
{
  "retailer": "Target",
  "purchaseDate": "2022-01-01",
  "purchaseTime": "13:01",
  "items": [
    {
      "shortDescription": "Mountain Dew 12PK",
      "price": "6.49"
    },
    {
      "shortDescription": "Emils Cheese Pizza",
      "price": "12.25"
    }
  ],
  "total": "18.74"
}
```

### Points Calculation Rules

Points are calculated based on the following rules:

1. One point for every alphanumeric character in the retailer name
2. 50 points if the total is a round dollar amount with no cents
3. 25 points if the total is a multiple of 0.25
4. 5 points for every two items on the receipt
5. If the trimmed length of the item description is a multiple of 3, multiply the price by 0.2 and round up to the nearest integer
6. 6 points if the day in the purchase date is odd
7. 10 points if the time of purchase is after 2:00pm and before 4:00pm

## How to Run

### Prerequisites
- Go 1.16+
- Following dependencies:
  - github.com/google/uuid
  - github.com/gorilla/mux
  - github.com/stretchr/testify (for tests)

### Installation

1. Clone the repository
2. Install dependencies:
   ```
   go get github.com/google/uuid
   go get github.com/gorilla/mux
   go get github.com/stretchr/testify/assert
   ```

### Running the Service
```
go run main.go
```

The service will start on port 8080.

### Running Tests
```
go test
```

## Example Usage

### Process a receipt
```bash
curl -X POST -H "Content-Type: application/json" -d '{
  "retailer": "Target",
  "purchaseDate": "2022-01-01",
  "purchaseTime": "13:01",
  "items": [
    {
      "shortDescription": "Mountain Dew 12PK",
      "price": "6.49"
    },
    {
      "shortDescription": "Emils Cheese Pizza",
      "price": "12.25"
    }
  ],
  "total": "18.74"
}' http://localhost:8080/receipts/process
```

### Get points for a receipt
```bash
curl http://localhost:8080/receipts/{id}/points
```
(Replace `{id}` with the ID returned from the process endpoint)
