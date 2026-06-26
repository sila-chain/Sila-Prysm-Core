package structs

type GetSilaDepositResponse struct {
	Data *SilaDepositData `json:"data"`
}

type SilaDepositData struct {
	ChainId string `json:"chain_id"`
	Address string `json:"address"`
}

type GetForkScheduleResponse struct {
	Data []*Fork `json:"data"`
}

type GetSpecResponse struct {
	Data any `json:"data"`
}
