package game

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"math"
	"math/rand"
	"strings"
	"time"

	"personal_projects/ebiten/typing/assets/fonts"
	"personal_projects/ebiten/typing/assets/images"
	"personal_projects/ebiten/typing/assets/words"
	"personal_projects/ebiten/typing/pkg/word"

	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/inpututil"

	"github.com/hajimehoshi/ebiten/text"
	"golang.org/x/image/font"
)

// state represents the different game states such as Title, Game, and Game Over
type state int

const (
	titleFontSize   = 18
	fontSize        = 14
	fontDpi         = 72
	defaultLives    = 5
	defaultFallRate = 1 // Defines the rate of which words fall in pixels
	defaultSpeed    = 2 // Defines new words per second

	stateTitle state = iota
	stateGame
	stateGameOver
)

var (
	fontColor = color.RGBA{100, 200, 200, 255} // Gopher Blue
)

// Game contains all properties and logic that is required for the Falling Words game
type Game struct {
	background *ebiten.Image

	state         state
	screenWidth   int
	screenHeight  int
	titleFontFace font.Face
	fontFace      font.Face
	gameStats     gameStats

	wordsList         map[string]string
	activeWord        string
	wordsOnScreen     map[string]word.Word
	secondLastDropped float64

	lives     int
	dropSpeed int
	fallRate  int
}

// gameStats contains all properties related to game statistics
type gameStats struct {
	startTime      time.Time
	timeElapsed    time.Duration
	wordsCompleted int
}

// NewGame creates an new instance of the Game with all default values
func NewGame(screenWidth int, screenHeight int) (Game, error) {
	tt, err := truetype.Parse(fonts.PressStart2P_ttf)
	if err != nil {
		return Game{}, err
	}
	fontFace := truetype.NewFace(tt, &truetype.Options{
		Size:    fontSize,
		DPI:     fontDpi,
		Hinting: font.HintingFull,
	})
	titleFontFace := truetype.NewFace(tt, &truetype.Options{
		Size:    titleFontSize,
		DPI:     fontDpi,
		Hinting: font.HintingFull,
	})

	// Load word list
	wordsList, err := LoadWordList(words.FiveLetterWords)
	if err != nil {
		return Game{}, err
	}

	// Load background image
	bgImage, _, err := image.Decode(bytes.NewReader(images.Background))
	if err != nil {
		return Game{}, err
	}
	bgEbitenImage, err := ebiten.NewImageFromImage(bgImage, ebiten.FilterDefault)
	if err != nil {
		return Game{}, err
	}

	return Game{
		background:    bgEbitenImage,
		state:         stateTitle,
		wordsList:     wordsList,
		wordsOnScreen: map[string]word.Word{},
		gameStats: gameStats{
			startTime: time.Now(),
		},
		screenWidth:   screenWidth,
		screenHeight:  screenHeight,
		titleFontFace: titleFontFace,
		fontFace:      fontFace,
		lives:         defaultLives,
		dropSpeed:     defaultSpeed,
		fallRate:      defaultFallRate,
	}, nil
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.screenWidth, g.screenHeight
}

