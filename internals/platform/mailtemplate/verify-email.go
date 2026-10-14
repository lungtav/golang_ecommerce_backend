package mailtemplate

import (
	"bytes"
	"html/template"
	"log/slog"
)

const verifyEmailHtml = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta name="color-scheme" content="light">
  <title>Verify your email</title>
</head>
<body style="margin:0;padding:0;background-color:#fafafa;-webkit-font-smoothing:antialiased;">

  <div style="display:none;max-height:0;overflow:hidden;opacity:0;color:#fafafa;">
    Confirm your email to finish setting up your account.
  </div>

  <table width="100%" cellpadding="0" cellspacing="0" role="presentation" style="background-color:#fafafa;">
    <tr>
      <td align="center" style="padding:48px 16px;">

        <table width="520" cellpadding="0" cellspacing="0" role="presentation" style="width:100%;max-width:520px;">

          <tr>
            <td style="padding:0 4px 24px 4px;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Helvetica,Arial,sans-serif;font-size:18px;font-weight:700;letter-spacing:-0.3px;color:#0a0a0a;">
              Ecommerce in Go
            </td>
          </tr>

          <tr>
            <td style="background-color:#ffffff;border:1px solid #ececec;border-radius:12px;padding:44px 40px;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Helvetica,Arial,sans-serif;">

              <h1 style="margin:0 0 16px 0;font-size:26px;line-height:32px;font-weight:600;letter-spacing:-0.6px;color:#0a0a0a;">
                Confirm your email address
              </h1>

              <p style="margin:0 0 12px 0;font-size:15px;line-height:24px;color:#525252;">
                Hi {{.Name}},
              </p>
              <p style="margin:0 0 32px 0;font-size:15px;line-height:24px;color:#525252;">
                Welcome aboard. Verify your email to activate your account and start shopping.
              </p>

              <table cellpadding="0" cellspacing="0" role="presentation">
                <tr>
                  <td style="background-color:#0a0a0a;border-radius:8px;">
                    <a href="{{.Url}}" style="display:inline-block;padding:13px 26px;font-size:15px;font-weight:500;color:#ffffff;text-decoration:none;border-radius:8px;">
                      Verify email
                    </a>
                  </td>
                </tr>
              </table>

              <p style="margin:32px 0 0 0;font-size:13px;line-height:20px;color:#737373;">
                This link expires in {{.ExpiresIn}}. If you didn't create an account, you can safely ignore this email.
              </p>

              <hr style="border:none;border-top:1px solid #f0f0f0;margin:32px 0 24px 0;">

              <p style="margin:0 0 6px 0;font-size:12px;line-height:18px;color:#a3a3a3;">
                Button not working? Paste this link into your browser:
              </p>
              <p style="margin:0;font-size:12px;line-height:18px;word-break:break-all;">
                <a href="{{.Url}}" style="color:#525252;text-decoration:underline;">{{.Url}}</a>
              </p>

            </td>
          </tr>

          <tr>
            <td style="padding:24px 4px 0 4px;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Helvetica,Arial,sans-serif;font-size:12px;line-height:18px;color:#a3a3a3;">
              You're receiving this because you signed up at Ecommerce in Go.<br>
              Lagos, Nigeria
            </td>
          </tr>
        </table>

      </td>
    </tr>
  </table>
</body>
</html>`

var verifyTmpl = template.Must(template.New("verify").Parse(verifyEmailHtml))

func VerifyEmail(name, url string) string {
	var buf bytes.Buffer
	err := verifyTmpl.Execute(&buf, map[string]string{
		"Name": name, "Url": url, "ExpiresIn": "2 minutes",
	})

	if err != nil {
		slog.Error("render verify email failed", "error", err)
		return ""
	}

	return buf.String()
}
