# Demo App - Event Ticketing System

A simple, interactive demo frontend for the Event Ticketing System backend. This demo showcases the key features of the API including authentication, event browsing, ticket purchasing, and organizer management.

## 🎯 Features

### For Attendees
- ✅ **Browse Events** - View all upcoming events with details
- ✅ **Search Events** - Find events by name or description
- ✅ **Purchase Tickets** - Buy tickets for events
- ✅ **View My Tickets** - See all purchased tickets
- ✅ **Download Tickets** - Get PDF tickets

### For Organizers
- ✅ **Dashboard** - View event statistics and revenue
- ✅ **Create Events** - Add new events with details
- ✅ **Manage Events** - View and edit your events
- ✅ **Track Sales** - Monitor ticket sales

### Security
- ✅ **User Authentication** - Login/Register with JWT
- ✅ **Protected Routes** - Authenticated endpoints
- ✅ **Token Management** - Persistent sessions

## 🚀 Quick Start

### Prerequisites
- Your backend API running on `http://localhost:8080`
- A web browser
- (Optional) A local web server

### Option 1: Direct File Access
Simply open `index.html` in your browser:
```bash
cd demo-app
open index.html  # macOS
# or
xdg-open index.html  # Linux
# or double-click the file
```

### Option 2: Using Python (Recommended)
```bash
cd demo-app
python3 -m http.server 3000
```
Then visit: `http://localhost:3000`

### Option 3: Using Node.js
```bash
cd demo-app
npx serve .
```

### Option 4: Using PHP
```bash
cd demo-app
php -S localhost:3000
```

## 🔧 Configuration

The API URL is configured in [app.js](app.js#L2):
```javascript
const API_URL = 'http://localhost:8080';
```

If your backend runs on a different port, update this value.

## 📖 How to Use

### 1. Register an Account
- Click **Sign Up** in the navigation
- Fill in your details (name, email, password)
- Click **Sign Up**

### 2. Login
- Click **Login** in the navigation
- Enter your email and password
- Click **Login**

### 3. Browse Events
- View all events on the home page
- Click on an event card to see details
- Use the search bar to find specific events

### 4. Purchase Tickets
- Click on an event to view details
- Select quantity using +/- buttons
- Click **Purchase Tickets**
- Follow payment flow (simplified in demo)

### 5. View Your Tickets
- Click **My Tickets** in navigation
- See all your purchased tickets
- Download tickets as PDF

### 6. Become an Organizer (Backend Required)
First, apply to become an organizer via API:
```bash
curl -X POST http://localhost:8080/organizers/apply \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "organization_name": "My Events Co",
    "business_type": "company",
    "contact_email": "contact@myevents.com",
    "contact_phone": "+254712345678"
  }'
```

Then verify the organizer (admin action):
```bash
# Get organizer ID from /admin/organizers/pending
curl -X POST http://localhost:8080/admin/organizers/{id}/verify \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

### 7. Create Events (Organizers Only)
- Click **Organizer** in navigation
- Click **Create Event** button
- Fill in event details
- Click **Create Event**

## 🎨 Features Showcase

### Authentication Flow
```
Register → Verify Email → Login → Get JWT Token → Access Protected Routes
```

### Ticket Purchase Flow
```
Browse Events → Select Event → Choose Quantity → Purchase → View Tickets → Download PDF
```

### Organizer Flow
```
Apply → Get Verified → Create Events → Monitor Sales → View Dashboard
```

## 📁 Project Structure

```
demo-app/
├── index.html      # Main HTML structure
├── styles.css      # All styling and responsive design
├── app.js          # JavaScript logic and API integration
└── README.md       # This file
```

## 🔌 API Endpoints Used

### Authentication
- `POST /register` - Register new user
- `POST /login` - User login
- `POST /logout` - User logout

### Events
- `GET /events` - List all events
- `GET /events/{id}` - Get event details
- `POST /organizers/events` - Create event (organizer)
- `GET /organizers/events` - List organizer's events

### Tickets
- `GET /tickets` - List user's tickets
- `GET /tickets/{id}/pdf` - Download ticket PDF
- `POST /orders` - Create ticket order
- `POST /orders/{id}/payment` - Process payment

### Organizer
- `GET /organizers/profile` - Get organizer profile
- `GET /organizers/dashboard/stats` - Get dashboard statistics

## 🎯 Test Credentials

Create test accounts using the signup form or use these if you have seed data:

**Regular User:**
```
Email: user@example.com
Password: password123
```

**Organizer:**
```
Email: organizer@example.com
Password: password123
```

## 🐛 Troubleshooting

### CORS Issues
If you see CORS errors, ensure your backend has CORS enabled:
```go
// In your Go backend
router.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"http://localhost:3000", "file://"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    AllowCredentials: true,
}))
```

### API Connection Failed
- Verify backend is running: `curl http://localhost:8080/health`
- Check API_URL in `app.js` matches your backend port
- Check browser console for error details