// Update contains the games main logic
func (g *Game) Update(screen *ebiten.Image) error {
	switch g.state {
	case stateTitle:
		// Game start! Wait for user input to start the game
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.state = stateGame
		}
	case stateGame:
		// Iterate through each keyboard key, was it pressed?
		for k := ebiten.Key(0); k <= ebiten.KeyMax; k++ {
			if inpututil.IsKeyJustPressed(k) {

				if g.activeWord != "" {
					// Ensure player completes the current active word
					currentActiveWord := g.wordsOnScreen[g.activeWord]

					textRunes := []rune(currentActiveWord.Text)
					if strings.EqualFold(string(textRunes[0]), k.String()) {
						currentActiveWord.UpdateText(string(textRunes[1:]))
					}

					if currentActiveWord.Text == "" {
						// Word finished, increase score and reset activeWord
						g.removeWord(g.activeWord)
						g.activeWord = ""

						g.gameStats.wordsCompleted++

						if len(g.wordsList) == 0 {
							// The player finished all the words we have for them!
							g.gameStats.timeElapsed = time.Since(g.gameStats.startTime)
							g.state = stateGameOver
							break
						}
					} else {
						g.wordsOnScreen[g.activeWord] = currentActiveWord
					}
				} else {
					// No active key, find a word to mark as active

					// Ensure if there are more than one word with the same first letter, that the one at the bottom gets
					// marked as active first
					potentialActiveWords := []string{}
					for key := range g.wordsOnScreen {
						word := g.wordsOnScreen[key]

						textRunes := word.Runes()
						if strings.EqualFold(string(textRunes[0]), k.String()) {
							potentialActiveWords = append(potentialActiveWords, word.Text)
						}
					}

					activeWord := word.Word{}
					for _, word := range potentialActiveWords {
						wordOnScreen := g.wordsOnScreen[word]
						if wordOnScreen.Y > activeWord.Y {
							activeWord = wordOnScreen
						}
					}

					if activeWord.Text != "" {
						g.activeWord = activeWord.Text

						textRunes := activeWord.Runes()
						activeWord.UpdateText(string(textRunes[1:]))
						activeWord.SetActiveStatus(true)

						g.wordsOnScreen[g.activeWord] = activeWord
					}
				}
			}
		}

		// Iterates through all words on the screen and increments the Y.
		// If word reaches the bottom of the screen, remove it from the game
		for key := range g.wordsOnScreen {
			word := g.wordsOnScreen[key]
			word.IncrementY(g.fallRate)
			g.wordsOnScreen[key] = word

			if g.wordHasReachedBottom(&word) {
				g.removeWord(key)
				if g.activeWord == key {
					g.activeWord = ""
				}

				g.lives--
				if g.lives == 0 {
					g.gameStats.timeElapsed = time.Since(g.gameStats.startTime)
					g.state = stateGameOver
					break
				}
			}
		}

		// If we're still in game state, drop a new word every X seconds
		if g.state != stateGameOver {
			elapsedSeconds := time.Since(g.gameStats.startTime).Seconds()
			if int(elapsedSeconds)%g.dropSpeed == 0 && int(g.secondLastDropped) != int(elapsedSeconds) {
				g.secondLastDropped = elapsedSeconds

				for newWord := range g.wordsList {
					if _, ok := g.wordsOnScreen[newWord]; !ok {
						newWordWidth := font.MeasureString(g.fontFace, newWord)
						g.wordsOnScreen[newWord] = word.NewWord(newWord, rand.Intn(g.screenWidth-newWordWidth.Ceil()))
						break
					}
				}
			}
		}

		// Increase the difficult if certain conditions are met
		// TODO: This is for demonstration purposes only. So it only increases the difficulty once
		// would be better to implement a more robust difficulty management system
		if g.gameStats.wordsCompleted >= 35 && g.dropSpeed == defaultSpeed {
			g.dropSpeed--
		}
	case stateGameOver:
		// Game over! Wait for user input to reset the game
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			err := g.resetGame()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Draw draws the screen based on the state of the game (this runs immediately after the Update)
func (g *Game) Draw(screen *ebiten.Image) {
	// Draw the background image
	options := &ebiten.DrawImageOptions{}
	bgWidth, bgHeight := g.background.Size()
	xScale := float64(g.screenWidth) / float64(bgWidth)
	yScale := float64(g.screenHeight) / float64(bgHeight)
	options.GeoM.Scale(xScale, yScale)
	screen.DrawImage(g.background, options)

	switch g.state {
	case stateTitle:
		titleLineTexts := []string{"FALLING WORDS", "", "", "", "", "", "", "", "", "", "", "PRESS SPACE KEY TO START"}
		for index, lineText := range titleLineTexts {
			x := (g.screenWidth - len(lineText)*titleFontSize) / 2
			y := (g.screenHeight / 2) + (index * titleFontSize)
			drawTextWithShadow(screen, lineText, g.titleFontFace, x, y, fontColor, color.Black)
		}
	case stateGame:
		// Draw Falling Words
		for _, word := range g.wordsOnScreen {
			drawTextWithShadow(screen, word.Text, g.fontFace, word.X, word.Y, word.Color, color.Black)
		}

		// Draw Stats Bar
		wordsCompletedText := fmt.Sprintf("WORDS: %d", g.gameStats.wordsCompleted)
		livesText := fmt.Sprintf("LIVES: %d", g.lives)
		drawTextWithShadow(screen, wordsCompletedText, g.fontFace, 0, fontSize, fontColor, color.Black)
		drawTextWithShadow(screen, livesText, g.fontFace, (g.screenWidth - len(livesText)*fontSize), fontSize, fontColor, color.Black)
	case stateGameOver:
		// Display Game Stats
		wordsPerMinute := roundTwoDecimalPlace(float64(g.gameStats.wordsCompleted) / g.gameStats.timeElapsed.Minutes())
		timeElapsedMinutes := roundTwoDecimalPlace(g.gameStats.timeElapsed.Minutes())
		gameOverTexts := []string{"GAME OVER", "", "THANKS FOR PLAYING!", "", "", fmt.Sprintf("PLAYTIME (MINS): %v", timeElapsedMinutes), "",
			fmt.Sprintf("WORDS COMPLETED: %v", g.gameStats.wordsCompleted), "", fmt.Sprintf("WORDS PER MINUTE: %v", wordsPerMinute),
			"", "", "", "", "PRESS SPACE KEY TO CONTINUE"}

		for index, lineText := range gameOverTexts {
			x := (g.screenWidth - len(lineText)*titleFontSize) / 2
			y := (index + 4) * titleFontSize
			drawTextWithShadow(screen, lineText, g.titleFontFace, x, y, fontColor, color.Black)
		}
	}
}

// removeWord removes a word from both the screen, and the words list
func (g *Game) removeWord(wordText string) {
	delete(g.wordsOnScreen, wordText)
	delete(g.wordsList, wordText)
}

// wordHasReachedBottom determines whether a word has hit the bottom of the screen
func (g *Game) wordHasReachedBottom(word *word.Word) bool {
	if word.Y > g.screenHeight {
		return true
	}

	return false
}

// resetGame resets the game state
func (g *Game) resetGame() error {
	// Start a new game
	wordsList, err := LoadWordList(words.FiveLetterWords)
	if err != nil {
		return err
	}

	g.state = stateTitle
	g.lives = defaultLives
	g.wordsList = wordsList
	g.wordsOnScreen = map[string]word.Word{}
	g.gameStats = gameStats{
		startTime: time.Now(),
	}

	return nil
}

// drawTextWithShadow draws the text with a single pixel offset shadow
func drawTextWithShadow(screen *ebiten.Image, str string, face font.Face, x, y int, clr color.Color, clrShadow color.Color) {
	text.Draw(screen, str, face, x+1, y+1, clrShadow)
	text.Draw(screen, str, face, x, y, clr)
}

// roundTwoDecimalPlace rounds all float64s to two decimal places
func roundTwoDecimalPlace(number float64) float64 {
	return math.Round(number*100) / 100
}
