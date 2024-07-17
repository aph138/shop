// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.680
package layout

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bytes"

import "github.com/aph138/shop/shared"

func Main(title string, user *shared.User) templ.Component {
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
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<html><head><script src=\"/public/js/htmx.min.js\"></script><script src=\"/public/js/response-targets.js\"></script><link href=\"/public/css/tailwind.css\" rel=\"stylesheet\"><title>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var2 string
		templ_7745c5c3_Var2, templ_7745c5c3_Err = templ.JoinStringErrs(title)
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `server/web/layout/main.templ`, Line: 10, Col: 22}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var2))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</title></head><body><nav class=\"bg-gray-800 p-4\"><div class=\"container mx-auto flex justify-between\"><div class=\"flex flex-1 items-center justify-center sm:items-stretch sm:justify-start\"><div class=\"flex flex-shrink-0 items-center text-white pr-4\"><a href=\"/\">SHOP</a></div><ul class=\"hidden md:flex space-x-4 mx-auto\"><li><a href=\"/\" class=\"text-gray-300 hover:text-white\">Home</a></li><li><a href=\"/item\" class=\"text-gray-300 hover:text-white\">Shop</a></li><li><a href=\"#\" class=\"text-gray-300 hover:text-white\">About</a></li></ul></div><div class=\"relative md:flex items-center space-x-4\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if user.ID != "" {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div id=\"user-info\" class=\"text-gray-300 relative\"><span>Hello, ")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var3 string
			templ_7745c5c3_Var3, templ_7745c5c3_Err = templ.JoinStringErrs(user.Username)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `server/web/layout/main.templ`, Line: 29, Col: 56}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var3))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("!</span> <button id=\"user-menu-button\" class=\"text-gray-300 hover:text-white ml-2 px-1\"><svg xmlns=\"http://www.w3.org/2000/svg\" fill=\"none\" viewBox=\"0 0 24 24\" stroke-width=\"1.5\" stroke=\"currentColor\" class=\"size-6\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" d=\"M17.982 18.725A7.488 7.488 0 0 0 12 15.75a7.488 7.488 0 0 0-5.982 2.975m11.963 0a9 9 0 1 0-11.963 0m11.963 0A8.966 8.966 0 0 1 12 21a8.966 8.966 0 0 1-5.982-2.275M15 9.75a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z\"></path></svg></button><div id=\"user-dropdown\" class=\"absolute right-0 mt-2 w-48 bg-white rounded-md shadow-lg hidden\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if user.Role != 0 {
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<a class=\"block px-4 py-2 text-gray-800 hover:bg-gray-200\" href=\"/admin/list\">User List</a> <a class=\"block px-4 py-2 text-gray-800 hover:bg-gray-200\" href=\"/admin/item\">Add Item</a> ")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<a href=\"/profile\" class=\"block px-4 py-2 text-gray-800 hover:bg-gray-200 px-1\">Edit Profile</a> <a href=\"/password\" class=\"block px-4 py-2 text-gray-800 hover:bg-gray-200\">Change Password</a> <a href=\"#\" class=\"block px-4 py-2 text-gray-800 hover:bg-gray-200\">Logout</a></div><a href=\"/cart\"><button><svg xmlns=\"http://www.w3.org/2000/svg\" fill=\"none\" viewBox=\"0 0 24 24\" stroke-width=\"1.5\" stroke=\"currentColor\" class=\"size-6\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" d=\"M15.75 10.5V6a3.75 3.75 0 1 0-7.5 0v4.5m11.356-1.993 1.263 12c.07.665-.45 1.243-1.119 1.243H4.25a1.125 1.125 0 0 1-1.12-1.243l1.264-12A1.125 1.125 0 0 1 5.513 7.5h12.974c.576 0 1.059.435 1.119 1.007ZM8.625 10.5a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Zm7.5 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Z\"></path></svg></button></a></div>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		} else {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<a href=\"/signin\"><button class=\"text-gray-300 hover:text-white\">Sign In</button></a> <a href=\"/signup\"><button class=\"text-gray-300 hover:text-white\">Sign Up</button></a>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</div><button class=\"md:hidden text-gray-300 hover:text-white\" id=\"menu-button\"><svg class=\"w-6 h-6\" fill=\"none\" stroke=\"currentColor\" viewBox=\"0 0 24 24\" xmlns=\"http://www.w3.org/2000/svg\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M4 6h16M4 12h16m-7 6h7\"></path></svg></button></div><div class=\"hidden\" id=\"menu\"><a href=\"/\" class=\"block text-gray-300 hover:text-white p-2\">Home</a> <a href=\"/item\" class=\"block text-gray-300 hover:text-white p-2\">About</a> <a href=\"#\" class=\"block text-gray-300 hover:text-white p-2\">Services</a></div></nav>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templ_7745c5c3_Var1.Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<script>\n            const menuButton = document.getElementById('menu-button');\n            const menu = document.getElementById('menu');\n            const userMenuButton = document.getElementById('user-menu-button');\n            const userDropdown = document.getElementById('user-dropdown');\n\n            menuButton.addEventListener('click', () => {\n                menu.classList.toggle('hidden');\n            });\n            userMenuButton.addEventListener('click', (e) => {\n                e.preventDefault();\n                userDropdown.classList.toggle('hidden');\n            });\n        </script></body></html>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if !templ_7745c5c3_IsBuffer {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteTo(templ_7745c5c3_W)
		}
		return templ_7745c5c3_Err
	})
}
