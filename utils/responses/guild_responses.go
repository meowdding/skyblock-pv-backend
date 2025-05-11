package responses

type GuildResponse struct {
	Success bool      `json:"success"`
	Guild   GuildData `json:"guild"`
}

type GuildData struct {
	Name    string            `json:"name"`
	Members []GuildMemberData `json:"members"`
}

type GuildMemberData struct {
	Uuid string `json:"uuid"`
}
