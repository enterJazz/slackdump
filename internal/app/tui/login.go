package tui

import (
	"fmt"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/pkg/browser"
	"github.com/rivo/tview"
	"github.com/rusq/slackdump/v2/internal/app"
)

func (ui *UI) makeScrLogin(creds app.SlackCreds) func() (string, tview.Primitive) {
	return func() (string, tview.Primitive) {
		loginScreens := []func(creds app.SlackCreds) (string, tview.Primitive){
			ui.scrEzLogin,
			ui.scrCookie,
		}

		pages := tview.NewPages()

		info := tview.NewTextView().
			SetDynamicColors(true).
			SetRegions(true).
			SetWrap(false).
			SetTextAlign(tview.AlignCenter).
			SetHighlightedFunc(func(added, removed, remaining []string) {
				pages.SwitchToPage(added[0])
			})
		info.SetBorder(true)

		flex := tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(ui.makeLogo(), 7, 0, false).
			AddItem(info, 3, 0, false).
			AddItem(pages, 0, 1, true)

		// switchFn := func(event *tcell.EventKey) *tcell.EventKey {
		// 	fmt.Fprintf(info, "Mod: %d, Key: %d, Rune: %d\n", event.Modifiers(), event.Key(), event.Rune())
		// 	return event
		// }

		switchFn := func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyRune {
				if event.Modifiers()&tcell.ModAlt > 0 {
					page := int(event.Rune() - rune('1'))
					if page < len(loginScreens) {
						info.Highlight(strconv.Itoa(page)).ScrollToHighlight()
						return nil
					}
				}
			}
			return event
		}
		// type capturer interface {
		// 	SetInputCapture(capture func(event *tcell.EventKey) *tcell.EventKey) *tview.Box
		// }
		ui.pages.SetInputCapture(switchFn)

		for i, screen := range loginScreens {
			title, primitive := screen(creds)
			// if p, ok := primitive.(capturer); ok {
			// 	p.SetInputCapture(switchFn)
			// }
			pages.AddPage(strconv.Itoa(i), primitive, true, i == 0)
			fmt.Fprintf(info, ui.colorize(`%d  ["%d"][$itc:$pbc] %s [-:-][""]  `), i+1, i, title)
			if i == 0 {
				info.Highlight(strconv.Itoa(i)).ScrollToHighlight()
			}
		}

		return "login", flex
	}
}

func (ui *UI) scrEzLogin(creds app.SlackCreds) (title string, content tview.Primitive) {
	items := []string{
		"Enter the Slack workspace name OR\n      paste the URL of your Slack workspace.",
		"Press [$ptc]ENTER[-], the browser will open.",
		"Login as usual (browser will close automatically).",
	}

	input := ui.newLoginInputField("Slack Workspace ")
	instrFlex := ui.modal(ui.makeInstructions(items), 60, 6)
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(ui.modal(input, 60, 3), 5, 1, true).
		AddItem(instrFlex, 6, 0, false)

	return "EZ-Login", flex
}

func (ui *UI) scrCookie(creds app.SlackCreds) (string, tview.Primitive) {
	var (
		token  = ui.newLoginInputField("Token (xoxc-)").SetText(creds.Token)
		cookie = ui.newLoginInputField("Cookie (xoxd-)").SetText(creds.Cookie)
	)

	form := tview.NewForm().
		AddFormItem(token).
		AddFormItem(cookie)

	instr := ui.makeInstructions([]string{
		`Follow the steps on [$ptc::u]["url"]this page[""][-::-]`,
		"Enter the values in the fields [::i]instead[::-] of writing text file.",
		"Press [$ptc]ENTER[-] to login",
	})
	instr.SetHighlightedFunc(func(added, removed, remaining []string) {
		if len(added) > 0 && added[0] == "url" {
			go browser.OpenURL("https://github.com/rusq/slackdump/blob/master/doc/login-manual.rst")
		}
	})

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(ui.modal(form, 76, 5), 5, 0, true).
		AddItem(ui.modal(instr, 70, 5), 6, 0, false)

	return "Cookie", flex
}

func (ui *UI) newLoginInputField(text string) *tview.InputField {
	input := tview.NewInputField().
		SetLabel(text + " ")
	return input
}
