# Outalator Web UI

A modern web interface for managing outage tracking and troubleshooting notes.

## Features

### 1. Outage List View
- View all outages with status and severity indicators
- Filter by status (open, investigating, resolved, closed)
- Filter by severity (critical, high, medium, low)
- Search by title or description
- Search by tags (key/value pairs)
- Real-time statistics (notes count, tags count, alerts count)

### 2. Outage Detail View
- Full outage information display
- Update outage status
- View and add troubleshooting notes
- View and add tags
- View linked alerts from PagerDuty/OpsGenie
- Merge multiple outages together

### 3. Create New Outage
- Create outages with title, description, status, and severity
- Clean form-based interface

### 4. Notes Management
- Add notes in plain text or markdown format
- Author attribution
- Timestamps with relative time display
- Chronologically sorted

### 5. Tag Management
- Add key-value tags (e.g., jira:OPS-123, service:api-gateway)
- Remove tags (Note: requires backend API enhancement)
- Search outages by tags

### 6. Merge Functionality
- Select multiple outages to merge into a primary outage
- Automatically imports all notes and tags from merged outages
- Notes are prefixed with source outage information

## Usage

### Starting the Application

1. Build the application:
   ```bash
   go build -o outalator cmd/outalator/main.go
   ```

2. Ensure your PostgreSQL database is running and configured in `config.yaml`

3. Run the application:
   ```bash
   ./outalator
   ```

4. Open your browser to `http://localhost:8080`

### Navigation

- **Outages Tab**: View and filter all outages
- **Create New Tab**: Create a new outage
- Click any outage card to view details
- Click "Back to List" to return to the outage list

### Working with Outages

#### Filtering
- Use the status and severity dropdowns to filter outages
- Enter search terms to find outages by title or description
- Use tag search to find outages with specific tags
- Click "Clear Filters" to reset all filters

#### Adding Notes
1. Navigate to an outage detail page
2. Scroll to the Notes section
3. Enter your note content
4. Optionally provide your name (defaults to "Anonymous")
5. Choose format (Plain Text or Markdown)
6. Click "Add Note"

#### Adding Tags
1. Navigate to an outage detail page
2. In the Tags section, click "+ Add Tag"
3. Enter a key (e.g., "jira") and value (e.g., "OPS-123")
4. Click "Add"

#### Merging Outages
1. Navigate to the primary outage (the one you want to keep)
2. Click "Merge Outages"
3. Select one or more outages to merge
4. Click "Merge Selected"
5. All notes and tags from selected outages will be imported

#### Updating Status
1. Navigate to an outage detail page
2. Use the "Update Status" dropdown
3. Select the new status
4. Click "Update"

## Architecture

The web UI is a single-page application (SPA) built with vanilla JavaScript:

- **HTML**: `/web/static/index.html` - Main page structure
- **CSS**: `/web/static/css/styles.css` - Styling and responsive design
- **JavaScript**: `/web/static/js/app.js` - Application logic and API integration

### API Integration

The UI communicates with the backend REST API at `/api/v1/`:

- `GET /api/v1/outages` - List outages
- `GET /api/v1/outages/{id}` - Get outage details
- `POST /api/v1/outages` - Create outage
- `PATCH /api/v1/outages/{id}` - Update outage
- `POST /api/v1/outages/{id}/notes` - Add note
- `POST /api/v1/outages/{id}/tags` - Add tag
- `GET /api/v1/tags/search` - Search by tags

## Browser Compatibility

The web UI is compatible with all modern browsers:
- Chrome/Edge 90+
- Firefox 88+
- Safari 14+

## Responsive Design

The UI is fully responsive and works on:
- Desktop (1400px+)
- Tablet (768px - 1400px)
- Mobile (< 768px)

## Color Coding

### Status
- **Open**: Blue
- **Investigating**: Yellow/Orange
- **Resolved**: Green
- **Closed**: Gray

### Severity
- **Critical**: Red
- **High**: Orange
- **Medium**: Blue
- **Low**: Green

## Future Enhancements

Potential improvements:
1. Delete tag API endpoint (currently shows error message)
2. Delete outage functionality
3. Advanced search with multiple filters
4. Export outages to JSON/CSV
5. Real-time updates using WebSockets
6. User authentication UI integration
7. Pagination for large outage lists
8. Rich markdown editor
9. Attachment support for notes
10. Timeline view of outage lifecycle
