// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.680
package stockview

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bytes"

import "github.com/aph138/shop/server/web/layout"
import "github.com/aph138/shop/shared"

func AddItem(u *shared.User) templ.Component {
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
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div class=\"container mx-auto p-8\"><form id=\"form\" hx-ext=\"response-targets\" hx-encoding=\"multipart/form-data\"><div class=\"mb-4\"><label class=\"block text-gray-700 text-sm mb-1\" for=\"name\">Name</label> <input type=\"text\" id=\"name\" name=\"name\" placeholder=\"name\" class=\"shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight\"></div><div class=\"mb-4\"><label class=\"block text-gray-700 text-sm mb-1\" for=\"link\">Link</label> <input type=\"text\" id=\"link\" name=\"link\" placeholder=\"link (default will be the name)\" class=\"shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight\"></div><div class=\"mb-4\"><lable class=\"block text-gray-700 text-sm mb-1\" for=\"des\">Desciption</lable> <textarea id=\"des\" name=\"des\" rows=\"4\" class=\"shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline\"></textarea></div><div class=\"mb-4\"><lable class=\"block text-gray-700 text-sm mb-1\" for=\"price\">Price</lable> <input type=\"number\" step=\"0.1\" id=\"price\" name=\"price\" class=\"shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline\"></div><div class=\"mb-4\"><lable class=\"block text-gray-700 text-sm mb-1\" for=\"numver\">Number</lable> <input type=\"number\" id=\"number\" name=\"number\" class=\"shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline\"></div><div class=\"mb-4\"><lable class=\"block text-gray-700 text-sm mb-1\" for=\"poster\">Poster</lable> <input type=\"file\" name=\"poster\" id=\"poster\" class=\"shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline\"></div><div class=\"mb-4\"><lable class=\"block text-gray-700 text-sm mb-1\" for=\"photos\">Photos</lable> <input type=\"file\" name=\"photos\" id=\"photos\" multiple class=\"shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline\"></div><button class=\"bg-blue-500 hover:bg-blue-700 rounded px-2 py-1 w-full\" hx-post=\"/admin/item\" hx-include=\"#form\" hx-target=\"#result\" hx-target-error=\"#result\">Add </button><div id=\"result\"></div></form></div>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if !templ_7745c5c3_IsBuffer {
				_, templ_7745c5c3_Err = io.Copy(templ_7745c5c3_W, templ_7745c5c3_Buffer)
			}
			return templ_7745c5c3_Err
		})
		templ_7745c5c3_Err = layout.Main("Add Item", u).Render(templ.WithChildren(ctx, templ_7745c5c3_Var2), templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if !templ_7745c5c3_IsBuffer {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteTo(templ_7745c5c3_W)
		}
		return templ_7745c5c3_Err
	})
}