### Events Not Loading
- Ensure you have events in your database
- Create test events via the organizer dashboard
- Check network tab in browser dev tools

### Cannot Purchase Tickets
- Ensure you're logged in
- Check that email is verified (if required)
- Verify event has available tickets
- Check console for payment errors

## 🚀 Production Deployment

### Static Hosting (Netlify, Vercel, GitHub Pages)
1. Update `API_URL` in `app.js` to your production backend URL
2. Deploy the `demo-app` folder to your hosting service

### With Docker
```dockerfile
FROM nginx:alpine
COPY demo-app/ /usr/share/nginx/html/
EXPOSE 80
```

## 🔐 Security Notes

⚠️ **This is a demo application**. For production:

1. **Environment Variables** - Use `.env` files for API URLs
2. **HTTPS** - Always use HTTPS in production
3. **Token Storage** - Consider more secure token storage
4. **Input Validation** - Add comprehensive client-side validation
5. **Error Handling** - Implement proper error boundaries
6. **Rate Limiting** - Handle rate limit responses gracefully
7. **Payment Integration** - Use proper payment gateway (IntaSend, Stripe)

## 📝 Customization

### Change Colors
Edit CSS variables in [styles.css](styles.css#L7-L17):
```css
:root {
    --primary-color: #6366f1;
    --primary-dark: #4f46e5;
    /* ... */
}
```

### Add Features
The code is modular and easy to extend:
- Add new pages in `index.html`
- Add navigation links
- Add API functions in `app.js`
- Add corresponding UI handlers

### Modify API URL
Update in [app.js](app.js#L2):
```javascript
const API_URL = 'https://your-api.com';
```

## 🤝 Contributing

Feel free to enhance this demo:
- Add more features (promotions, refunds, etc.)
- Improve UI/UX
- Add animations
- Make it responsive for mobile
- Add more error handling

## 📄 License

This demo app is part of the Event Ticketing System project.

## 🔗 Related Documentation

- [API Routes](../API_ROUTES.md) - Complete API documentation
- [Backend README](../README.md) - Backend setup instructions
- [Email Verification](../EMAIL_VERIFICATION_API.md) - Email verification guide
- [Payment System](../PAYMENT_SYSTEM_INTASEND.md) - Payment integration

## 💡 Tips

1. **Test Different Roles**: Create both regular user and organizer accounts to test all features
2. **Check Console**: Open browser dev tools to see API requests and responses
3. **Network Tab**: Monitor API calls in the network tab for debugging
4. **Local Storage**: Check Application > Local Storage in dev tools to see stored tokens
5. **Responsive**: Test on different screen sizes for mobile compatibility

## 📞 Support

If you encounter issues:
1. Check the browser console for errors
2. Verify backend is running and accessible
3. Check API endpoint documentation
4. Review backend logs for API errors

---

**Built with ❤️ for demonstrating the Event Ticketing System API**
