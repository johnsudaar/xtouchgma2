package gui

import (
	"fyne.io/fyne/widget"
)

type EncoderTab struct {
	attributes []*widget.Entry
	updateBtn  *widget.Button
	gui        *GUI
}

func NewEncoderTab(gui *GUI) *EncoderTab {
	tab := EncoderTab{
		attributes: make([]*widget.Entry, 8),
		gui:        gui,
	}
	for i := 0; i < 8; i++ {
		tab.attributes[i] = widget.NewEntry()
	}

	tab.updateBtn = widget.NewButton("Update", tab.onUpdate)
	return &tab
}

func (g *EncoderTab) getTabItem() *widget.TabItem {
	return widget.NewTabItem(
		"Encoders",
		widget.NewVBox(
			widget.NewForm(
				widget.NewFormItem("Encoder 1 attribute: ", g.attributes[0]),
				widget.NewFormItem("Encoder 2 attribute: ", g.attributes[1]),
				widget.NewFormItem("Encoder 3 attribute: ", g.attributes[2]),
				widget.NewFormItem("Encoder 4 attribute: ", g.attributes[3]),
				widget.NewFormItem("Encoder 5 attribute: ", g.attributes[4]),
				widget.NewFormItem("Encoder 6 attribute: ", g.attributes[5]),
				widget.NewFormItem("Encoder 7 attribute: ", g.attributes[6]),
				widget.NewFormItem("Encoder 8 attribute: ", g.attributes[7]),
			),
			g.updateBtn,
		),
	)
}

func (t *EncoderTab) onUpdate() {
	t.gui.saveSettings()
	t.updateEncoders()
}

func (t *EncoderTab) updateEncoders() {
	t.gui.link.SetEncoderAttributes([8]string{
		t.attributes[0].Text,
		t.attributes[1].Text,
		t.attributes[2].Text,
		t.attributes[3].Text,
		t.attributes[4].Text,
		t.attributes[5].Text,
		t.attributes[6].Text,
		t.attributes[7].Text,
	})
}
