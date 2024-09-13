package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/goburrow/modbus"
	"github.com/tarm/serial"
)

type Response struct {
	Data uint16 `json:"data"` // Tek bir veri için struct
}

func main() {
	// Seri port konfigürasyonu
	config := &serial.Config{
		Name:     "COM8", // Port adı
		Baud:     19200,  // Baud rate
		Parity:   serial.ParityNone,
		StopBits: serial.Stop1, // Stop bits
		Size:     8,            // Data bits
	}

	// Modbus RTU client handler oluşturma
	handler := modbus.NewRTUClientHandler(config.Name)
	handler.BaudRate = config.Baud
	handler.Parity = string(config.Parity) // Parity değerini string'e dönüştür
	handler.Timeout = 10 * time.Second     // Zaman aşımı ayarı

	// Modbus RTU client oluşturma
	client := modbus.NewClient(handler)

	// Slave ID ve adres
	slaveID := byte(10)
	address := uint16(9)   // Okuyacağınız veri adresi
	quantity := uint16(10) // Okuyacağınız veri miktarı (bu örnekte 10 register okuyacağız)

	// Slave ID ayarlama
	handler.SlaveId = slaveID

	http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		// Veriyi okuma
		results, err := client.ReadHoldingRegisters(address, quantity)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(results) < int(quantity)*2 {
			http.Error(w, "No data returned", http.StatusInternalServerError)
			return
		}

		// Results byte dizisini uint16 dizisine dönüştürme
		var data []uint16
		for i := 0; i < len(results); i += 2 {
			value := uint16(results[i])<<8 | uint16(results[i+1])
			data = append(data, value)
		}

		// İlk öğeyi almak
		if len(data) > 0 {
			resp := Response{Data: data[0]}
			jsonResp, err := json.Marshal(resp)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonResp)
		} else {
			http.Error(w, "No data available", http.StatusInternalServerError)
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
