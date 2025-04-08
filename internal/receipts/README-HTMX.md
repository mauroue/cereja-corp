# Receipt Scanner Web Interface with HTMX

This document explains the web interface implementation for the receipt scanner app using HTMX.

## What is HTMX?

HTMX is a lightweight JavaScript library that allows you to access modern browser features directly from HTML using attributes. It enables building dynamic web apps without writing custom JavaScript for common tasks like form submissions, loading content dynamically, etc.

## Web Interface Structure

The web interface for the receipt scanner app consists of:

1. **Templates**: HTML templates for each page/view
2. **CSS**: Styling for the web interface
3. **HTMX**: For dynamic content loading and form submission
4. **Web Handlers**: Go functions that serve these templates and handle HTMX requests

## Routes

The web interface is available at the following routes:

- `/receipts-web/` - Home page
- `/receipts-web/upload` - Upload page
- `/receipts-web/list` - List of receipts
- `/receipts-web/view/:id` - View a specific receipt

## HTMX Endpoints

The following HTMX endpoints are available:

- `POST /receipts-web/htmx/upload` - Upload a receipt image
- `GET /receipts-web/htmx/receipts` - Get a list of receipts
- `GET /receipts-web/htmx/receipt/:id` - Get details of a specific receipt
- `GET /receipts-web/htmx/receipt/:id/items` - Get items for a specific receipt

## How It Works

1. **Page Load**: Initial HTML is served from the server when visiting a route.
2. **Dynamic Content**: HTMX attributes (`hx-get`, `hx-post`, etc.) define how/when content is fetched.
3. **Content Swapping**: HTMX swaps in the new content without a full page reload.

Example of HTMX in action:

```html
<!-- This div will load content from /receipts-web/htmx/receipts on page load -->
<div id="receipts-list"
     hx-get="/receipts-web/htmx/receipts"
     hx-trigger="load"
     hx-indicator="#receipts-loading">
    <div id="receipts-loading" class="text-center htmx-indicator">
        <div class="loading-spinner"></div>
        <p>Loading receipts...</p>
    </div>
</div>
```

## Image Upload with HTMX

The image upload process uses HTMX to handle the form submission:

1. User selects or drags an image file
2. JavaScript shows a preview of the image
3. User clicks "Process Receipt"
4. HTMX submits the form with multipart encoding
5. Server processes the image and saves the data
6. HTMX redirects to the receipts list on success

## Features

- **Drag & Drop**: Support for dragging and dropping receipt images
- **Image Preview**: See the image before uploading
- **Progress Indication**: Loading spinners for HTMX requests
- **Pagination**: "Load More" button for receipts list
- **Search**: Search functionality for finding receipts

## Running the App

The web interface is served automatically when running the main application. Visit http://localhost:8080/receipts-web/ to access it. 