package word

import "image/color"

var (
	defaultColor = color.RGBA{100, 200, 200, 255} // Gopher Blue
	activeColor  = color.RGBA{18, 252, 10, 255}   // Green
)

// Word represents the word that will be displayed on screen
type Word struct {
	Text   string
	Color  color.Color
	Active bool
	X      int
	Y      int
}

// NewWord creates a new word
func NewWord(text string, x int) Word {
	return Word{
		Text:   text,
		Color:  defaultColor,
		Active: false,
		X:      x,
		Y:      0,
	}
}

// Runes returns the word text value as Runes
func (w *Word) Runes() []rune {
	return []rune(w.Text)
}

// SetActiveStatus sets the Active status of the Word and changes the color
func (w *Word) SetActiveStatus(status bool) {
	if status == true {
		w.Color = activeColor
	} else {
		w.Color = defaultColor
	}

	w.Active = status
}

// IncrementY increments the Y position of the word by the value passed
func (w *Word) IncrementY(value int) {
	w.Y = w.Y + value
}

// UpdateText updates the Text of the word
func (w *Word) UpdateText(text string) {
	w.Text = text
}
