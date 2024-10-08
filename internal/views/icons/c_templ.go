// Code generated by templ - DO NOT EDIT.

// templ: version: 0.2.476
package icons

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bytes"

func C() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, templ_7745c5c3_W io.Writer) (templ_7745c5c3_Err error) {
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templ_7745c5c3_W.(*bytes.Buffer)
		if !templ_7745c5c3_IsBuffer {
			templ_7745c5c3_Buffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templ_7745c5c3_Buffer)
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div class=\"brand-icon-block\" title=\"C\"><svg class=\"w-12\" xmlns=\"http://www.w3.org/2000/svg\" viewBox=\"0 0 24 24\"><path fill=\"currentColor\" d=\"m15.45 15.97l.42 2.44c-.26.14-.68.27-1.24.39c-.57.13-1.24.2-2.01.2c-2.21-.04-3.87-.7-4.98-1.96c-1.14-1.27-1.68-2.88-1.68-4.83C6 9.9 6.68 8.13 8 6.89C9.28 5.64 10.92 5 12.9 5c.75 0 1.4.07 1.94.19s.94.25 1.2.4l-.6 2.49l-1.04-.34c-.4-.1-.87-.15-1.4-.15c-1.15-.01-2.11.36-2.86 1.1c-.76.73-1.14 1.85-1.18 3.34c.01 1.36.37 2.42 1.08 3.2c.71.77 1.7 1.17 2.99 1.18l1.33-.12c.43-.08.79-.19 1.09-.32\"></path></svg></div>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if !templ_7745c5c3_IsBuffer {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteTo(templ_7745c5c3_W)
		}
		return templ_7745c5c3_Err
	})
}
