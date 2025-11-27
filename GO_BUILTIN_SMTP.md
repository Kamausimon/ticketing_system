# Using Go's Built-in SMTP (No Third-Party Services Required)

The email system uses **Go's standard library** (`net/smtp`) - no external dependencies or third-party services required!

## 🎯 Key Benefits

✅ **Pure Go** - Uses only standard library packages  
✅ **No Vendor Lock-in** - Works with ANY SMTP server  
✅ **Simple** - Just configure host, port, and credentials  
✅ **Flexible** - Easy to switch between providers  
✅ **Secure** - Built-in TLS/SSL support  

## 🚀 Quick Start

### 1. Choose Your SMTP Server

You can use ANY SMTP server:
- **Gmail** (free, easy setup)
- **Outlook/Hotmail** (free)
- **Your own server** (full control)
- **Mailtrap** (testing only)
- **SendGrid, SES, etc.** (professional)

### 2. Configure Environment Variables

```bash
# Example: Using Gmail
EMAIL_HOST=smtp.gmail.com
EMAIL_PORT=587
EMAIL_USERNAME=your-email@gmail.com
EMAIL_PASSWORD=your-app-password
EMAIL_FROM=your-email@gmail.com
EMAIL_FROM_NAME=Your App Name
EMAIL_USE_TLS=true
```

### 3. Send Emails

```go
import (
    "ticketing_system/internal/config"
    "ticketing_system/internal/notifications"
)

// Load config
cfg := config.LoadOrPanic()

// Create service
notifService := notifications.NewNotificationService(cfg)

// Send email (pure Go, no third-party!)
err := notifService.SendWelcomeEmail("user@example.com", "John")
```

## 📖 How It Works

### The email system uses Go's standard library:

1. **`net/smtp`** - SMTP protocol implementation
   ```go
   smtp.SendMail(addr, auth, from, to, msg)
   ```

2. **`crypto/tls`** - Secure connections (TLS/SSL)
   ```go
   tls.Dial("tcp", addr, &tls.Config{...})
   ```

3. **`html/template`** - Email template rendering
   ```go
   template.Execute(buf, data)
   ```

### No external packages required!

The `go.mod` doesn't need any email-specific dependencies. It's all built-in.

## 🔧 Configuration Examples

### Gmail (Easiest for Development)

```env
EMAIL_HOST=smtp.gmail.com
EMAIL_PORT=587
EMAIL_USERNAME=yourname@gmail.com
EMAIL_PASSWORD=app_specific_password
EMAIL_USE_TLS=true
```

**Note:** Use [App Password](https://myaccount.google.com/apppasswords), not your regular password.

### Outlook/Hotmail

```env
EMAIL_HOST=smtp-mail.outlook.com
EMAIL_PORT=587
EMAIL_USERNAME=yourname@outlook.com
EMAIL_PASSWORD=your_password
EMAIL_USE_TLS=true
```

### Local SMTP Server (No Auth)

```env
EMAIL_HOST=localhost
EMAIL_PORT=25
EMAIL_USERNAME=
EMAIL_PASSWORD=
EMAIL_USE_TLS=false
```

Perfect for local development with tools like MailHog or local Postfix.

### Custom SMTP Server

```env
EMAIL_HOST=mail.yourdomain.com
EMAIL_PORT=587
EMAIL_USERNAME=smtp_user
EMAIL_PASSWORD=smtp_pass
EMAIL_USE_TLS=true
```

Works with ANY SMTP server!

## 🔒 Security

The system supports:

- **STARTTLS** (Port 587) - Most common
- **SSL/TLS** (Port 465) - Full encryption
- **Plain** (Port 25) - Local only

All using Go's built-in `crypto/tls` package.

## 🎨 Features

Even without third-party libraries, you get:

- ✅ HTML email templates
- ✅ Plain text fallback
- ✅ Automatic retry on failure
- ✅ TLS/SSL encryption
- ✅ Multiple recipients
- ✅ Custom headers
- ✅ Template rendering

## 📝 Code Example

Here's the core sending logic (pure Go):

```go
// From internal/notifications/email.go

func (s *EmailService) sendWithTLS(addr string, to []string, msg []byte) error {
    // Connect to SMTP server
    client, err := smtp.Dial(addr)
    if err != nil {
        return err
    }
    defer client.Close()

    // Start TLS
    tlsConfig := &tls.Config{
        ServerName: s.config.Host,
    }
    if err = client.StartTLS(tlsConfig); err != nil {
        return err
    }

    // Authenticate
    auth := smtp.PlainAuth("", username, password, host)
    if err = client.Auth(auth); err != nil {
        return err
    }

    // Send email
    if err = client.Mail(from); err != nil {
        return err
    }
    for _, addr := range to {
        if err = client.Rcpt(addr); err != nil {
            return err
        }
    }
    
    w, err := client.Data()
    if err != nil {
        return err
    }
    
    _, err = w.Write(msg)
    w.Close()
    
    return client.Quit()
}
```

Simple, clean, no magic - just Go!

## 🔄 Switching Providers

To switch from one provider to another, just update environment variables. No code changes needed:

```bash
# From Gmail...
EMAIL_HOST=smtp.gmail.com

# ...to Outlook
EMAIL_HOST=smtp-mail.outlook.com

# ...to custom server
EMAIL_HOST=mail.mycompany.com
```

The code stays the same!

## 🧪 Testing

### Option 1: Test Mode
```env
EMAIL_TEST_MODE=true
```
Logs emails without sending (perfect for unit tests).

### Option 2: Mailtrap
```env
EMAIL_HOST=smtp.mailtrap.io
EMAIL_PORT=587
```
Catches all emails in a test inbox.

### Option 3: Local SMTP
```env
EMAIL_HOST=localhost
EMAIL_PORT=1025
```
Use MailHog, MailCatcher, or similar tools.

## 📚 Standard Library Packages Used

- `net/smtp` - SMTP client
- `crypto/tls` - TLS/SSL
- `html/template` - Templates
- `bytes` - Buffer handling
- `fmt` - Formatting
- `time` - Retry delays

All part of Go's standard library - no `go get` required!

## 🎯 When to Add Third-Party Libraries

You might want third-party packages for:

- **Advanced templating** (e.g., Pongo2)
- **Email validation** (e.g., govalidator)
- **Email tracking** (open/click tracking)
- **Advanced attachments** (e.g., gomail)
- **HTML to text conversion**

But for basic email sending, **Go's standard library is perfect!**

## 💡 Pro Tips

1. **Use App Passwords** for Gmail (not your regular password)
2. **Test locally** with MailHog or Mailtrap first
3. **Enable TLS** for all production servers
4. **Send async** using goroutines (`go notifService.Send(...)`)
5. **Log errors** but don't fail main operations
6. **Keep credentials** in environment variables, never in code

## 🔗 Resources

- [Go net/smtp docs](https://pkg.go.dev/net/smtp)
- [Go crypto/tls docs](https://pkg.go.dev/crypto/tls)
- [Gmail App Passwords](https://myaccount.google.com/apppasswords)
- [MailHog (local testing)](https://github.com/mailhog/MailHog)

## 🎉 Conclusion

You don't need third-party services or libraries to send emails in Go!

The standard library provides everything you need:
- SMTP protocol support
- TLS/SSL encryption
- Template rendering
- Clean, simple API

Just configure your SMTP server and go! 🚀

---

**See `examples/smtp_builtin_example.go` for complete working examples.**
