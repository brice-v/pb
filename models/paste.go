package models

import "fmt"

type Paste struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

func (p *Paste) Validate() error {
	if p.Title == "" {
		return fmt.Errorf("paste title cannot be empty string")
	}
	if p.Text == "" {
		return fmt.Errorf("paste text cannot be empty string")
	}
	return nil
}
