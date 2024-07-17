// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.680
package userview

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bytes"

import "github.com/aph138/shop/server/web/layout"
import "github.com/aph138/shop/shared"

func Password(u *shared.User) templ.Component {
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
		templ_7745c5c3_Var2 := templ.ComponentFunc(func(ctx context.Context, templ_7745c5c3_W io.Writer) (templ_7745c5c3_Err error) {
			templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templ_7745c5c3_W.(*bytes.Buffer)
			if !templ_7745c5c3_IsBuffer {
				templ_7745c5c3_Buffer = templ.GetBuffer()
				defer templ.ReleaseBuffer(templ_7745c5c3_Buffer)
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div hx-ext=\"response-targets\" class=\"bg-gray-100 flex min-h-screen justify-center\"><form id=\"form\" class=\"pt-6 flex flex-col items-center\"><div class=\"mb-4\"><label class=\"block text-gray-700 text-sm font-bold mb-2\" for=\"old\">Old Password</label> <input id=\"old\" type=\"password\" name=\"oldPassword\" placeholder=\"*****\" class=\"shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight\"></div><div class=\"mb-4\"><label class=\"block text-gray-700 text-sm font-bold mb-2\" for=\"new\">New Password</label> <input id=\"new\" type=\"password\" name=\"newPassword\" placeholder=\"*****\" class=\"shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight\"></div><div class=\"mb-4\"><label class=\"block text-gray-700 text-sm font-bold mb-2\" for=\"confirm\">Confirm New Password</label> <input id=\"confirm\" type=\"password\" name=\"confirmPassword\" placeholder=\"*****\" class=\"shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight\"></div><div class=\"text-black-500\" id=\"result\"></div><div class=\"text-red-500\" id=\"error\"></div><button class=\"bg-blue-500 rounded p-2 w-full\" hx-swap=\"innerHTML\" type=\"button\" hx-put=\"/password\" hx-include=\"#form\" hx-target-4*=\"#error\" hx-target=\"#result\" hx-on:click=\"clear()\">Change</button></form></div><script>\n        function clear(){\n            document.getElementById(\"result\").innerHTML=\"\";\n            document.getElementById(\"error\").innerHTML=\"\";\n        }\n        </script>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if !templ_7745c5c3_IsBuffer {
				_, templ_7745c5c3_Err = io.Copy(templ_7745c5c3_W, templ_7745c5c3_Buffer)
			}
			return templ_7745c5c3_Err
		})
		templ_7745c5c3_Err = layout.Main("Change Password", u).Render(templ.WithChildren(ctx, templ_7745c5c3_Var2), templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if !templ_7745c5c3_IsBuffer {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteTo(templ_7745c5c3_W)
		}
		return templ_7745c5c3_Err
	})
}
