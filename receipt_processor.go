package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Data structures based on the API specification
type Receipt struct {
	Retailer     string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Items        []Item `json:"items"`
	Total        string `json:"total"`
}

type Item struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

type ReceiptResponse struct {
	ID string `json:"id"`
}

type PointsResponse struct {
	Points int `json:"points"`
}

// In-memory storage
type ReceiptStore struct {
	sync.RWMutex
	receipts map[string]Receipt
	points   map[string]int
}

func NewReceiptStore() *ReceiptStore {
	return &ReceiptStore{
		receipts: make(map[string]Receipt),
		points:   make(map[string]int),
	}
}

func (rs *ReceiptStore) AddReceipt(receipt Receipt) string {
	rs.Lock()
	defer rs.Unlock()

	id := uuid.New().String()
	rs.receipts[id] = receipt
	
	// Calculate points for the receipt
	points := calculatePoints(receipt)
	rs.points[id] = points
	
	return id
}

func (rs *ReceiptStore) GetPoints(id string) (int, bool) {
	rs.RLock()
	defer rs.RUnlock()
	
	points, exists := rs.points[id]
	return points, exists
}

// Points calculation logic
func calculatePoints(receipt Receipt) int {
	points := 0

	// Rule 1: One point for every alphanumeric character in the retailer name
	alphanumericRegex := regexp.MustCompile(`[a-zA-Z0-9]`)
	retailerAlphanumeric := alphanumericRegex.FindAllString(receipt.Retailer, -1)
	points += len(retailerAlphanumeric)

	// Rule 2: 50 points if the total is a round dollar amount with no cents
	total, _ := strconv.ParseFloat(receipt.Total, 64)
	if total == math.Floor(total) {
		points += 50
	}

	// Rule 3: 25 points if the total is a multiple of 0.25
	if math.Mod(total*100, 25) == 0 {
		points += 25
	}

	// Rule 4: 5 points for every two items on the receipt
	points += (len(receipt.Items) / 2) * 5

	// Rule 5: If the trimmed length of the item description is a multiple of 3, 
	// multiply the price by 0.2 and round up to the nearest integer
	for _, item := range receipt.Items {
		trimmedDesc := strings.TrimSpace(item.ShortDescription)
		if len(trimmedDesc)%3 == 0 {
			price, _ := strconv.ParseFloat(item.Price, 64)
			points += int(math.Ceil(price * 0.2))
		}
	}

	// Rule 6: 6 points if the day in the purchase date is odd
	purchaseDate, _ := time.Parse("2006-01-02", receipt.PurchaseDate)
	if purchaseDate.Day()%2 == 1 {
		points += 6
	}

	// Rule 7: 10 points if the time of purchase is after 2:00pm and before 4:00pm
	purchaseTime, _ := time.Parse("15:04", receipt.PurchaseTime)
	purchaseHour := purchaseTime.Hour()
	purchaseMinute := purchaseTime.Minute()
	if (purchaseHour == 14 && purchaseMinute > 0) || 
	   (purchaseHour == 15) || 
	   (purchaseHour == 16 && purchaseMinute == 0) {
		points += 10
	}

	return points
}

// HTTP Handlers
func (rs *ReceiptStore) ProcessReceiptHandler(w http.ResponseWriter, r *http.Request) {
	var receipt Receipt
	err := json.NewDecoder(r.Body).Decode(&receipt)
	if err != nil {
		http.Error(w, "Invalid receipt format", http.StatusBadRequest)
		return
	}

	// Basic validation
	if receipt.Retailer == "" || receipt.PurchaseDate == "" || receipt.PurchaseTime == "" || receipt.Total == "" {
		http.Error(w, "Missing required receipt fields", http.StatusBadRequest)
		return
	}

	// Validate date format (YYYY-MM-DD)
	_, err = time.Parse("2006-01-02", receipt.PurchaseDate)
	if err != nil {
		http.Error(w, "Invalid purchase date format. Expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	// Validate time format (HH:MM)
	_, err = time.Parse("15:04", receipt.PurchaseTime)
	if err != nil {
		http.Error(w, "Invalid purchase time format. Expected HH:MM", http.StatusBadRequest)
		return
	}

	// Validate total format (number with optional decimal point)
	_, err = strconv.ParseFloat(receipt.Total, 64)
	if err != nil {
		http.Error(w, "Invalid total format", http.StatusBadRequest)
		return
	}

	// Process receipt and generate ID
	id := rs.AddReceipt(receipt)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ReceiptResponse{ID: id})
}

func (rs *ReceiptStore) GetPointsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	points, exists := rs.GetPoints(id)
	if !exists {
		http.Error(w, "No receipt found for that id", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(PointsResponse{Points: points})
}

func main() {
	store := NewReceiptStore()
	router := mux.NewRouter()

	// Define API routes
	router.HandleFunc("/receipts/process", store.ProcessReceiptHandler).Methods("POST")
	router.HandleFunc("/receipts/{id}/points", store.GetPointsHandler).Methods("GET")

	// Start the server
	fmt.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}
