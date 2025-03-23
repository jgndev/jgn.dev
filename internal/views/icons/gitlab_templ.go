// Code generated by templ - DO NOT EDIT.

// templ: version: 0.2.476
package icons

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bytes"

func GitLab() templ.Component {
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
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div class=\"brand-icon-block\" title=\"GitLab\"><svg class=\"w-12\" xmlns=\"http://www.w3.org/2000/svg\" viewBox=\"0 0 24 24\"><path fill=\"currentColor\" d=\"m12 21.42l3.684-11.333H8.32z\"></path> <path fill=\"currentColor\" d=\"m3.16 10.087l-1.123 3.444a.76.76 0 0 0 .277.852l9.685 7.038z\" opacity=\".25\"></path> <path fill=\"currentColor\" d=\"M3.16 10.087h5.16L6.1 3.262a.383.383 0 0 0-.728 0z\"></path> <path fill=\"currentColor\" d=\"m20.845 10.087l1.118 3.444a.76.76 0 0 1-.276.852l-9.688 7.038z\" opacity=\".25\"></path> <path fill=\"currentColor\" d=\"M20.845 10.087h-5.161L17.9 3.262a.383.383 0 0 1 .727 0l2.217 6.825Z\"></path> <path fill=\"currentColor\" d=\"m11.999 21.421l3.685-11.334h5.161zm0 0l-8.84-11.334H8.32z\" opacity=\".5\"></path></svg></div>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if !templ_7745c5c3_IsBuffer {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteTo(templ_7745c5c3_W)
		}
		return templ_7745c5c3_Err
	})
}
