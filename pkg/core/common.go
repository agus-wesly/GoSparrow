package core

import (
	"context"

	"github.com/chromedp/chromedp"
)

func CreateNewContext() (context.Context, context.CancelFunc) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
	)
	actx, acancel := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, _ := chromedp.NewContext(actx)
	return ctx, acancel
}
