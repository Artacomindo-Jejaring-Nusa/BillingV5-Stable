package main

import (
	"fmt"
	"log"

	"billing-backend/config"
	"billing-backend/pkg/utils"
)

func main() {
	cfg := config.LoadConfig()
	targetPhone := "628986937819"

	brands := []string{"JAKINET", "JELANTIK"}

	for _, brand := range brands {
		integrationID := cfg.GetQontakIntegrationID(brand)
		templateID := cfg.GetQontakTemplateID(brand)

		log.Printf("==================================================")
		log.Printf("Testing Mekari Qontak Broadcast for brand: %s", brand)
		log.Printf("Target Phone : %s", targetPhone)
		log.Printf("Integration ID: %s", integrationID)
		log.Printf("Template ID   : %s", templateID)
		log.Printf("==================================================")

		client := utils.NewMekariQontakClient(cfg.QontakClientID, cfg.QontakClientSecret, integrationID)

		payload := utils.DirectBroadcastPayload{
			ToName:               "Adolf Renaldy (Test)",
			ToNumber:             targetPhone,
			MessageTemplateID:    templateID,
			ChannelIntegrationID: integrationID,
			Language:             utils.QontakLanguage{Code: "id"},
			Parameters: utils.QontakParameters{
				Body: []utils.QontakParamBody{
					{Key: "1", Value: "nama", ValueText: "Adolf Renaldy (Test " + brand + ")"},
					{Key: "2", Value: "alamat", ValueText: "Jl. Test No. 123"},
					{Key: "3", Value: "periode", ValueText: "Agustus 2026"},
					{Key: "4", Value: "jumlah_tagihan", ValueText: "Rp 150.000"},
					{Key: "5", Value: "jatuh_tempo", ValueText: "10 Agustus 2026"},
					{Key: "6", Value: "link_pembayaran", ValueText: "https://checkout.xendit.co/web/test-" + brand},
				},
			},
		}

		resp, err := client.SendDirectBroadcast(payload)
		if err != nil {
			log.Printf("❌ Failed to send broadcast for %s: %v", brand, err)
		} else {
			log.Printf("✅ SUCCESS! Broadcast sent for %s. Response status: %s, Data: %+v", brand, resp.Status, resp.Data)
		}
		fmt.Println()
	}
}
