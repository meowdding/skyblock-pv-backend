package utils

import (
	"encoding/json"
	"fmt"
	"skyblock-pv-backend/utils/nbt"
)

type Item struct {
	*nbt.Compound
}

type PetData struct {
	Type        string  `json:"type"`
	Exp         float64 `json:"exp"`
	Tier        string  `json:"tier"`
	HideInfo    bool    `json:"hideInfo"`
	CandiesUsed int16   `json:"candiesUsed"`
}

func (item Item) GetExtrAttributes() *nbt.Compound {
	if !item.Contains("tag") {
		return nil
	}
	tag := item.Get("tag").AsCompound()
	if !tag.Contains("ExtraAttributes") {
		return nil
	}
	return tag.Get("ExtraAttributes").AsCompound()
}

func (item Item) GetPetData() *PetData {
	tag := item.GetExtrAttributes()
	if tag == nil {
		return nil
	}
	petJson := tag.Get("petInfo").AsString()
	var petData PetData
	err := json.Unmarshal([]byte(petJson), &petData)
	if err != nil {
		return nil
	}
	return &petData
}

func (item Item) GetSbId() *string {
	tag := item.GetExtrAttributes()
	if tag == nil {
		return nil
	}
	data := tag.Get("id").AsString()
	if data == "PET" {
		petData := item.GetPetData()
		if petData == nil {
			return nil
		}
		data = fmt.Sprintf("pet:%s:%s", petData.Type, petData.Tier)
	}
	return &data
}
