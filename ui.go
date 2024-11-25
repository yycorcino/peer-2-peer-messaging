package main

import (
	"fmt"
	"io"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ChatUI is a Text User Interface (TUI) for a ChatRoom.
// The Run method will draw the UI to the terminal in "fullscreen"
// mode. You can quit with Ctrl-C, or by typing "/quit" into the
// chat prompt.
type ChatUI struct {
	cr        *ChatRoom
	app       *tview.Application
	peersList *tview.TextView

	msgW    io.Writer
	inputCh chan string
	doneCh  chan struct{}
	friends []string
}

// NewChatUI returns a new ChatUI struct that controls the text UI.
// It won't actually do anything until you call Run().
func NewChatUI(cr *ChatRoom) *ChatUI {
	app := tview.NewApplication()

	// make a text view to contain our chat messages
	msgBox := tview.NewTextView()
	msgBox.SetDynamicColors(true)
	msgBox.SetBorder(true)
	msgBox.SetTitle(fmt.Sprintf("Room: %s", cr.roomName))

	// text views are io.Writers, but they don't automatically refresh.
	// this sets a change handler to force the app to redraw when we get
	// new messages to display.
	msgBox.SetChangedFunc(func() {
		app.Draw()
	})

	// an input field for typing messages into
	inputCh := make(chan string, 32)
	input := tview.NewInputField().
		SetLabel(cr.nick + " > ").
		SetFieldWidth(0).
		SetFieldBackgroundColor(tcell.ColorBlack)

	// the done func is called when the user hits enter, or tabs out of the field
	input.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			// we don't want to do anything if they just tabbed away
			return
		}
		line := input.GetText()
		if len(line) == 0 {
			// ignore blank lines
			return
		}

		// bail if requested
		if line == "/quit" {
			app.Stop()
			return
		}

		// send the line onto the input chan and reset the field text
		inputCh <- line
		input.SetText("")
	})

	// make a text view to hold the list of peers in the room, updated by ui.refreshPeers()
	peersList := tview.NewTextView()
	peersList.SetBorder(true)
	peersList.SetTitle("Peers")
	peersList.SetChangedFunc(func() { app.Draw() })

	// chatPanel is a horizontal box with messages on the left and peers on the right
	// the peers list takes 20 columns, and the messages take the remaining space
	chatPanel := tview.NewFlex().
		AddItem(msgBox, 0, 1, false).
		AddItem(peersList, 20, 1, false)

	// flex is a vertical box with the chatPanel on top and the input field at the bottom.

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(chatPanel, 0, 1, false).
		AddItem(input, 1, 1, true)

	app.SetRoot(flex, true)

	return &ChatUI{
		cr:        cr,
		app:       app,
		peersList: peersList,
		msgW:      msgBox,
		inputCh:   inputCh,
		doneCh:    make(chan struct{}, 1),
		friends:   []string{},
	}
}

// Run starts the chat event loop in the background, then starts
// the event loop for the text UI.
func (ui *ChatUI) Run() error {
	go ui.handleEvents()
	defer ui.end()

	return ui.app.Run()
}

// end signals the event loop to exit gracefully
func (ui *ChatUI) end() {
	ui.doneCh <- struct{}{}
}

// refreshPeers pulls the list of peers currently in the chat room and
// displays the last 8 chars of their peer id in the Peers panel in the ui.
func (ui *ChatUI) refreshPeers() {
	peers := ui.cr.ListPeers()

	// clear is thread-safe
	ui.peersList.Clear()

	for i, p := range peers {
		fmt.Fprintf(ui.peersList, "[%d] %s\n", i+1, p.String()) // Display peer IDs in numerical order
	}

	ui.app.Draw()
}

// displayChatMessage writes a ChatMessage from the room to the message window,
// with the sender's nick highlighted in green.
func (ui *ChatUI) displayChatMessage(cm *ChatMessage) {
	prompt := withColor("green", fmt.Sprintf("<%s>:", cm.SenderNick))
	fmt.Fprintf(ui.msgW, "%s %s\n", prompt, cm.Message)
}

// displaySelfMessage writes a message from ourselves to the message window,
// with our nick highlighted in yellow.
func (ui *ChatUI) displaySelfMessage(msg string) {
	prompt := withColor("yellow", fmt.Sprintf("<%s>:", ui.cr.nick))
	fmt.Fprintf(ui.msgW, "%s %s\n", prompt, msg)
}

// handleEvents runs an event loop that sends user input to the chat room
// and displays messages received from the chat room. It also periodically
// refreshes the list of peers in the UI.
func (ui *ChatUI) handleEvents() {
	peerRefreshTicker := time.NewTicker(time.Second)
	defer peerRefreshTicker.Stop()

	for {
		select {
		case input := <-ui.inputCh:
			if len(input) > 8 && input[:8] == "/status " {
				peerID := input[8:] // Extract Peer ID after "/status "
				isOnline := isPeerOnline(peerID, ui.cr.ListPeers())
				if isOnline {
					ui.displaySelfMessage(fmt.Sprintf("Peer %s is online!", peerID))
				} else {
					ui.displaySelfMessage(fmt.Sprintf("Peer %s is offline.", peerID))
				}
				continue
			}

			// Handle /addFriend command
			if len(input) > 10 && input[:10] == "/addFriend" {
				friendID := input[11:] // Extract Peer ID after "/addFriend "

				// Check if the peer ID is already in the list
				if contains(ui.friends, friendID) {
					ui.displaySelfMessage(fmt.Sprintf("Peer %s is already in your friends list.", friendID))
				} else {
					ui.friends = append(ui.friends, friendID) // Add to the slice
					ui.displaySelfMessage(fmt.Sprintf("Added peer %s to your friends list.", friendID))
				}
				continue
			}

			// Handle /viewFriends command
			if input == "/viewFriends" {
				if len(ui.friends) == 0 {
					ui.displaySelfMessage("Your friends list is empty.")
				} else {
					ui.displaySelfMessage("Your friends list:")
					for _, friend := range ui.friends {
						ui.displaySelfMessage(friend)
					}
				}
				continue
			}

			if input == "/dm" {
				ui.displaySelfMessage("You're prompting to direct message another user.")
				ui.displaySelfMessage("Please enter the Peer ID of the user you want to DM:")

				peer_ID := <-ui.inputCh
				ui.displaySelfMessage(fmt.Sprintf("You entered Peer ID: %s", peer_ID))

				continue
			}

			//currently not working
			// if len(input) > 7 && input[:7] == "/peers " {
			// 	peers := ui.cr.ListPeers()
			// 	if len(peers) == 0 {
			// 		ui.displaySelfMessage("No peers are currently connected.")
			// 	} else {
			// 		ui.displaySelfMessage("Connected peers:")
			// 		for _, p := range peers {
			// 			ui.displaySelfMessage(p.String())
			// 		}
			// 	}
			// }

			// when the user types in a line, publish it to the chat room and print to the message window
			err := ui.cr.Publish(input)
			if err != nil {
				printErr("publish error: %s", err)
			}
			ui.displaySelfMessage(input)

		case m := <-ui.cr.Messages:
			// when we receive a message from the chat room, print it to the message window
			ui.displayChatMessage(m)

		case <-peerRefreshTicker.C:
			// refresh the list of peers in the chat room periodically
			ui.refreshPeers()

		case <-ui.cr.ctx.Done():
			return

		case <-ui.doneCh:
			return
		}
	}
}

// withColor wraps a string with color tags for display in the messages text box.
func withColor(color, msg string) string {
	return fmt.Sprintf("[%s]%s[-]", color, msg)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
