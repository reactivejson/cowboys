package domain

type PlayerConfig struct {
	RedisAddr  string `envconfig:"REDIS_ADDR"           required:"false" default:"redis:6379"`
	MasterAddr string `envconfig:"MASTER_ADDR"          required:"false" default:"http://master:8080"`
	Name       string `envconfig:"NAME"                 required:"true"`
	Health     int    `envconfig:"HEALTH"               required:"false" default:"10"`
	Damage     int    `envconfig:"DAMAGE"               required:"false" default:"1"`
}

type Player struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Health int    `json:"health"`
	Damage int    `json:"damage"`
}

func (c *Player) IsEmpty() bool {
	return Player{} == *c
}

type Round struct {
	Players map[string]*Player
}

type Action struct {
	Src  string `json:"from"`
	Dest string `json:"to"`
}
