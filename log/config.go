package log

type Config struct {
	Level   string `json:"level" yaml:"level" default:"info" validate:"required,oneof=debug warn info error"`
	Format  string `json:"format" yaml:"format" default:"json" validate:"required,oneof=json console"`
	Colored bool   `json:"colored" yaml:"colored" default:"false"`
}
