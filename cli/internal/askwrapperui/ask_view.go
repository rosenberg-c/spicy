//go:build gio

package askwrapperui

import (
	"context"
	"image"
	"image/color"
	"strings"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

func (u *askUI) draw(gtx layout.Context, th *material.Theme, parent context.Context) layout.Dimensions {
	paint.FillShape(gtx.Ops, color.NRGBA{R: 13, G: 17, B: 23, A: 255}, clip.Rect{Max: gtx.Constraints.Max}.Op())
	for i := range u.historyClicks {
		if !u.running && u.historyClicks[i].Clicked(gtx) {
			u.selected = i
			if u.modeFollowUp {
				u.setStatus(copyFollowUpContextPick)
			}
			u.setPreview(previewForHistory(u.history[i]))
		}
	}
	for i := 0; i < len(u.historyDelete); i++ {
		if !u.running && u.historyDelete[i].Clicked(gtx) {
			u.deleteAt(i)
			break
		}
	}
	if u.submit.Clicked(gtx) {
		if u.running {
			u.cancelAsk()
		} else {
			if u.canSubmit() {
				u.startAsk(parent)
			} else {
				u.setStatus(copyPickHistoryContext)
			}
		}
	}

	inset := layout.UniformInset(unit.Dp(12))
	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(
			gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								r := material.RadioButton(th, &u.modeEnum, modeAskValue, copyModeAskLabel)
								if u.running {
									gtx = gtx.Disabled()
								}
								return r.Layout(gtx)
							}),
							layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								r := material.RadioButton(th, &u.modeEnum, modeFollowUpValue, copyModeFollowUpLabel)
								if u.running {
									gtx = gtx.Disabled()
								}
								return r.Layout(gtx)
							}),
						)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
					layout.Rigid(material.H6(th, u.askLabel()).Layout),
					layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						fieldHeight := gtx.Dp(unit.Dp(52))
						return layout.Flex{Axis: u.inputRowAxis(), Alignment: layout.Middle}.Layout(gtx,
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								return layout.Inset{Right: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									gtx.Constraints.Min.Y = fieldHeight
									gtx.Constraints.Max.Y = fieldHeight
									inputInset := layout.UniformInset(unit.Dp(8))
									return inputInset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
										bg := color.NRGBA{R: 21, G: 28, B: 38, A: 255}
										border := color.NRGBA{R: 72, G: 90, B: 114, A: 255}
										r := gtx.Dp(unit.Dp(8))
										rect := clip.RRect{Rect: image.Rectangle{Max: gtx.Constraints.Max}, NW: r, NE: r, SW: r, SE: r}
										paint.FillShape(gtx.Ops, bg, rect.Op(gtx.Ops))
										paint.FillShape(gtx.Ops, border, clip.Stroke{Path: rect.Path(gtx.Ops), Width: float32(gtx.Dp(unit.Dp(1)))}.Op())
										return layout.UniformInset(unit.Dp(6)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
											u.question.ReadOnly = u.isQuestionReadOnly()
											ed := material.Editor(th, &u.question, copyQuestionHint)
											ed.TextSize = unit.Sp(18)
											ed.Color = color.NRGBA{R: 236, G: 241, B: 248, A: 255}
											ed.HintColor = color.NRGBA{R: 141, G: 153, B: 171, A: 255}
											ed.SelectionColor = color.NRGBA{R: 76, G: 132, B: 214, A: 120}
											return ed.Layout(gtx)
										})
									})
								})
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								label := u.primaryLabel()
								btn := material.Button(th, &u.submit, label)
								buttonGTX := gtx
								if !u.canSubmit() {
									buttonGTX = gtx.Disabled()
									btn.Background = color.NRGBA{R: 85, G: 85, B: 85, A: 255}
								} else if u.running {
									btn.Background = color.NRGBA{R: 167, G: 63, B: 63, A: 255}
								}
								return layout.Inset{Top: unit.Dp(8), Bottom: unit.Dp(8)}.Layout(buttonGTX, btn.Layout)
							}),
						)
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Flexed(0.34, func(gtx layout.Context) layout.Dimensions {
						listGTX := gtx
						if u.isHistoryDisabled() {
							listGTX = gtx.Disabled()
						}
						pad := layout.UniformInset(unit.Dp(10))
						return pad.Layout(listGTX, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
								layout.Rigid(material.H6(th, copyHistoryTitle).Layout),
								layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
									if len(u.history) == 0 {
										return material.Body1(th, copyNoHistory).Layout(gtx)
									}
									listStyle := material.List(th, &u.historyList)
									listStyle.AnchorStrategy = material.Occupy
									return listStyle.Layout(gtx, len(u.history), func(gtx layout.Context, i int) layout.Dimensions {
										title := strings.TrimSpace(u.history[i].Question)
										if title == "" {
											title = copyEmptyQuestion
										}
										if len(title) > 72 {
											title = title[:72] + "..."
										}
										return layout.Inset{Bottom: unit.Dp(6)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
											return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
												layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
													return u.historyClicks[i].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
														rowHeight := gtx.Dp(unit.Dp(36))
														if gtx.Constraints.Min.Y < rowHeight {
															gtx.Constraints.Min.Y = rowHeight
														}
														if gtx.Constraints.Max.Y < rowHeight {
															gtx.Constraints.Max.Y = rowHeight
														}
														bg := color.NRGBA{R: 35, G: 43, B: 56, A: 255}
														fg := color.NRGBA{R: 224, G: 232, B: 245, A: 255}
														if i == u.selected {
															bg = color.NRGBA{R: 44, G: 95, B: 181, A: 255}
															fg = color.NRGBA{R: 248, G: 251, B: 255, A: 255}
														} else if gtx.Focused(&u.historyClicks[i]) {
															bg = color.NRGBA{R: 72, G: 88, B: 112, A: 255}
														}
														paint.FillShape(gtx.Ops, bg, clip.RRect{Rect: image.Rectangle{Max: gtx.Constraints.Max}, NW: 6, NE: 6, SW: 6, SE: 6}.Op(gtx.Ops))
														return layout.Inset{Left: unit.Dp(10), Right: unit.Dp(10), Top: unit.Dp(8), Bottom: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
															lbl := material.Body1(th, title)
															lbl.Color = fg
															lbl.MaxLines = 1
															return lbl.Layout(gtx)
														})
													})
												}),
												layout.Rigid(layout.Spacer{Width: unit.Dp(6)}.Layout),
												layout.Rigid(func(gtx layout.Context) layout.Dimensions {
													del := material.Button(th, &u.historyDelete[i], copyRowDelete)
													del.Background = color.NRGBA{R: 122, G: 53, B: 53, A: 255}
													del.Color = color.NRGBA{R: 255, G: 245, B: 245, A: 255}
													if u.running {
														del.Background = color.NRGBA{R: 85, G: 85, B: 85, A: 255}
													}
													return del.Layout(gtx)
												}),
											)
										})
									})
								}),
							)
						})
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
					layout.Flexed(0.66, func(gtx layout.Context) layout.Dimensions {
						pad := layout.UniformInset(unit.Dp(10))
						return pad.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
								layout.Rigid(material.H6(th, copyPreviewTitle).Layout),
								layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								ed := material.Editor(th, &u.preview, "")
								ed.TextSize = unit.Sp(15)
								ed.Color = color.NRGBA{R: 224, G: 232, B: 245, A: 255}
								ed.SelectionColor = color.NRGBA{R: 76, G: 132, B: 214, A: 120}
								return ed.Layout(gtx)
							}),
							)
						})
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Body2(th, u.statusLine())
				lbl.Font.Typeface = "Go Mono"
				lbl.Color = color.NRGBA{R: 150, G: 163, B: 181, A: 255}
				return lbl.Layout(gtx)
			}),
		)
	})
}
