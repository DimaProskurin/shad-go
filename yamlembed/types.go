package yamlembed

import (
	"strings"
)

type Foo struct {
	A string `yaml:"aa"`
	p int64  `yaml:"-"`
}

type Bar struct {
	I      int64    `yaml:"-"`
	B      string   `yaml:"b"`
	UpperB string   `yaml:"-"`
	OI     []string `yaml:"oi,omitempty"`
	F      []any    `yaml:"f,flow"`
}

type BarAlias Bar

func (b *Bar) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var temp BarAlias
	if err := unmarshal(&temp); err != nil {
		return err
	}
	b.I = temp.I
	b.B = temp.B
	b.UpperB = strings.ToUpper(b.B)
	b.OI = temp.OI
	b.F = temp.F
	return nil
}

type Baz struct {
	Foo `yaml:",inline"`
	Bar `yaml:",inline"`
}

func (b *Baz) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := unmarshal(&b.Foo); err != nil {
		return err
	}
	if err := unmarshal(&b.Bar); err != nil {
		return err
	}
	return nil
}
