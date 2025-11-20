# Event Management System API Testing Guide

## Overview
The event management system provides comprehensive functionality for organizers to create, manage, and publish events. This guide outlines all the available endpoints and how to test them.

## API Endpoints

### Public Event Endpoints
These endpoints don't require authentication and are accessible to all users:

1. **GET /events**
   - Get list of published events
   - Query parameters:
     - `page=1` - Page number (default: 1)
     - `limit=20` - Items per page (default: 20, max: 100)
     - `category=music` - Filter by event category
     - `location=nairobi` - Filter by location (partial match)
     - `start_date=2024-01-01` - Filter events starting from this date
     - `end_date=2024-12-31` - Filter events ending before this date
     - `search=concert` - Search in title, description, and location
     - `sort_by=date|popularity|created` - Sort criteria
     - `sort_order=asc|desc` - Sort order

2. **GET /events/{id}**
   - Get detailed information about a specific published event

3. **GET /events/{id}/images**
   - Get all images for a specific event

### Organizer Event Endpoints
These endpoints require authentication with organizer role:

1. **GET /organizers/events**
   - Get all events created by the authenticated organizer
   - Supports same query parameters as public events endpoint
   - Shows events in all statuses (draft, live, cancelled)

2. **POST /organizers/events**
   - Create a new event
   - Request body example:
   ```json
   {
     "title": "Music Concert",
     "location": "Nairobi",
     "description": "An amazing music concert",
     "start_date": "2024-06-01T19:00:00Z",
     "end_date": "2024-06-01T23:00:00Z",
     "category": "music",
     "currency": "KSH",
     "max_capacity": 500,
     "is_private": false,
     "min_age": 18,
     "location_address": "123 Main St",
     "location_country": "Kenya",
     "bg_type": "color",
     "bg_color": "#1a1a1a",
     "ticket_border_color": "#ffffff",
     "ticket_bg_color": "#000000",
     "ticket_text_color": "#ffffff",
     "ticket_sub_text_color": "#cccccc",
     "barcode_type": "qr",
     "is_barcode_enabled": true,
     "enable_offline_payment": false,
     "organizer_fee_fixed": 100.0,
     "organizer_fee_percentage": 5.0,
     "venue_ids": [1, 2]
   }
   ```

3. **PUT /organizers/events/{id}**
   - Update an existing event
   - Same request body format as create

4. **DELETE /organizers/events/{id}**
   - Delete an event
   - Draft events are permanently deleted
   - Published events are marked as cancelled

5. **POST /organizers/events/{id}/publish**
   - Publish a draft event (make it live and visible to public)

6. **POST /organizers/events/{id}/images**
   - Upload an image for an event
   - Form data with "image" field containing the image file
   - Supported formats: PNG, JPEG, JPG, GIF
   - Maximum size: 5MB

7. **DELETE /organizers/events/{id}/images/{imageId}**
   - Delete a specific image from an event

## Event Statuses

- **draft**: Event is being created but not yet published
- **live**: Event is published and visible to the public
- **cancelled**: Event has been cancelled
- **completed**: Event has finished (future enhancement)

## Event Categories

Available categories:
- `music`
- `conference`
- `seminar`
- `TradeShow`
- `Product Launch`
- `Team Building`
- `Corporate Meeting`
- `Corporate Retreat`
- `sports`
- `educational`
- `festival`
- `art`

## Testing Flow

1. **Register and Login as Organizer**:
   - Register a user
   - Apply as organizer
   - Complete organizer profile
   - Get admin verification (if required)

2. **Create Events**:
   - Create draft events with different categories
   - Upload images for events
   - Publish events when ready

3. **Manage Events**:
   - List organizer's events
   - Update event details
   - Delete unwanted events

4. **Public Access**:
   - Test public event listing
   - Test event details view
   - Test filtering and searching

## Example cURL Commands

### Create Event:
```bash
curl -X POST http://localhost:8080/organizers/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "title": "Tech Conference 2024",
    "location": "Nairobi",
    "description": "Annual technology conference",
    "start_date": "2024-06-15T09:00:00Z",
    "end_date": "2024-06-15T17:00:00Z",
    "category": "conference",
    "currency": "KSH",
    "max_capacity": 200
  }'
```

### List Public Events:
```bash
curl "http://localhost:8080/events?category=music&limit=10&sort_by=date&sort_order=asc"
```

### Upload Event Image:
```bash
curl -X POST http://localhost:8080/organizers/events/1/images \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "image=@event_poster.jpg"
```

## Key Features

1. **Comprehensive Event Management**: Full CRUD operations for events
2. **Role-Based Access**: Public viewing vs organizer management
3. **Image Management**: Upload and manage event images
4. **Advanced Filtering**: Search, filter, and sort events
5. **Status Management**: Draft → Live → Cancelled workflow
6. **Validation**: Comprehensive validation for all inputs
7. **File Security**: Image type and size validation
8. **Pagination**: Efficient pagination for large event lists
9. **Venue Integration**: Support for multiple venues per event
10. **Responsive Design**: JSON APIs suitable for web and mobile clients

The event management system is now fully functional and ready for testing and integration with the frontend!