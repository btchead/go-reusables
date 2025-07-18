package log

type Config struct {
	Level   string `json:"level" default:"info" validate:"required,oneof=debug warn info error"`
	Format  string `json:"format" default:"json" validate:"required,oneof=json console"`
	Colored bool   `json:"colored" default:"false"`
}
