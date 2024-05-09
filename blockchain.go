package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

)

type Block struct {
	Data         map[string]interface{} `json:"data"`
	Hash         string                 `json:"hash"`
	PreviousHash string                 `json:"previous_hash"`
	Timestamp    time.Time              `json:"timestamp"`
	Pow          int                    `json:"pow"`
}

type Blockchain struct {
	GenesisBlock Block   `json:"genesis_block"`
	Chain        []Block `json:"chain"`
	Difficulty   int     `json:"difficulty"`
}

var blockchain Blockchain

func (b *Block) calculateHash() string {
	data, _ := json.Marshal(b.Data)
	blockData := b.PreviousHash + string(data) + b.Timestamp.String() + strconv.Itoa(b.Pow)
	blockHash := sha256.Sum256([]byte(blockData))
	return fmt.Sprintf("%x", blockHash)
}

func (b *Block) mine(difficulty int) {
	for !strings.HasPrefix(b.Hash, strings.Repeat("0", difficulty)) {
		b.Pow++
		b.Hash = b.calculateHash()
	}
}

func CreateBlockchain(difficulty int) {
	genesisBlock := Block{
		Hash:      "0",
		Timestamp: time.Now(),
	}
	blockchain = Blockchain{
		GenesisBlock: genesisBlock,
		Chain:        []Block{genesisBlock},
		Difficulty:   difficulty,
	}
}

func (b *Blockchain) addBlock(from, to string, amount float64) {
	blockData := map[string]interface{}{
		"from":   from,
		"to":     to,
		"amount": amount,
	}
	lastBlock := b.Chain[len(b.Chain)-1]
	newBlock := Block{
		Data:         blockData,
		PreviousHash: lastBlock.Hash,
		Timestamp:    time.Now(),
	}
	newBlock.mine(b.Difficulty)
	b.Chain = append(b.Chain, newBlock)
}

func (b Blockchain) isValid() bool {
	for i := range b.Chain[1:] {
		previousBlock := b.Chain[i]
		currentBlock := b.Chain[i+1]
		if currentBlock.Hash != currentBlock.calculateHash() || currentBlock.PreviousHash != previousBlock.Hash {
			return false
		}
	}
	return true
}

func getBlockchain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(blockchain)
}

func addBlock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var data map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	from, ok := data["from"].(string)
	if !ok {
		http.Error(w, "Invalid 'from' field", http.StatusBadRequest)
		return
	}

	to, ok := data["to"].(string)
	if !ok {
		http.Error(w, "Invalid 'to' field", http.StatusBadRequest)
		return
	}

	amount, ok := data["amount"].(float64)
	if !ok {
		http.Error(w, "Invalid 'amount' field", http.StatusBadRequest)
		return
	}

	blockchain.addBlock(from, to, amount)

	json.NewEncoder(w).Encode(blockchain)
}

func main() {
	CreateBlockchain(2)

	r := mux.NewRouter()
	r.HandleFunc("/blockchain", getBlockchain).Methods("GET")
	r.HandleFunc("/blockchain/add", addBlock).Methods("POST")

	http.ListenAndServe(":8088", r)
}
